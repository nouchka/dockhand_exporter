# dockhand_exporter

Prometheus exporter for [Dockhand](https://dockhand.dev) written in Go.

It scrapes the `/api/dashboard/stats` endpoint of a Dockhand instance and exposes per-host metrics on `:9090/metrics`.

## Configuration

| Environment variable | Required | Default | Description |
|---|---|---|---|
| `DOCKHAND_URL` | ✅ | — | Base URL of the Dockhand instance (e.g. `http://host:3002`) |
| `DOCKHAND_TOKEN` | ✅ | — | Bearer token for the Dockhand API |
| `LISTEN_ADDR` | | `:9090` | Address the exporter listens on |

## Exposed metrics

All metrics carry the labels `id`, `name`, and `host` (Docker host name as configured in Dockhand).

| Metric | Type | Description |
|---|---|---|
| `dockhand_host_online` | Gauge | 1 if the host is reachable, 0 otherwise |
| `dockhand_host_cpu_percent` | Gauge | CPU usage % |
| `dockhand_host_memory_percent` | Gauge | Memory usage % |
| `dockhand_host_memory_used_bytes` | Gauge | Memory used in bytes |
| `dockhand_host_memory_total_bytes` | Gauge | Total memory in bytes |
| `dockhand_containers_total` | Gauge | Total containers |
| `dockhand_containers_running` | Gauge | Running containers |
| `dockhand_containers_stopped` | Gauge | Stopped containers |
| `dockhand_containers_paused` | Gauge | Paused containers |
| `dockhand_containers_restarting` | Gauge | Restarting containers |
| `dockhand_containers_unhealthy` | Gauge | Unhealthy containers |
| `dockhand_containers_pending_updates` | Gauge | Containers with a pending image update |
| `dockhand_containers_size_bytes` | Gauge | Writable layer disk usage |
| `dockhand_images_total` | Gauge | Total images |
| `dockhand_images_size_bytes` | Gauge | Total image disk usage |
| `dockhand_volumes_total` | Gauge | Total volumes |
| `dockhand_volumes_size_bytes` | Gauge | Total volume disk usage |
| `dockhand_networks_total` | Gauge | Total networks |
| `dockhand_stacks_total` | Gauge | Total stacks |
| `dockhand_stacks_running` | Gauge | Running stacks |
| `dockhand_stacks_partial` | Gauge | Partially running stacks |
| `dockhand_stacks_stopped` | Gauge | Stopped stacks |
| `dockhand_build_cache_size_bytes` | Gauge | Build cache disk usage |
| `dockhand_events_total` | Counter | Total Docker events recorded |
| `dockhand_events_today` | Gauge | Docker events recorded today |

## Run with Docker

```bash
docker build -t dockhand_exporter .

docker run -d \
  -e DOCKHAND_URL=http://your-dockhand-host:3002 \
  -e DOCKHAND_TOKEN=dh_xxxxxxxxxxxx \
  -p 9090:9090 \
  dockhand_exporter
```

## Run locally

```bash
export DOCKHAND_URL=http://your-dockhand-host:3002
export DOCKHAND_TOKEN=dh_xxxxxxxxxxxx
go run .
```