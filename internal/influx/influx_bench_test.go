package influx

import (
	"testing"
)

// Benchmark tests for InfluxDB data marshaling
func BenchmarkInfluxDataMarshal(b *testing.B) {
	m := New()
	m.Name = "weather"
	m.Tags["station"] = "ST-123456"
	m.Tags["location"] = "backyard"
	m.Fields["temp"] = "25.50"
	m.Fields["humidity"] = "60.00"
	m.Fields["pressure"] = "1013.25"
	m.Fields["wind_speed"] = "5.50"
	m.Fields["wind_direction"] = "180"
	m.Timestamp = 1640995200

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.Marshal()
	}
}

func BenchmarkInfluxDataMarshalLargeDataset(b *testing.B) {
	m := New()
	m.Name = "weather_detailed"
	
	// Add many tags
	for i := 0; i < 10; i++ {
		m.Tags[string(rune('a'+i))] = "value"
	}
	
	// Add many fields
	for i := 0; i < 20; i++ {
		m.Fields[string(rune('A'+i))] = "123.45"
	}
	
	m.Timestamp = 1640995200

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.Marshal()
	}
}

func BenchmarkNewInfluxData(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = New()
	}
}

func BenchmarkInfluxDataMarshalParallel(b *testing.B) {
	m := New()
	m.Name = "weather"
	m.Tags["station"] = "ST-123456"
	m.Fields["temp"] = "25.50"
	m.Fields["humidity"] = "60.00"
	m.Timestamp = 1640995200

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = m.Marshal()
		}
	})
}