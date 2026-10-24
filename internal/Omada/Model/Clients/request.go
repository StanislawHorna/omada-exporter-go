package Clients

import (
	"omada_exporter_go/internal/Log"
	"omada_exporter_go/internal/Omada/HttpClient/ApiClient"
)

// Get returns the site-wide client list. It is only called when per-client
// metrics are enabled (see Prometheus.CollectMetrics), because every call costs
// one request per page on the controller.
func Get() (*[]Client, error) {
	Log.Debug("Fetching clients data")
	client := ApiClient.GetInstance()

	result, err := ApiClient.Get[Client](client, path_OpenApiClientsList, nil, nil, true)
	if err != nil {
		return nil, Log.Error(err, "Failed to get clients data")
	}
	return result, nil
}
