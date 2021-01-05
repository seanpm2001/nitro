package apply

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"github.com/craftcms/nitro/command/apply/internal/databasecontainer"
	"github.com/craftcms/nitro/command/apply/internal/mountcontainer"
	"github.com/craftcms/nitro/command/apply/internal/sitecontainer"
	"github.com/craftcms/nitro/pkg/backup"
	"github.com/craftcms/nitro/pkg/config"
	"github.com/craftcms/nitro/pkg/datetime"
	"github.com/craftcms/nitro/pkg/hostedit"
	"github.com/craftcms/nitro/pkg/labels"
	"github.com/craftcms/nitro/pkg/proxycontainer"
	"github.com/craftcms/nitro/pkg/sudo"
	"github.com/craftcms/nitro/pkg/terminal"
	"github.com/craftcms/nitro/protob"
)

var (
	// ErrNoProxyContainer is returned when the proxy container is not found for an environment
	ErrNoProxyContainer = fmt.Errorf("unable to locate the proxy container")

	knownContainers = map[string]bool{}
)

const exampleText = `  # apply changes from a config
  nitro apply

  # skip editing the hosts file
  nitro apply --skip-hosts

  # you can also set the environment variable "NITRO_EDIT_HOSTS" to "false"`

// NewCommand returns the command used to apply configuration file changes to a nitro environment.
func NewCommand(home string, docker client.CommonAPIClient, nitrod protob.NitroClient, output terminal.Outputer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "apply",
		Short:   "Apply changes",
		Example: exampleText,
		PostRunE: func(cmd *cobra.Command, args []string) error {
			// create a filter for the environment
			filter := filters.NewArgs()
			filter.Add("label", labels.Nitro+"=true")

			// look for a container for the site
			containers, err := docker.ContainerList(cmd.Context(), types.ContainerListOptions{All: true, Filters: filter})
			if err != nil {
				return fmt.Errorf("error getting a list of containers")
			}

			// if there are no matching containers we are done
			if len(containers) == 0 {
				return nil
			}

			for _, c := range containers {
				if _, ok := knownContainers[c.ID]; !ok {
					// don't remove the proxy container
					if c.Labels[labels.Proxy] != "" {
						continue
					}

					// set the container name
					name := strings.TrimLeft(c.Names[0], "/")

					output.Info("removing", name)

					// only perform a backup if the container is for databases
					if c.Labels[labels.DatabaseEngine] != "" {
						// get all of the databases
						databases, err := backup.Databases(cmd.Context(), docker, c.ID, c.Labels[labels.DatabaseCompatability])
						if err != nil {
							output.Info("Unable to get the databases from", name, err.Error())

							break
						}

						// backup each database
						for _, db := range databases {
							// create the database specific backup options
							opts := &backup.Options{
								BackupName:    fmt.Sprintf("%s-%s.sql", db, datetime.Parse(time.Now())),
								ContainerID:   c.ID,
								ContainerName: name,
								Database:      db,
								Home:          home,
							}

							// create the backup command based on the compatability type
							switch c.Labels[labels.DatabaseCompatability] {
							case "postgres":
								opts.Commands = []string{"pg_dump", "--username=nitro", db, "-f", "/tmp/" + opts.BackupName}
							default:
								opts.Commands = []string{"/usr/bin/mysqldump", "-h", "127.0.0.1", "-unitro", "--password=nitro", db, "--result-file=" + "/tmp/" + opts.BackupName}
							}

							output.Pending("creating backup", opts.BackupName)

							// backup the container
							if err := backup.Perform(cmd.Context(), docker, opts); err != nil {
								output.Warning()
								output.Info("Unable to backup database", db, err.Error())

								break
							}

							output.Done()
						}

						// show where all backups are saved for this container
						output.Info("Backups saved in", filepath.Join(home, ".nitro", name), "💾")
					}

					// stop and remove a container we don't know about
					if err := docker.ContainerStop(cmd.Context(), c.ID, nil); err != nil {
						return err
					}

					// remove container
					if err := docker.ContainerRemove(cmd.Context(), c.ID, types.ContainerRemoveOptions{}); err != nil {
						return err
					}
				}
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				// when we call commands from other commands (e.g. init)
				// the context could be nil, so we set it to the parent
				// context just in case.
				ctx = cmd.Parent().Context()
			}

			// load the config
			cfg, err := config.Load(home)
			if err != nil {
				return err
			}

			// create a filter for the environment
			filter := filters.NewArgs()
			filter.Add("label", labels.Nitro+"=true")

			// add the filter for the network name
			filter.Add("name", "nitro-network")

			output.Info("Checking network...")

			// check the network
			var network types.NetworkResource
			networks, err := docker.NetworkList(ctx, types.NetworkListOptions{Filters: filter})
			if err != nil {
				return fmt.Errorf("unable to list docker networks\n%w", err)
			}

			// get the network for the environment
			for _, n := range networks {
				if n.Name == "nitro-network" {
					network = n
					break
				}
			}

			// if the network is not found
			if network.ID == "" {
				output.Info("No network was found...\nrun `nitro init` to get started")
				return nil
			}

			// remove the filter
			filter.Del("name", "nitro-network")

			output.Success("network ready")

			output.Info("Checking proxy...")

			// check the proxy and ensure its started
			_, err = proxycontainer.FindAndStart(ctx, docker)
			if errors.Is(err, ErrNoProxyContainer) {
				output.Info("unable to find the nitro proxy...\n run `nitro init` to resolve")
				return nil
			}
			if err != nil {
				return err
			}

			output.Success("proxy ready")

			output.Info("Checking databases...")

			// check the databases
			for _, db := range cfg.Databases {
				n, _ := db.GetHostname()
				output.Pending("checking", n)

				// start or create the database
				id, err := databasecontainer.StartOrCreate(ctx, docker, network.ID, db)
				if err != nil {
					output.Warning()
					return err
				}

				knownContainers[id] = true

				output.Done()
			}

			// check the mounts
			if len(cfg.Mounts) > 0 {
				output.Info("Checking mounts...")

				fmt.Println(len(cfg.Mounts))

				for _, m := range cfg.Mounts {
					output.Pending("checking", m.Path)

					id, err := mountcontainer.FindOrCreate(ctx, docker, home, network.ID, m)
					if err != nil {
						output.Warning()
						return err
					}

					// set the container id as known
					knownContainers[id] = true

					output.Done()
				}
			}

			output.Info("Checking services...")

			// check dynamodb service
			if cfg.Services.DynamoDB {
				output.Pending("checking dynamodb service")

				id, err := dynamodb(ctx, docker, cfg.Services.DynamoDB, network.ID)
				if err != nil {
					return err
				}
				if id != "" {
					knownContainers[id] = true
				}

				output.Done()
			}

			// check dynamodb service
			if cfg.Services.Mailhog {
				output.Pending("checking mailhog service")

				id, err := mailhog(ctx, docker, cfg.Services.Mailhog, network.ID)
				if err != nil {
					return err
				}
				if id != "" {
					knownContainers[id] = true
				}

				output.Done()
			}

			if len(cfg.Sites) > 0 {
				// get all of the sites, their local path, the php version, and the type of project (nginx or PHP-FPM)
				output.Info("Checking sites...")

				// get the envs for the sites
				for _, site := range cfg.Sites {
					output.Pending("checking", site.Hostname)

					// start, update or create the site container
					id, err := sitecontainer.StartOrCreate(ctx, docker, home, network.ID, site)
					if err != nil {
						output.Warning()
						return err
					}

					knownContainers[id] = true

					output.Done()
				}
			}

			output.Info("Checking proxy...")

			output.Pending("updating proxy")

			if err := updateProxy(ctx, docker, nitrod, *cfg); err != nil {
				output.Warning()
				return err
			}

			output.Done()

			// should we update the hosts file?
			if os.Getenv("NITRO_EDIT_HOSTS") == "false" || cmd.Flag("skip-hosts").Value.String() == "true" {
				// skip updating the hosts file
				return nil
			}

			// get all possible hostnames
			var hostnames []string
			for _, s := range cfg.Sites {
				hostnames = append(hostnames, s.Hostname)
				hostnames = append(hostnames, s.Aliases...)
			}

			if len(hostnames) > 0 {
				// set the hosts file based on the OS
				defaultFile := "/etc/hosts"
				if runtime.GOOS == "windows" {
					defaultFile = `C:\Windows\System32\Drivers\etc\hosts`
				}

				// check if hosts is already up to date
				updated, err := hostedit.IsUpdated(defaultFile, "127.0.0.1", hostnames...)
				if err != nil {
					return err
				}

				// if the hosts file is not updated
				if !updated {
					// get the executable
					nitro, err := os.Executable()
					if err != nil {
						return fmt.Errorf("unable to locate the nitro path, %w", err)
					}

					// run the hosts command
					switch runtime.GOOS {
					case "windows":
						// windows users should be running as admin, so just execute the hosts command
						// as is
						c := exec.Command(nitro, "hosts", "--hostnames="+strings.Join(hostnames, ","))

						c.Stdout = os.Stdout
						c.Stderr = os.Stderr

						if c.Run() != nil {
							return err
						}
					default:
						output.Info("Modifying hosts file (you might be prompted for your password)")

						// add the hosts
						if err := sudo.Run(nitro, "nitro", "hosts", "--hostnames="+strings.Join(hostnames, ",")); err != nil {
							return err
						}
					}
				}
			}

			output.Info("Nitro is up and running 😃")

			return nil
		},
	}

	// add flag to skip pulling images
	cmd.Flags().Bool("skip-hosts", false, "skip modifying the hosts file")

	return cmd
}

func updateProxy(ctx context.Context, docker client.ContainerAPIClient, nitrod protob.NitroClient, cfg config.Config) error {
	// convert the sites into the gRPC API Apply request
	sites := make(map[string]*protob.Site)
	for _, s := range cfg.Sites {
		// create the site
		sites[s.Hostname] = &protob.Site{
			Hostname: s.Hostname,
			Aliases:  strings.Join(s.Aliases, ","),
			Port:     8080,
		}
	}

	// if there are no sites, we are done
	if len(sites) == 0 {
		return nil
	}

	// wait for the api to be ready
	for {
		_, err := nitrod.Ping(ctx, &protob.PingRequest{})
		if err == nil {
			break
		}
	}

	// configure the proxy with the sites
	resp, err := nitrod.Apply(ctx, &protob.ApplyRequest{Sites: sites})
	if err != nil {
		return err
	}

	if resp.Error {
		return fmt.Errorf("unable to update the proxy, %s", resp.GetMessage())
	}

	return nil
}
