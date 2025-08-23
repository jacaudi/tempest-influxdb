package tempest

import (
	"errors"
	"net"
	"testing"

	"github.com/jacaudi/tempest_influx/internal/config"
	"github.com/jacaudi/tempest_influx/internal/influx"
)

func TestPrecipType_String(t *testing.T) {
	tests := []struct {
		name   string
		precip PrecipType
		want   string
	}{
		{"none", PrecipNone, "none"},
		{"rain", PrecipRain, "rain"},
		{"hail", PrecipHail, "hail"},
		{"rain+hail", PrecipRainHail, "rain+hail"},
		{"unknown", PrecipType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.precip.String(); got != tt.want {
				t.Errorf("PrecipType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseObservationSuccess(t *testing.T) {
	cfg := &config.Config{Debug: false}
	report := Report{
		ReportType: "obs_st",
		Obs: [1][]float64{
			{
				1640995200, // timestamp
				1.5,        // wind_lull
				2.3,        // wind_avg
				3.8,        // wind_gust
				180,        // wind_direction
				3,          // wind_sample_interval
				1013.25,    // station_pressure
				25.5,       // air_temperature
				65.0,       // relative_humidity
				50000,      // illuminance
				5.2,        // uv
				800,        // solar_radiation
				0.5,        // precipitation_accumulation
				0,          // precipitation_type
				5,          // strike_avg_distance
				2,          // strike_count
				3.7,        // battery
				1,          // interval
			},
		},
		StationSerial: "ST-123456",
	}

	m := influx.New()
	err := parseObservation(cfg, report, m)

	if err != nil {
		t.Fatalf("parseObservation() error = %v", err)
	}

	if m.Timestamp != 1640995200 {
		t.Errorf("Expected timestamp 1640995200, got %d", m.Timestamp)
	}

	// Check specific fields
	expectedFields := map[string]bool{
		"battery":            true,
		"dew_point":          true,
		"illuminance":        true,
		"p":                  true,
		"precipitation":      true,
		"precipitation_type": true,
		"solar_radiation":    true,
		"strike_count":       true,
		"strike_distance":    true,
		"temp":               true,
		"uv":                 true,
		"wind_avg":           true,
		"wind_direction":     true,
		"wind_gust":          true,
		"wind_lull":          true,
	}

	for field := range expectedFields {
		if _, exists := m.Fields[field]; !exists {
			t.Errorf("Expected field %s not found", field)
		}
	}

	if m.Fields["temp"] != "25.50" {
		t.Errorf("Expected temp=25.50, got %s", m.Fields["temp"])
	}

	if m.Fields["wind_direction"] != "180" {
		t.Errorf("Expected wind_direction=180, got %s", m.Fields["wind_direction"])
	}
}

func TestParseObservationInsufficientData(t *testing.T) {
	cfg := &config.Config{Debug: false}
	report := Report{
		ReportType: "obs_st",
		Obs: [1][]float64{
			{1640995200, 1.5, 2.3}, // Only 3 fields, need 18
		},
	}

	m := influx.New()
	err := parseObservation(cfg, report, m)

	if err == nil {
		t.Fatal("Expected error for insufficient data, got nil")
	}

	if !errors.Is(err, ErrInsufficientData) {
		t.Errorf("Expected ErrInsufficientData, got %v", err)
	}
}

func TestParseRapidWindSuccess(t *testing.T) {
	cfg := &config.Config{Debug: false}
	report := Report{
		ReportType: "rapid_wind",
		Ob:         [3]float64{1640995200, 5.5, 270},
	}

	m := influx.New()
	err := parseRapidWind(cfg, report, m)

	if err != nil {
		t.Fatalf("parseRapidWind() error = %v", err)
	}

	if m.Timestamp != 1640995200 {
		t.Errorf("Expected timestamp 1640995200, got %d", m.Timestamp)
	}

	if m.Fields["rapid_wind_speed"] != "5.50" {
		t.Errorf("Expected rapid_wind_speed=5.50, got %s", m.Fields["rapid_wind_speed"])
	}

	if m.Fields["rapid_wind_direction"] != "270" {
		t.Errorf("Expected rapid_wind_direction=270, got %s", m.Fields["rapid_wind_direction"])
	}
}

func TestParseRapidWindInsufficientData(t *testing.T) {
	// This test requires directly accessing the parser with a Report that has
	// insufficient data. Since Ob is a fixed-size array [3]float64,
	// we need to simulate insufficient data differently.
	// For now, we'll test that the parser correctly processes valid data
	// and skip the insufficient data test case since the struct itself
	// enforces having 3 elements.
	t.Skip("parseRapidWind uses fixed-size array [3]float64, cannot test insufficient data case")
}

func TestParseValidObsStReport(t *testing.T) {
	cfg := &config.Config{
		Debug:         false,
		Influx_Bucket: "test-bucket",
	}

	jsonData := `{
		"serial_number": "ST-123456",
		"type": "obs_st",
		"obs": [[
			1640995200, 1.5, 2.3, 3.8, 180, 3, 1013.25, 25.5, 65.0, 50000,
			5.2, 800, 0.5, 0, 5, 2, 3.7, 1
		]]
	}`

	addr, _ := net.ResolveUDPAddr("udp", "192.168.1.100:50222")

	m, err := Parse(cfg, addr, []byte(jsonData), len(jsonData))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if m == nil {
		t.Fatal("Expected non-nil InfluxData")
	}

	if m.Name != "weather" {
		t.Errorf("Expected measurement name 'weather', got %s", m.Name)
	}

	if m.Tags["station"] != "ST-123456" {
		t.Errorf("Expected station tag ST-123456, got %s", m.Tags["station"])
	}

	if m.Bucket != "test-bucket" {
		t.Errorf("Expected bucket test-bucket, got %s", m.Bucket)
	}
}

func TestParseValidRapidWindReport(t *testing.T) {
	cfg := &config.Config{
		Debug:                    false,
		Rapid_Wind:               true,
		Influx_Bucket:            "test-bucket",
		Influx_Bucket_Rapid_Wind: "rapid-wind-bucket",
	}

	jsonData := `{
		"serial_number": "ST-123456",
		"type": "rapid_wind",
		"ob": [1640995200, 5.5, 270]
	}`

	addr, _ := net.ResolveUDPAddr("udp", "192.168.1.100:50222")

	m, err := Parse(cfg, addr, []byte(jsonData), len(jsonData))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if m == nil {
		t.Fatal("Expected non-nil InfluxData")
	}

	if m.Name != "weather" {
		t.Errorf("Expected measurement name 'weather', got %s", m.Name)
	}

	if m.Bucket != "rapid-wind-bucket" {
		t.Errorf("Expected bucket rapid-wind-bucket, got %s", m.Bucket)
	}
}

func TestParseRapidWindDisabled(t *testing.T) {
	cfg := &config.Config{
		Debug:         false,
		Rapid_Wind:    false, // Disabled
		Influx_Bucket: "test-bucket",
	}

	jsonData := `{
		"serial_number": "ST-123456",
		"type": "rapid_wind",
		"ob": [1640995200, 5.5, 270]
	}`

	addr, _ := net.ResolveUDPAddr("udp", "192.168.1.100:50222")

	m, err := Parse(cfg, addr, []byte(jsonData), len(jsonData))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if m != nil {
		t.Error("Expected nil InfluxData when rapid wind disabled")
	}
}

func TestParseIgnoredReportTypes(t *testing.T) {
	cfg := &config.Config{Debug: false, Influx_Bucket: "test-bucket"}
	addr, _ := net.ResolveUDPAddr("udp", "192.168.1.100:50222")

	ignoredTypes := []string{"hub_status", "evt_precip", "evt_strike"}

	for _, reportType := range ignoredTypes {
		t.Run(reportType, func(t *testing.T) {
			jsonData := `{"type": "` + reportType + `"}`

			m, err := Parse(cfg, addr, []byte(jsonData), len(jsonData))

			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if m != nil {
				t.Errorf("Expected nil InfluxData for ignored report type %s", reportType)
			}
		})
	}
}

func TestParseInvalidJSON(t *testing.T) {
	cfg := &config.Config{Debug: false}
	addr, _ := net.ResolveUDPAddr("udp", "192.168.1.100:50222")

	invalidJSON := `{"type": "obs_st", "obs": [invalid json}`

	m, err := Parse(cfg, addr, []byte(invalidJSON), len(invalidJSON))

	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}

	if m != nil {
		t.Error("Expected nil InfluxData for invalid JSON")
	}
}

func TestParseUnknownReportType(t *testing.T) {
	cfg := &config.Config{Debug: false, Influx_Bucket: "test-bucket"}
	addr, _ := net.ResolveUDPAddr("udp", "192.168.1.100:50222")

	jsonData := `{"type": "unknown_type"}`

	m, err := Parse(cfg, addr, []byte(jsonData), len(jsonData))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if m != nil {
		t.Error("Expected nil InfluxData for unknown report type")
	}
}

// Benchmark tests
func BenchmarkParseObsStReport(b *testing.B) {
	cfg := &config.Config{Debug: false, Influx_Bucket: "test-bucket"}
	addr, _ := net.ResolveUDPAddr("udp", "192.168.1.100:50222")

	jsonData := `{
		"serial_number": "ST-123456",
		"type": "obs_st",
		"obs": [[
			1640995200, 1.5, 2.3, 3.8, 180, 3, 1013.25, 25.5, 65.0, 50000,
			5.2, 800, 0.5, 0, 5, 2, 3.7, 1
		]]
	}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse(cfg, addr, []byte(jsonData), len(jsonData))
	}
}

func BenchmarkParseRapidWindReport(b *testing.B) {
	cfg := &config.Config{Debug: false, Rapid_Wind: true, Influx_Bucket: "test-bucket"}
	addr, _ := net.ResolveUDPAddr("udp", "192.168.1.100:50222")

	jsonData := `{
		"serial_number": "ST-123456",
		"type": "rapid_wind",
		"ob": [1640995200, 5.5, 270]
	}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Parse(cfg, addr, []byte(jsonData), len(jsonData))
	}
}
