# Tempest InfluxDB Collector

A high-performance weather data collector that receives UDP broadcasts from WeatherFlow Tempest weather stations and forwards the data to InfluxDB for storage and analysis.

[![Go Report Card](https://goreportcard.com/badge/github.com/jacaudi/tempest-influxdb?style=flat-square)](https://goreportcard.com/report/github.com/jacaudi/tempest-influxdb)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/jacaudi/tempest-influxdb)

## Features

- **UDP Listener**: Receives real-time weather data broadcasts from Tempest stations
- **InfluxDB Integration**: Converts and forwards data using InfluxDB line protocol
- **Structured Logging**: Uses Go's structured logging with configurable levels
- **Performance Optimized**: Includes buffer pooling, optimized HTTP client, and efficient data parsing
- **Configuration Flexible**: Supports YAML files, environment variables, and command-line flags
- **Rapid Wind Support**: Optional high-frequency wind data collection
- **Graceful Shutdown**: Proper signal handling for clean service termination

The [Tempest Weather System](https://shop.weatherflow.com/products/tempest) sends UDP broadcasts with weather data and system status periodically. This program receives those broadcasts and generates InfluxDB wire protocol messages to import the data into InfluxDB.

Docker host networking is required to receive the UDP broadcasts, unless some type of proxy is used.

## Tempest WX Broadcast Formats

The format of the UDP broadcasts are documented
[here](https://weatherflow.github.io/Tempest/api/udp.html).

The reports of interest are the `obs_st` message which has full
weather data and the `rapid_wind` has instantaneous Wind data.
The later is generated every few seconds and the former once a minute.

## Configuration

There are three ways to pass configuration information:

  1. A optional YAML configuration file may be provided in /config/tempest-influxdb.yml* which is read at startup.
  2. Environement variables as described in the table below.  These override configuration file data.
  3. Command line flags, also described in the table below.  These override configuration file adata and environment variables.

| Value                              | Config File              | Environment                             | Flag                       | Default                             |
|------------------------------------|--------------------------|-----------------------------------------|----------------------------|-------------------------------------|
| Read buffer size                   | buffer                   | TEMPEST_INFLUX_BUFFER                   | --buffer                   | 10240                               |
| Listen Address                     | listen_address           | TEMPEST_INFLUX_LISTEN_ADDRESS           | --listen_address           | :50222                              |
| InfluxDB write URL                 | influx_url               | TEMPEST_INFLUX_INFLUX_URL               | --influx_url               | https://localhost:8086/api/v2/write |
| Influx authentication token        | influx_token             | TEMPEST_INFLUX_INFLUX_TOKEN             | --influx_token             |                                     |
| Influx bucket                      | influx_bucket            | TEMPEST_INFLUX_INFLUX_BUCKET            | --influx_bucket            |                                     |
| Influx bucket for rapid wind       | influx_bucket_rapid_wind | TEMPEST_INFLUX_INFLUX_BUCKET_RAPID_WIND | --influx_bucket_rapid_wind |                                     |
| Verbose logging                    | verbose                  | TEMPEST_INFLUX_VERBOSE                  | -v, --verbose              | False (True if Debug set)           |
| Debug logging                      | debug                    | TEMPEST_INFLUX_DEBUG                    | -d, --debug                | False                               |
| Do not send packets                | noop                     | TEMPEST_INFLUX_NOOP                     | -n, --noop                 | False                               |
| Send rapid wind reports (every 3s) | rapid_wind               | TEMPEST_RAPID_WIND                      | -rapid_wind                | False                               |

Notes:

   + *influx_token* is required by *InfluxDB* or *Telegraf* to authenticate requests.
   + *influx_bucket* is not required if configured on the receiving end.

## TODO

 + [ ] Hack around firmware_version being and int and string in
       different packets
 + [ ] Optionally send `device_status` and `hub_status` data
   + [ ] Allow specification of a bucket
   + [ ] Structure config?

## Examples

### docker-compose.yml

Following is a sample docker-compose file to run this container.

```yaml
version: "3"

services:
  tempest-influxdb:
    image: "jchonig/tempest-influxdb:latest"
    network_mode: host
    environment:
      TEMPEST_INFLUX_INFLUX_URL: "https://metrics.example.com:8086/api/v2/write"
      TEMPEST_INFLUX_INFLUX_TOKEN: "SOMEARBITRARYSTRING"
      TEMPEST_INFLUX_INFLUX_BUCKET: "weather"
    ports:
      - 50222/udp
```

### Telegraf

The output is designed to be passed to Telegraf for forwarding to
InfluxDB via the influxdb_v2_listener, for example:

```
[[inputs.influxdb_v2_listener]]
  service_address = ":8086"
  tls_cert = "/etc/telegraf/keys/cert.pem"
  tls_key = "/etc/telegraf/keys/key.pem"
  token = "SOMEARBITRARYSTRING"
```

## Credits

Core UDP packet processing code based on code in [udpproxy](https://github.com/Akagi201/udpproxy)

## Development

This project uses a local module name (`tempest-influx`) for development. When ready to publish as a Go module, update the `go.mod` file to use the full GitHub path:

```bash
# Update module name in go.mod
sed -i 's/module tempest-influx/module github.com\/jacaudi\/tempest-influxdb/' go.mod

# Update import statements in all Go files
find . -name "*.go" -exec sed -i 's|tempest-influx/internal/|github.com/jacaudi/tempest-influxdb/internal/|g' {} +

# Clean up dependencies
go mod tidy
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
# Trigger rebuild
