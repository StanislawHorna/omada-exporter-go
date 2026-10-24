# Omada Exporter

Prometheus Exporter written in Go for TP-Link Omada SDN.

This exporter is built with native Go HTTP and [prometheus/client_golang](https://github.com/prometheus/client_golang)
libraries to expose metrics from Omada SDN to Prometheus.
Metrics are queried live from the controller when the `/metrics` endpoint is accessed,
avoiding unnecessary polling when not in use.

Supports both Omada OpenAPI and Web API to accommodate current platform limitations.

## Grafana Dashboards

### [Site Overview](/Dashboards/Site_Overview.json)

![image](/Pictures/Site_Overview.png)

### [Router](/Dashboards/Router.json)

![image](/Pictures/Router_1.png)

### [Switch](/Dashboards/Switch.json)

![image](/Pictures/Switch_1.png)

### [Access Point](/Dashboards/AccessPoint.json)

![image](/Pictures/AccessPoint_1.png)

## 🚀 Getting Started

### Prerequisites

- **Omada Controller** with configured:
  - [**OpenAPI Client**](#omada-authentication-setup)
  - [**Service user account**](#omada-authentication-setup)
- **Prometheus** (or another metrics scraper) to collect and store the metrics
- **Grafana** to visualize the data using built-in dashboards
- (Optional) **Grafana Loki** if you want to collect and view logs from the exporter

### Setup

1. Clone the Repository

```shell
git clone https://github.com/StanislawHorna/Omada_Exporter_Go.git
cd Omada_Exporter_Go
```

2. Create `.env` File with variables from [Configuration Parameters](#configuration-parameters)

```
LOG_LEVEL=info

OMADA_URL=https://<omada_ip_address>:<omada_port>
OMADA_SITE_NAME=<your_omada_site_name>
OMADA_CLIENT_ID=<open_api_client_id>
OMADA_CLIENT_SECRET=<open_api_client_secret>
OMADA_USERNAME=<web_ui_username>
OMADA_PASSWORD=<web_ui_password>

LOKI_URL=http://<loki_ip_address>:<loki_port>
LOKI_ENV=prod
LOKI_APP_VERSION=1.0.0
```

3. Build the Docker image

```shell
docker build -t omada_exporter .
```

4. Run the container

```shell
docker run -d \
  --env-file .env \
  -p 8080:8080 \
  --name omada_exporter \
  omada_exporter
```

5. Configure Prometheus to scrape the data:

```YAML
scrape_configs:
  - job_name: "Omada"
    static_configs:
      - targets: ["<docker_host_ip>:8080"]
        labels:
          device_name: "<omada_controller_friendly_name>"
    metrics_path: /metrics
```

6. Import dashboards from [`Dashboards`](/Dashboards/) directory to your Grafana instance.

> [!TIP]
> After starting up the exporter navigate to `http://"<docker_host_ip>:8080/metrics`
> and check `omada_http_client_requests_total` metrics.
> There should be **no** metrics with code different different than 200

### Omada Authentication Setup

- OpenAPI Client – Created via: `Settings -> Platform Integration`.
  Assign admin role for full API access.
- Service User – Create under: `Account section` at `Global level`.
  Assign viewer role for read-only access.

### Configuration Parameters

All configuration values are read from environment variables. These can be provided via a .env file, environment variables in your system, or directly in Docker Compose.

| Variable                | Description                                                                                            | Default    |
| ----------------------- | ------------------------------------------------------------------------------------------------------ | ---------- |
| `LOG_LEVEL`             | Logging verbosity level (_debug_, _info_, _warn_, _error_)                                             | `error`    |
| `OMADA_URL`             | Full base URL to the Omada controller (e.g., https://192.168.1.1:8043)                                 | -          |
| `OMADA_SITE_NAME`       | Name of the site you wish to monitor                                                                   | -          |
| `OMADA_CLIENT_ID`       | OpenAPI Client ID created in Omada Platform Integration                                                | -          |
| `OMADA_CLIENT_SECRET`   | OpenAPI Client Secret                                                                                  | -          |
| `OMADA_USERNAME`        | Username for Web API access                                                                            | -          |
| `OMADA_PASSWORD`        | Password for Web API access                                                                            | -          |
| `METRICS_PATH`          | HTTP path the metrics are served on                                                                    | `/metrics` |
| `METRICS_HOST`          | Address the metrics server binds to (empty = all interfaces)                                           | -          |
| `METRICS_PORT`          | Port the metrics server binds to                                                                       | `8080`     |
| `ENABLE_CLIENT_METRICS` | Expose per-client statistics: one extra API request per scrape and one series set per connected client | `false`    |

## Tested on devices:

- **Switch (regular)**: [SG2218 v1.20](https://www.tp-link.com/en/business-networking/omada-switch-smart/sg2218/)
- **Switch (Agile Series / Easy Managed, `ES…`)**: [ES210GMP v1.0](https://www.omadanetworks.com/en/business-networking/omada-switch-agile/es210gmp/) (controller 6.3.0.45)
- **Router**: [ER707-M2 v1.0](https://www.tp-link.com/en/business-networking/omada-router-wired-router/er707-m2/v1/)
- **Access Point**: [EAP650(EU) v1.0](https://www.tp-link.com/en/business-networking/omada-wifi-ceiling-mount/eap650/v1/)
- **Omada controller**: [Software Controller v5.15.8.s hosted in Docker container](https://github.com/mbentley/docker-omada-controller)
