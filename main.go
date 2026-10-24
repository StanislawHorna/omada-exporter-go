package main

import (
	"net"
	"net/http"

	"omada_exporter_go/internal"
	"omada_exporter_go/internal/Log"
	"omada_exporter_go/internal/Prometheus"
)

func main() {
	Log.Init()
	Log.Info("Starting %s", internal.AppName)

	metrics := internal.GetConfig().Prometheus
	address := net.JoinHostPort(metrics.MetricsHost, metrics.MetricsPort)

	http.Handle(metrics.MetricsPath, Prometheus.OmadaMetricsHandler())
	Log.Info("Serving metrics on %s%s", address, metrics.MetricsPath)
	if err := http.ListenAndServe(address, nil); err != nil {
		panic(err)
	}
}
