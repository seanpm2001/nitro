package config

import (
	"reflect"
	"testing"

	"github.com/spf13/viper"
)

func TestGetInt(t *testing.T) {
	type args struct {
		key  string
		flag int
	}
	tests := []struct {
		name       string
		keyToSet   string
		valueToSet interface{}
		args       args
		want       int
	}{
		{
			name: "can get the flag when viper is not set",
			args: args{
				key:  "some.key",
				flag: 4,
			},
			want: 4,
		},
		{
			name:       "can get the flag when viper is set",
			keyToSet:   "some.key",
			valueToSet: 5,
			args: args{
				key:  "some.key",
				flag: 0,
			},
			want: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.keyToSet != "" {
				viper.Set(tt.keyToSet, tt.valueToSet)
			}

			if got := GetInt(tt.args.key, tt.args.flag); got != tt.want {
				t.Errorf("GetString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetString(t *testing.T) {
	type args struct {
		key  string
		flag string
	}
	tests := []struct {
		name       string
		keyToSet   string
		valueToSet interface{}
		args       args
		want       string
	}{
		{
			name: "can get the flag when viper is not set",
			args: args{
				key:  "some.key",
				flag: "value",
			},
			want: "value",
		},
		{
			name:       "can get the flag when viper is set",
			keyToSet:   "some.key",
			valueToSet: "thevalue",
			args: args{
				key:  "some.key",
				flag: "",
			},
			want: "thevalue",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.keyToSet != "" {
				viper.Set(tt.keyToSet, tt.valueToSet)
			}

			if got := GetString(tt.args.key, tt.args.flag); got != tt.want {
				t.Errorf("GetString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_RemoveSite(t *testing.T) {
	type fields struct {
		Name      string
		PHP       string
		CPUs      string
		Disk      string
		Memory    string
		Databases []Database
		Sites     []Site
	}
	type args struct {
		site string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []Site
		wantErr bool
	}{
		{
			name: "remove a site by its domain",
			args: args{
				site: "anotherexample.test",
			},
			fields: fields{
				Sites: []Site{
					{
						Domain:  "example.test",
						Path:    "/some/path",
						Docroot: "web",
					},
					{
						Domain:  "anotherexample.test",
						Path:    "/some/path/to/anotherexample",
						Docroot: "web",
					},
					{
						Domain:  "finalexample.test",
						Path:    "/some/path/to/finalexample",
						Docroot: "web",
					},
				},
			},
			want: []Site{
				{
					Domain:  "example.test",
					Path:    "/some/path",
					Docroot: "web",
				},
				{
					Domain:  "finalexample.test",
					Path:    "/some/path/to/finalexample",
					Docroot: "web",
				},
			},
			wantErr: false,
		},
		{
			name: "sites not in the slice return an error",
			args: args{
				site: "doesnotexist.test",
			},
			fields: fields{
				Sites: []Site{
					{
						Domain:  "example.test",
						Path:    "/some/path",
						Docroot: "web",
					},
					{
						Domain:  "anotherexample.test",
						Path:    "/some/path/to/anotherexample",
						Docroot: "web",
					},
					{
						Domain:  "finalexample.test",
						Path:    "/some/path/to/finalexample",
						Docroot: "web",
					},
				},
			},
			want: []Site{
				{
					Domain:  "example.test",
					Path:    "/some/path",
					Docroot: "web",
				},
				{
					Domain:  "anotherexample.test",
					Path:    "/some/path/to/anotherexample",
					Docroot: "web",
				},
				{
					Domain:  "finalexample.test",
					Path:    "/some/path/to/finalexample",
					Docroot: "web",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Name:      tt.fields.Name,
				PHP:       tt.fields.PHP,
				CPUs:      tt.fields.CPUs,
				Disk:      tt.fields.Disk,
				Memory:    tt.fields.Memory,
				Databases: tt.fields.Databases,
				Sites:     tt.fields.Sites,
			}

			err := c.RemoveSite(tt.args.site)
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveSite() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.want != nil {
				if !reflect.DeepEqual(c.Sites, tt.want) {
					t.Errorf("RemoveSite() got = \n%v, \nwant \n%v", c.Sites, tt.want)
				}
			}
		})
	}
}