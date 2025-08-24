package influx

import (
	"testing"
)

// Test InfluxData marshaling
func TestInfluxDataMarshal(t *testing.T) {
	m := New()
	m.Name = "weather"
	m.Tags["station"] = "ST-123"
	m.Fields["temp"] = "25.5"
	m.Fields["humidity"] = "60.0"
	m.Timestamp = 1640995200

	line := m.Marshal()
	expected := "weather,station=ST-123 humidity=60.0,temp=25.5 1640995200\n"

	if line != expected {
		t.Errorf("InfluxData.Marshal() = %v, want %v", line, expected)
	}
}
