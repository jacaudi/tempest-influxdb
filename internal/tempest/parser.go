package tempest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net"

	"github.com/de-wax/go-pkg/dewpoint"
	"github.com/jacaudi/tempest-influxdb/internal/config"
	"github.com/jacaudi/tempest-influxdb/internal/influx"
)

// Error constants for better error handling
var (
	ErrInvalidReportType   = errors.New("invalid or unsupported report type")
	ErrInsufficientData    = errors.New("insufficient observation data")
	ErrDewPointCalculation = errors.New("dewpoint calculation failed")
)

// PrecipType represents different types of precipitation
type PrecipType int

const (
	PrecipNone PrecipType = iota
	PrecipRain
	PrecipHail
	PrecipRainHail
)

// String returns the string representation of precipitation type
func (p PrecipType) String() string {
	types := []string{"none", "rain", "hail", "rain+hail"}
	if int(p) < len(types) {
		return types[p]
	}
	return "unknown"
}

// PrecipitationTypeStrings provides backward compatibility
var PrecipitationTypeStrings = []string{"none", "rain", "hail", "rain+hail"}

// Report represents a weather report from Tempest station
type Report struct {
	StationSerial    string       `json:"serial_number,omitempty"`
	ReportType       string       `json:"type"`
	HubSerial        string       `json:"hub_sn,omitempty"`
	Obs              [1][]float64 `json:"obs,omitempty"`
	Ob               [3]float64   `json:"ob,omitempty"`
	Evt              []float64    `json:"evt,omitempty"`
	FirmwareRevision int          `json:"firmware_revision,omitempty"`
	Uptime           int          `json:"uptime,omitempty"`
	Timestamp        int          `json:"timestamp,omitempty"`
	ResetFlags       string       `json:"reset_flags,omitempty"`
	Seq              int          `json:"seq,omitempty"`
	Fs               []float64    `json:"fs,omitempty"`
	Radio_Stats      []float64    `json:"radio_stats,omitempty"`
	Mqtt_Stats       []float64    `json:"mqtt_stats,omitempty"`
	Voltage          float64      `json:"voltage,omitempty"`
	RSSI             float64      `json:"rssi,omitempty"`
	HubRSSI          float64      `json:"hub_rssi,omitempty"`
	SensorStatus     int          `json:"sensor_status,omitempty"`
	Debug            int          `json:"debug,omitempty"`
}

// parseObservation parses Tempest observation data
func parseObservation(cfg *config.Config, report Report, m *influx.Data) error {
	type Obs struct {
		Timestamp                 int64   // seconds
		WindLull                  float64 // m/s
		WindAvg                   float64 // m/s
		WindGust                  float64 // m/s
		WindDirection             int     // Degrees
		WindSampleInterval        int     // seconds
		StationPressure           float64 // MB
		AirTemperature            float64 // C
		RelativeHumidity          float64 // %
		Illuminance               int     // Lux
		UV                        float64 // Index
		SolarRadiation            int     // W/m*2
		PrecipitationAccumulation float64 // mm
		PrecipitationType         int     //
		StrikeAvgDistance         int     // km
		StrikeCount               int     // count
		Battery                   float64 // Voltags
		Interval                  int     // Minutes
	}
	var observation Obs

	if len(report.Obs[0]) < 18 {
		return fmt.Errorf("%w: expected 18 fields, got %d", ErrInsufficientData, len(report.Obs[0]))
	}

	data := report.Obs[0]
	observation.Timestamp = int64(data[0])
	observation.WindLull = data[1]
	observation.WindAvg = data[2]
	observation.WindGust = data[3]
	observation.WindDirection = int(math.Round(data[4]))
	observation.WindSampleInterval = int(math.Round(data[5]))
	observation.StationPressure = data[6]
	observation.AirTemperature = data[7]
	observation.RelativeHumidity = data[8]
	observation.Illuminance = int(math.Round(data[9]))
	observation.UV = data[10]
	observation.SolarRadiation = int(math.Round(data[11]))
	observation.PrecipitationAccumulation = data[12]
	observation.PrecipitationType = int(math.Round(data[13]))
	observation.StrikeAvgDistance = int(math.Round(data[14]))
	observation.StrikeCount = int(math.Round(data[15]))
	observation.Battery = data[16]
	observation.Interval = int(math.Round(data[17]))
	if cfg.Debug {
		log.Printf("OBS_ST %+v %+v", report, observation)
	}

	// Calculate Dew Point from RH and Temp
	dp, err := dewpoint.Calculate(observation.AirTemperature, observation.RelativeHumidity)
	if err != nil {
		log.Printf("dewpoint.Calculate(%f, %f): %v", observation.AirTemperature, observation.RelativeHumidity, err)
	}

	m.Timestamp = observation.Timestamp
	// Set fields and sort into alphabetical order to keep InfluxDB happy
	m.Fields = map[string]string{
		"battery":            fmt.Sprintf("%.2f", observation.Battery),
		"dew_point":          fmt.Sprintf("%.2f", dp),
		"illuminance":        fmt.Sprintf("%d", observation.Illuminance),
		"p":                  fmt.Sprintf("%.2f", observation.StationPressure),
		"precipitation":      fmt.Sprintf("%.2f", observation.PrecipitationAccumulation),
		"precipitation_type": fmt.Sprintf("%d", observation.PrecipitationType),
		"solar_radiation":    fmt.Sprintf("%d", observation.SolarRadiation),
		"strike_count":       fmt.Sprintf("%d", observation.StrikeCount),
		"strike_distance":    fmt.Sprintf("%d", observation.StrikeAvgDistance),
		"temp":               fmt.Sprintf("%.2f", observation.AirTemperature),
		"uv":                 fmt.Sprintf("%.2f", observation.UV),
		"wind_avg":           fmt.Sprintf("%.2f", observation.WindAvg),
		"wind_direction":     fmt.Sprintf("%d", observation.WindDirection),
		"wind_gust":          fmt.Sprintf("%.2f", observation.WindGust),
		"wind_lull":          fmt.Sprintf("%.2f", observation.WindLull),
	}
	return nil
}

// parseRapidWind parses Tempest rapid wind data
func parseRapidWind(cfg *config.Config, report Report, m *influx.Data) error {
	type RapidWind struct {
		Timestamp     int64   // seconds
		WindSpeed     float64 // m/s
		WindDirection int     // degrees
	}
	var rapidWind RapidWind

	if len(report.Ob) < 3 {
		return fmt.Errorf("%w: expected 3 fields, got %d", ErrInsufficientData, len(report.Ob))
	}

	rapidWind.Timestamp = int64(report.Ob[0])
	rapidWind.WindSpeed = report.Ob[1]
	rapidWind.WindDirection = int(math.Round(report.Ob[2]))
	if cfg.Debug {
		log.Printf("RAPID_WIND %+v %+v", report, rapidWind)
	}

	m.Timestamp = rapidWind.Timestamp
	m.Fields = map[string]string{
		"rapid_wind_speed":     fmt.Sprintf("%.2f", rapidWind.WindSpeed),
		"rapid_wind_direction": fmt.Sprintf("%d", rapidWind.WindDirection),
	}
	return nil
}

// parseRainStartEvent parses precipitation start events
func parseRainStartEvent(cfg *config.Config, report Report, m *influx.Data) error {
	type RainEvent struct {
		Timestamp int64 // seconds
	}
	var rainEvent RainEvent

	// evt_precip uses "evt" field with single timestamp
	if len(report.Evt) < 1 {
		return fmt.Errorf("%w: expected 1 field in evt", ErrInsufficientData)
	}

	rainEvent.Timestamp = int64(report.Evt[0])
	if cfg.Debug {
		log.Printf("EVT_PRECIP %+v %+v", report, rainEvent)
	}

	m.Timestamp = rainEvent.Timestamp
	m.Fields = map[string]string{
		"rain_start_event": "1",
	}
	return nil
}

// parseLightningStrike parses lightning strike events
func parseLightningStrike(cfg *config.Config, report Report, m *influx.Data) error {
	type LightningStrike struct {
		Timestamp int64 // seconds
		Distance  int   // km
		Energy    int   // energy value
	}
	var lightning LightningStrike

	// evt_strike uses "evt" field with [timestamp, distance, energy]
	if len(report.Evt) < 3 {
		return fmt.Errorf("%w: expected 3 fields in evt", ErrInsufficientData)
	}

	lightning.Timestamp = int64(report.Evt[0])
	lightning.Distance = int(math.Round(report.Evt[1]))
	lightning.Energy = int(math.Round(report.Evt[2]))
	if cfg.Debug {
		log.Printf("EVT_STRIKE %+v %+v", report, lightning)
	}

	m.Timestamp = lightning.Timestamp
	m.Fields = map[string]string{
		"lightning_distance": fmt.Sprintf("%d", lightning.Distance),
		"lightning_energy":   fmt.Sprintf("%d", lightning.Energy),
	}
	return nil
}

// parseAirObservation parses AIR sensor observations
func parseAirObservation(cfg *config.Config, report Report, m *influx.Data) error {
	type AirObs struct {
		Timestamp         int64   // seconds
		StationPressure   float64 // MB
		AirTemperature    float64 // C
		RelativeHumidity  float64 // %
		StrikeCount       int     // count
		StrikeAvgDistance int     // km
		Battery           float64 // volts
		ReportInterval    int     // minutes
	}
	var airObs AirObs

	if len(report.Obs[0]) < 8 {
		return fmt.Errorf("%w: expected 8 fields, got %d", ErrInsufficientData, len(report.Obs[0]))
	}

	data := report.Obs[0]
	airObs.Timestamp = int64(data[0])
	airObs.StationPressure = data[1]
	airObs.AirTemperature = data[2]
	airObs.RelativeHumidity = data[3]
	airObs.StrikeCount = int(math.Round(data[4]))
	airObs.StrikeAvgDistance = int(math.Round(data[5]))
	airObs.Battery = data[6]
	airObs.ReportInterval = int(math.Round(data[7]))
	if cfg.Debug {
		log.Printf("OBS_AIR %+v %+v", report, airObs)
	}

	// Calculate Dew Point from RH and Temp
	dp, err := dewpoint.Calculate(airObs.AirTemperature, airObs.RelativeHumidity)
	if err != nil {
		log.Printf("dewpoint.Calculate(%f, %f): %v", airObs.AirTemperature, airObs.RelativeHumidity, err)
	}

	m.Timestamp = airObs.Timestamp
	m.Fields = map[string]string{
		"air_temperature":    fmt.Sprintf("%.2f", airObs.AirTemperature),
		"battery":            fmt.Sprintf("%.2f", airObs.Battery),
		"dew_point":          fmt.Sprintf("%.2f", dp),
		"humidity":           fmt.Sprintf("%.2f", airObs.RelativeHumidity),
		"pressure":           fmt.Sprintf("%.2f", airObs.StationPressure),
		"strike_count":       fmt.Sprintf("%d", airObs.StrikeCount),
		"strike_distance":    fmt.Sprintf("%d", airObs.StrikeAvgDistance),
	}
	return nil
}

// parseSkyObservation parses Sky sensor observations  
func parseSkyObservation(cfg *config.Config, report Report, m *influx.Data) error {
	type SkyObs struct {
		Timestamp                 int64   // seconds
		Illuminance               int     // lux
		UV                        int     // index
		RainAccumulation          float64 // mm
		WindLull                  float64 // m/s
		WindAvg                   float64 // m/s
		WindGust                  float64 // m/s
		WindDirection             int     // degrees
		Battery                   float64 // volts
		ReportInterval            int     // minutes
		SolarRadiation            int     // W/m^2
		LocalDayRainAccumulation  float64 // mm (can be null)
		PrecipitationType         int     // 0-2
		WindSampleInterval        int     // seconds
	}
	var skyObs SkyObs

	if len(report.Obs[0]) < 14 {
		return fmt.Errorf("%w: expected 14 fields, got %d", ErrInsufficientData, len(report.Obs[0]))
	}

	data := report.Obs[0]
	skyObs.Timestamp = int64(data[0])
	skyObs.Illuminance = int(math.Round(data[1]))
	skyObs.UV = int(math.Round(data[2]))
	skyObs.RainAccumulation = data[3]
	skyObs.WindLull = data[4]
	skyObs.WindAvg = data[5]
	skyObs.WindGust = data[6]
	skyObs.WindDirection = int(math.Round(data[7]))
	skyObs.Battery = data[8]
	skyObs.ReportInterval = int(math.Round(data[9]))
	skyObs.SolarRadiation = int(math.Round(data[10]))
	skyObs.LocalDayRainAccumulation = data[11] // may be null
	skyObs.PrecipitationType = int(math.Round(data[12]))
	skyObs.WindSampleInterval = int(math.Round(data[13]))
	if cfg.Debug {
		log.Printf("OBS_SKY %+v %+v", report, skyObs)
	}

	m.Timestamp = skyObs.Timestamp
	m.Fields = map[string]string{
		"battery":            fmt.Sprintf("%.2f", skyObs.Battery),
		"illuminance":        fmt.Sprintf("%d", skyObs.Illuminance),
		"precipitation":      fmt.Sprintf("%.2f", skyObs.RainAccumulation),
		"precipitation_type": fmt.Sprintf("%d", skyObs.PrecipitationType),
		"solar_radiation":    fmt.Sprintf("%d", skyObs.SolarRadiation),
		"uv":                 fmt.Sprintf("%d", skyObs.UV),
		"wind_avg":           fmt.Sprintf("%.2f", skyObs.WindAvg),
		"wind_direction":     fmt.Sprintf("%d", skyObs.WindDirection),
		"wind_gust":          fmt.Sprintf("%.2f", skyObs.WindGust),
		"wind_lull":          fmt.Sprintf("%.2f", skyObs.WindLull),
	}
	
	// Add daily rain accumulation if not null
	if !math.IsNaN(skyObs.LocalDayRainAccumulation) {
		m.Fields["daily_rain"] = fmt.Sprintf("%.2f", skyObs.LocalDayRainAccumulation)
	}
	
	return nil
}

// parseDeviceStatus parses device status messages
func parseDeviceStatus(cfg *config.Config, report Report, m *influx.Data) error {
	if cfg.Debug {
		log.Printf("DEVICE_STATUS %+v", report)
	}

	m.Timestamp = int64(report.Timestamp)
	m.Fields = map[string]string{
		"device_uptime":     fmt.Sprintf("%d", report.Uptime),
		"device_voltage":    fmt.Sprintf("%.2f", report.Voltage),
		"device_rssi":       fmt.Sprintf("%.2f", report.RSSI),
		"device_hub_rssi":   fmt.Sprintf("%.2f", report.HubRSSI),
		"sensor_status":     fmt.Sprintf("%d", report.SensorStatus),
		"firmware_revision": fmt.Sprintf("%d", report.FirmwareRevision),
	}
	return nil
}

// parseHubStatus parses hub status messages
func parseHubStatus(cfg *config.Config, report Report, m *influx.Data) error {
	if cfg.Debug {
		log.Printf("HUB_STATUS %+v", report)
	}

	m.Timestamp = int64(report.Timestamp)
	m.Fields = map[string]string{
		"hub_uptime":        fmt.Sprintf("%d", report.Uptime),
		"hub_rssi":          fmt.Sprintf("%.2f", report.RSSI),
		"firmware_revision": fmt.Sprintf("%d", report.FirmwareRevision),
		"sequence":          fmt.Sprintf("%d", report.Seq),
	}
	
	// Add reset flags if present
	if report.ResetFlags != "" {
		m.Fields["reset_flags"] = report.ResetFlags
	}
	
	return nil
}

// Parse parses weather data from Tempest station
func Parse(cfg *config.Config, addr *net.UDPAddr, b []byte, n int) (m *influx.Data, err error) {
	var report Report
	decoder := json.NewDecoder(bytes.NewReader(b[:n]))
	err = decoder.Decode(&report)
	if err != nil {
		err = fmt.Errorf("ERROR Could not Unmarshal %d bytes from %v: %v: %v", n, addr, err, string(b[:n]))
		return
	}

	m = influx.New()

	m.Bucket = cfg.Influx_Bucket

	switch report.ReportType {
	case "obs_st":
		m.Name = "weather"
		if err = parseObservation(cfg, report, m); err != nil {
			return nil, fmt.Errorf("parsing observation: %w", err)
		}
		m.Tags["station"] = report.StationSerial
	case "rapid_wind":
		if !cfg.Rapid_Wind {
			return nil, nil
		}
		m.Name = "weather"
		if err = parseRapidWind(cfg, report, m); err != nil {
			return nil, fmt.Errorf("parsing rapid wind: %w", err)
		}
		m.Tags["station"] = report.StationSerial
		if cfg.Influx_Bucket_Rapid_Wind != "" {
			m.Bucket = cfg.Influx_Bucket_Rapid_Wind
		}
	case "evt_precip":
		m.Name = "weather_events"
		if err = parseRainStartEvent(cfg, report, m); err != nil {
			return nil, fmt.Errorf("parsing rain start event: %w", err)
		}
		m.Tags["station"] = report.StationSerial
		m.Tags["event_type"] = "rain_start"
	case "evt_strike":
		m.Name = "weather_events"
		if err = parseLightningStrike(cfg, report, m); err != nil {
			return nil, fmt.Errorf("parsing lightning strike: %w", err)
		}
		m.Tags["station"] = report.StationSerial
		m.Tags["event_type"] = "lightning_strike"
	case "obs_air":
		m.Name = "weather"
		if err = parseAirObservation(cfg, report, m); err != nil {
			return nil, fmt.Errorf("parsing AIR observation: %w", err)
		}
		m.Tags["station"] = report.StationSerial
		m.Tags["sensor_type"] = "air"
	case "obs_sky":
		m.Name = "weather"
		if err = parseSkyObservation(cfg, report, m); err != nil {
			return nil, fmt.Errorf("parsing Sky observation: %w", err)
		}
		m.Tags["station"] = report.StationSerial
		m.Tags["sensor_type"] = "sky"
	case "device_status":
		m.Name = "device_status"
		if err = parseDeviceStatus(cfg, report, m); err != nil {
			return nil, fmt.Errorf("parsing device status: %w", err)
		}
		m.Tags["device"] = report.StationSerial
		m.Tags["hub"] = report.HubSerial
	case "hub_status":
		m.Name = "hub_status"
		if err = parseHubStatus(cfg, report, m); err != nil {
			return nil, fmt.Errorf("parsing hub status: %w", err)
		}
		m.Tags["hub"] = report.StationSerial
	default:
		return nil, nil
	}

	return
}
