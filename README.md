# liquidctl-exporter

A Prometheus exporter for [liquidctl](https://github.com/liquidctl/liquidctl), exposing metrics about AIO coolers connected to your system.

This exporter collects metrics such as fan speeds, temperatures, and device information from liquidctl-compatible devices.

## Features

- Exposes Prometheus metrics for liquidctl devices
- Device information including bus, address, vendor ID, product ID, and serial number
- Fan speed metrics in RPM
- Temperature metrics in Celsius
- Device attributes and status information
- Static binary builds for Linux, macOS, and Windows

## Metrics

The exporter exposes the following metrics:

| Metric Name | Description | Labels |
|-------------|-------------|--------|
| `liquidctl_up` | Was the exporter able to run liquidctl and collect metrics? | - |
| `liquidctl_meta_info` | Information about the device | `bus`, `address`, `description`, `vendor_id`, `product_id`, `release_number`, `serial_number`, `driver` |
| `liquidctl_attributes_info` | The device attributes | `bus`, `address`, `key`, `value`, `unit` |
| `liquidctl_fan_speed_rpm` | The speed of the fans connected to the controller | `bus`, `address`, `key` |
| `liquidctl_temperature_celsius` | The temperature of the sensors connected to the controller | `bus`, `address`, `key` |

## Requirements

- Linux (recommended), macOS, or Windows
- Go 1.26 or later
- Docker (for containerized deployment)
- liquidctl installed on the host system (for Docker)

## Installation

### Building from Source

```bash
# Clone the repository
git clone https://github.com/thelande/liquidctl-exporter.git
cd liquidctl-exporter

# Build the binary
make build

# The binary will be available at ./output/liquidctl-exporter
./liquidctl-exporter
```

### Using Docker

```bash
# Build the Docker image
make docker-build

# Run the container
docker run -p 9530:9530 --device=/dev/hidraw1:/dev/hidraw1 -d thelande/liquidctl-exporter
```

#### Passing /dev/hidraw1 to the Container

To expose liquidctl device metrics to the container, you must pass the USB device file (e.g., `/dev/hidraw1`) through to the container. This device file provides access to the liquidctl-compatible hardware.

#### Docker Compose Example

Create a `docker-compose.yml` file:

```yaml
services:
  liquidctl-exporter:
    image: thelande/liquidctl-exporter
    container_name: liquidctl-exporter
    restart: unless-stopped
    ports:
      - "9530:9530"
    devices:
      - /dev/hidraw1:/dev/hidraw1
```

Start the service:

```bash
docker-compose up -d
```

## Configuration

### Command-Line Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--web.telemetry-path` | Path under which to expose metrics | `/metrics` |
| `--web.listen-address` | Address to listen on | `:9530` |

### Environment Variables

The exporter supports the following environment variables:

| Variable | Description |
|----------|-------------|
| `WEB_LISTEN_ADDRESS` | Address to listen on |
| `WEB_TELEMETRY_PATH` | Path under which to expose metrics |

### Example with Custom Port

```bash
./liquidctl-exporter --web.listen-address=:9531
```

## Usage

### Prometheus Configuration

Add the following to your Prometheus `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'liquidctl-exporter'
    static_configs:
      - targets: ['localhost:9530']
```

### Querying Metrics

Once Prometheus is running, you can query the metrics:

```bash
curl http://localhost:9090/api/v1/query?query=liquidctl_up
curl http://localhost:9090/api/v1/query?query=liquidctl_fan_speed_rpm
```

## Development

### Building

```bash
make build
```

### Cross-Compilation

```bash
make crossbuild
```

### Testing

```bash
make test
```

### Linting

```bash
make fmt
```

### Docker Images

```bash
# Build local image
docker buildx bake image-local

# Build for multiple architectures
docker buildx bake image-all

# Push to Docker Hub
docker buildx bake image-local --push
```

## License

This project is licensed under the Apache License, Version 2.0 - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Links

- [liquidctl Repository](https://github.com/liquidctl/liquidctl)
- [Prometheus](https://prometheus.io/)
- [Prometheus Exporter Toolkit](https://github.com/prometheus/exporter-toolkit)
