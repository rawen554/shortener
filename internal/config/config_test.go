package config

import (
	"os"
	"reflect"
	"testing"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		want              *ServerConfig
		name              string
		configFileContent string
	}{
		{
			name: "test configs merge",
			configFileContent: `{
				"server_address": "localhost:8080",
				"base_url": "http://localhost",
				"file_storage_path": "/path/to/file.db",
				"database_dsn": "",
				"enable_https": true
			}`,
			want: &ServerConfig{
				RunAddr:         ":8080",
				RedirectBaseURL: "http://localhost:8080",
				FileStoragePath: "/path/to/file.db",
				DatabaseDSN:     "",
				Secret:          "b4952c3809196592c026529df00774e46bfb5be0",
				TLSCertPath:     "./certs/cert.pem",
				TLSKeyPath:      "./certs/private.pem",
				EnableHTTPS:     true,
				ProfileMode:     false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile, err := os.CreateTemp(t.TempDir(), "config-*.json")
			if err != nil {
				t.Error(err)
			}

			if _, err := configFile.WriteString(tt.configFileContent); err != nil {
				t.Error(err)
			}

			if err := configFile.Close(); err != nil {
				t.Error(err)
			}

			//nolint // only for testing purposes hack
			os.Args = append(os.Args, "-c", configFile.Name())

			got, err := ParseFlags()
			if err != nil {
				t.Error(err)
			}
			tt.want.Config = configFile.Name()

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseFlags() = %v, want %v", got, tt.want)
			}
		})
	}
}
