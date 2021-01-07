package nitro

import (
	"log"
	"os"

	nitroclient "github.com/craftcms/nitro/client"
	"github.com/craftcms/nitro/command/add"
	"github.com/craftcms/nitro/command/apply"
	"github.com/craftcms/nitro/command/clean"
	"github.com/craftcms/nitro/command/completion"
	"github.com/craftcms/nitro/command/composer"
	"github.com/craftcms/nitro/command/context"
	"github.com/craftcms/nitro/command/craft"
	"github.com/craftcms/nitro/command/create"
	"github.com/craftcms/nitro/command/database"
	"github.com/craftcms/nitro/command/destroy"
	"github.com/craftcms/nitro/command/disable"
	"github.com/craftcms/nitro/command/edit"
	"github.com/craftcms/nitro/command/enable"
	"github.com/craftcms/nitro/command/hosts"
	"github.com/craftcms/nitro/command/initialize"
	"github.com/craftcms/nitro/command/logs"
	"github.com/craftcms/nitro/command/npm"
	"github.com/craftcms/nitro/command/queue"
	"github.com/craftcms/nitro/command/restart"
	"github.com/craftcms/nitro/command/ssh"
	"github.com/craftcms/nitro/command/start"
	"github.com/craftcms/nitro/command/stop"
	"github.com/craftcms/nitro/command/trust"
	"github.com/craftcms/nitro/command/update"
	"github.com/craftcms/nitro/command/validate"
	"github.com/craftcms/nitro/command/version"
	"github.com/craftcms/nitro/command/xoff"
	"github.com/craftcms/nitro/command/xon"
	"github.com/craftcms/nitro/pkg/downloader"
	"github.com/craftcms/nitro/pkg/terminal"
	"github.com/docker/docker/client"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

var rootCommand = &cobra.Command{
	Use:   "nitro",
	Short: "Local Craft CMS dev made easy",
	Long: `Nitro is a command-line tool focused on making local Craft CMS development quick and easy.

Version: ` + version.Version,
	RunE:         rootMain,
	SilenceUsage: true,
	Version:      version.Version,
}

func rootMain(command *cobra.Command, _ []string) error {
	return command.Help()
}

func NewCommand() *cobra.Command {
	// get the users home directory
	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}

	// create the docker client
	docker, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatal(err)
	}

	// get the port for the nitrod API
	apiPort := "5000"
	if os.Getenv("NITRO_API_PORT") != "" {
		apiPort = os.Getenv("NITRO_API_PORT")
	}

	// create the nitrod gRPC API
	nitrod, err := nitroclient.NewClient("127.0.0.1", apiPort)
	if err != nil {
		log.Fatal(err)
	}

	// create the "terminal" for capturing output
	term := terminal.New()

	// create the downloaded for creating projects
	downloader := downloader.NewDownloader()

	// register all of the commands
	commands := []*cobra.Command{
		add.NewCommand(home, docker, term),
		apply.NewCommand(home, docker, nitrod, term),
		clean.NewCommand(home, docker, term),
		completion.New(),
		composer.NewCommand(docker, term),
		context.NewCommand(home, docker, term),
		craft.NewCommand(home, docker, term),
		create.New(docker, downloader, term),
		database.NewCommand(home, docker, term),
		destroy.NewCommand(home, docker, term),
		disable.NewCommand(home, docker, term),
		enable.NewCommand(home, docker, term),
		edit.NewCommand(home, docker, term),
		hosts.New(home, term),
		initialize.NewCommand(home, docker, term),
		logs.NewCommand(home, docker, term),
		npm.NewCommand(docker, term),
		queue.NewCommand(home, docker, term),
		restart.New(docker, term),
		ssh.NewCommand(home, docker, term),
		start.NewCommand(docker, term),
		stop.New(docker, term),
		trust.New(docker, term),
		update.NewCommand(docker, term),
		validate.NewCommand(home, docker, term),
		version.New(docker, nitrod, term),
		xon.NewCommand(home, docker, term),
		xoff.NewCommand(home, docker, term),
	}

	// add the commands
	rootCommand.AddCommand(commands...)

	return rootCommand
}