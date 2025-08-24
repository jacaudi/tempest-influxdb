package config

import (
	"os"
	"testing"
)

// Benchmark tests for configuration loading and validation
func BenchmarkLoadConfig(b *testing.B) {
	// Set up environment variables for consistent testing
	os.Setenv("TEMPEST_INFLUX_INFLUX_URL", "http://localhost:8086/api/v2/write")
	os.Setenv("TEMPEST_INFLUX_INFLUX_TOKEN", "test-token")
	os.Setenv("TEMPEST_INFLUX_INFLUX_BUCKET", "test-bucket")
	defer func() {
		os.Unsetenv("TEMPEST_INFLUX_INFLUX_URL")
		os.Unsetenv("TEMPEST_INFLUX_INFLUX_TOKEN")
		os.Unsetenv("TEMPEST_INFLUX_INFLUX_BUCKET")
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Load("/tmp", "tempest_influx")
	}
}

func BenchmarkConfigValidation(b *testing.B) {
	config := &Config{
		Influx_URL:     "http://localhost:8086/api/v2/write",
		Influx_Token:   "test-token",
		Influx_Bucket:  "test-bucket",
		Listen_Address: ":50222",
		Buffer:         1024,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.Validate()
	}
}

func BenchmarkConfigValidationParallel(b *testing.B) {
	config := &Config{
		Influx_URL:     "http://localhost:8086/api/v2/write",
		Influx_Token:   "test-token",
		Influx_Bucket:  "test-bucket",
		Listen_Address: ":50222",
		Buffer:         1024,
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = config.Validate()
		}
	})
}
