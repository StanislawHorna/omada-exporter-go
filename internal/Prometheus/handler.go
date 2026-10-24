package Prometheus

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"omada_exporter_go/internal/Log"
)

func OmadaMetricsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Partial failures still expose the metrics that were collected, so the
		// scrape must not fail; without this log the failure was invisible.
		if err := CollectMetrics(); err != nil {
			Log.Error(err, "Omada metrics collection reported errors")
		}
		promhttp.Handler().ServeHTTP(w, r)
	})
}
