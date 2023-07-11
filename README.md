
# CleverCloud exporter for Prometheus

## Setup

Get credentials:

```sh
npm install -g clever-tools
clever login
```

## Run

```
export CLEVER_TOKEN=xxxx
export CLEVER_SECRET=yyyy
```

Locally:

```sh
go build -o clevercloud-exporter *.go
./clevercloud-exporter
```

In Docker:

```sh
docker build -t samber/clevercloud-exporter:0.1.0 .
docker push samber/clevercloud-exporter:0.1.0

docker run -d -p 9217:9217 samber/clevercloud-exporter
```

## Metrics

```txt
clevercloud_application_status{org_id, org_name, app_id, app_name, region, archived, type}

clevercloud_addon_status{org_id, org_name, addon_id, addon_name, region}

clevercloud_instance_status{org_id, org_name, app_id, app_name, region, state, type}
clevercloud_instance_memory_total{org_id, org_name, app_id, app_name, region, type}
clevercloud_instance_cpu_total{org_id, org_name, app_id, app_name, region, type}
clevercloud_instance_hourly_price_total{org_id, org_name, app_id, app_name, region, type}
```
