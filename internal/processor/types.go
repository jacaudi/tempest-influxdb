package processor

import (
	"context"
	"net"
	"net/http"
)

// UDPListener interface for UDP operations
type UDPListener interface {
	ReadFromUDP([]byte) (int, *net.UDPAddr, error)
	SetReadDeadline(deadline interface{}) error
	Close() error
}

// HTTPClient interface for HTTP operations
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// Logger interface for structured logging
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
}

// PacketProcessor interface for processing weather data packets
type PacketProcessor interface {
	ProcessPacket(ctx context.Context, addr *net.UDPAddr, data []byte, length int) error
}

// ConfigValidator interface for configuration validation
type ConfigValidator interface {
	Validate() error
}

// WeatherStation represents a weather station configuration
type WeatherStation struct {
	Serial   string
	Name     string
	Location string
}