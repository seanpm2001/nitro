package hostedit

import (
	"testing"
)

func TestUpdate(t *testing.T) {
	type args struct {
		file  string
		addr  string
		hosts []string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "can append to the file when no section is found",
			args: args{file: "testdata/no-section.txt", addr: "127.0.0.1", hosts: []string{"one", "two", "three"}},
			want: `##
# Host Database
#
# localhost is used to configure the loopback interface
# when the system is booting.  Do not change this entry.
##
127.0.0.1        localhost
255.255.255.255  broadcasthost
::1              localhost

127.0.0.1        kubernetes.docker.internal
# Added by Docker Desktop
# To allow the same kube context to work on the host and the container:
127.0.0.1        kubernetes.docker.internal
# End of section

# <nitro>
127.0.0.1	one two three
# </nitro>
`,
		},
		{
			name: "can update the right section",
			args: args{file: "testdata/has-section.txt", addr: "127.0.0.1", hosts: []string{"one", "two", "three"}},
			want: `##
# Host Database
#
# localhost is used to configure the loopback interface
# when the system is booting.  Do not change this entry.
##
127.0.0.1        localhost
255.255.255.255  broadcasthost
::1              localhost

# <nitro>
127.0.0.1	one two three
# </nitro>

127.0.0.1        kubernetes.docker.internal
# Added by Docker Desktop
# To allow the same kube context to work on the host and the container:
127.0.0.1        kubernetes.docker.internal
# End of section
`,
		},
		{
			name:    "no file returns an error",
			args:    args{file: "testdata/empty"},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Update(tt.args.file, tt.args.addr, tt.args.hosts...)

			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Update() = %v, want %v", got, tt.want)
			}
		})
	}
}