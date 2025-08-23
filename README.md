# Tempest InfluxDB Collector

A high-performance weather data collector that receives UDP broadcasts from WeatherFlow Tempest weather stations and forwards the data to InfluxDB for storage and analysis.

[![Go Report Card](https://goreportcard.com/badge/github.com/jacaudi/tempest-influxdb?style=flat-square)](https://goreportcard.com/report/github.com/jacaudi/tempest-influxdb)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/jacaudi/tempest-influxdb)

## Features

- **UDP Listener**: Receives real-time weather data from [Tempest Weather System](https://shop.weatherflow.com/products/tempest) broadcasts
- **InfluxDB Integration**: Forwards data using InfluxDB line protocol
- **Performance Optimized**: Buffer pooling, optimized HTTP client, efficient parsing
- **Flexible Configuration**: YAML files, environment variables, and CLI flags
- **Optional Rapid Wind**: High-frequency wind data collection (every 3s)
- **Graceful Shutdown**: Proper signal handling

Requires Docker host networking to receive UDP broadcasts.

## Broadcast Formats

UDP broadcast formats are documented [here](https://weatherflow.github.io/Tempest/api/udp.html). Key messages:
- `obs_st`: Full weather data (every minute)
- `rapid_wind`: Instantaneous wind data (every few seconds)

## Configuration

Configuration priority: CLI flags > environment variables > YAML file (`/config/tempest-influxdb.yml`)

| Value                              | Config File              | Environment        | Flag                       | Required | Default                 |
|------------------------------------|--------------------------|--------------------|----------------------------|----------|-------------------------|
| InfluxDB base URL                  | influx_url               | INFLUX_URL         | --influx_url               | Yes      | https://localhost:8086  |
| InfluxDB organization              | influx_org               | INFLUX_ORG         | --influx_org               | Yes      | -                       |
| Influx authentication token        | influx_token             | INFLUX_TOKEN       | --influx_token             | Yes      | -                       |
| Influx bucket                      | influx_bucket            | INFLUX_BUCKET      | --influx_bucket            | Yes      | -                       |
| Read buffer size                   | buffer                   | BUFFER             | --buffer                   | No       | 10240                   |
| Listen Address                     | listen_address           | LISTEN_ADDRESS     | --listen_address           | No       | :50222                  |
| InfluxDB API path                  | influx_api_path          | INFLUX_API_PATH    | --influx_api_path          | No       | /api/v2/write           |
| Influx bucket for rapid wind       | influx_bucket_rapid_wind | INFLUX_BUCKET_RAPID_WIND | --influx_bucket_rapid_wind | No       | -                       |
| Verbose logging                    | verbose                  | VERBOSE            | -v, --verbose              | No       | false (true if debug)   |
| Debug logging                      | debug                    | DEBUG              | -d, --debug                | No       | false                   |
| Raw UDP packet logging             | raw_udp                  | RAW_UDP            | --raw_udp                  | No       | false                   |
| Do not send packets                | noop                     | NOOP               | -n, --noop                 | No       | false                   |
| Send rapid wind reports (every 3s) | rapid_wind               | RAPID_WIND         | --rapid_wind               | No       | false                   |

## Examples

### Docker Compose

```yaml
services:
  tempest-influxdb:
    image: "jacaudi/tempest-influxdb:latest"
    network_mode: host
    environment:
      INFLUX_URL: "https://metrics.example.com"
      INFLUX_TOKEN: "SOMEARBITRARYSTRING"
      INFLUX_BUCKET: "weather"
      INFLUX_ORG: "myorg"
    ports:
      - 50222/udp
```

## Credits

Original Source Code and Ideas by [jchonig/tempest_influxdb](https://github.com/jchonig/tempest_influxdb)
UDP processing based on [udpproxy](https://github.com/Akagi201/udpproxy)


## License

MIT License - see [LICENSE](LICENSE) file.
