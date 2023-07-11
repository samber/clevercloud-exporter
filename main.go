package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/samber/clevercloud-exporter/collector"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func getEnv(key string, defaultVal string) string {
	if envVal, ok := os.LookupEnv(key); ok {
		return envVal
	}

	return defaultVal
}

var (
	listenAddr        = flag.String("web.listen-address", getEnv("WEB_LISTEN_ADDRESS", "0.0.0.0:9217"), "Address to listen on for web interface and telemetry, defaults to 0.0.0.0:9217")
	metricsPath       = flag.String("web.telemetry-path", getEnv("WEB_TELEMETRY_PATH", "/metrics"), "A path under which to expose metrics.")
	metricsNamespace  = flag.String("namespace", getEnv("NAMESPACE", "clevercloud"), "Prometheus metrics namespace, as the prefix of metrics name")
	clevercloudToken  = flag.String("cc.token", getEnv("CLEVER_TOKEN", "xxxx"), "Clever Cloud token")
	clevercloudSecret = flag.String("cc.secret", getEnv("CLEVER_SECRET", "xxxx"), "Clever Cloud secret")
)

func main() {
	flag.Parse()

	clevercloudClient := NewClient(*clevercloudToken, *clevercloudSecret)

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector.NewDeploymentCollector(*metricsNamespace, clevercloudClient))

	http.Handle(*metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>A Prometheus Exporter</title></head>
			<body>
			<h1>A Prometheus Exporter</h1>
			<p><a href='/metrics'>Metrics</a></p>
			</body>
			</html>`))
	})

	log.Printf("Starting Server at %s%s", *listenAddr, *metricsPath)
	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
