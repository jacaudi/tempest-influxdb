package config

import (
	"testing"
)

// Test configuration validation
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Influx_URL:     "http://localhost:8086/api/v2/write",
				Influx_Token:   "test-token",
				Influx_Bucket:  "test-bucket",
				Listen_Address: ":50222",
				Buffer:         1024,
			},
			wantErr: false,
		},
		{
			name: "missing URL",
			config: &Config{
				Influx_Token:   "test-token",
				Influx_Bucket:  "test-bucket",
				Listen_Address: ":50222",
				Buffer:         1024,
			},
			wantErr: true,
		},
		{
			name: "invalid URL",
			config: &Config{
				Influx_URL:     "://invalid-url",
				Influx_Token:   "test-token",
				Influx_Bucket:  "test-bucket",
				Listen_Address: ":50222",
				Buffer:         1024,
			},
			wantErr: true,
		},
		{
			name: "invalid buffer size",
			config: &Config{
				Influx_URL:     "http://localhost:8086/api/v2/write",
				Influx_Token:   "test-token",
				Influx_Bucket:  "test-bucket",
				Listen_Address: ":50222",
				Buffer:         0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}