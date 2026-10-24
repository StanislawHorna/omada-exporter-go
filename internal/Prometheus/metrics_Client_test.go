package Prometheus

import (
	"math"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"

	"omada_exporter_go/internal/Omada/Model/Clients"
)

// wirelessClient and wiredClient are sanitised copies of live client rows (the
// same values as the fixture in Clients/model_client_test.go), except for the
// health score: the live rows report -1, so 87 is used where a present score is
// what the test is about.
func wirelessClient() Clients.Client {
	return Clients.Client{
		MacAddress:     "AA-BB-CC-00-00-31",
		Name:           "Phone Living Room",
		HostName:       "phone-01",
		Vendor:         "Test Vendor",
		DeviceType:     "phone",
		DeviceCategory: "phone",
		OSName:         "Android",
		Model:          "Test Phone v1.0",
		IP:             "192.0.2.31",
		IPv6List:       []string{"2001:db8::10"},
		ConnectType:    1,
		ConnectDevType: "ap",
		Wireless:       true,
		HealthScore:    87,
		APName:         "AP Office",
		APMac:          "AA-BB-CC-00-00-20",
		Activity:       2,
		TrafficDown:    4277272606,
		TrafficUp:      4068719423,
		DownPacket:     12100447,
		UpPacket:       12426241,
		Uptime:         170886,
		LastSeen:       1792840148946,
		AuthStatus:     1,
		Active:         true,
		SSID:           "Home",
		RSSI:           -64,
		SNR:            31,
		SignalLevel:    65,
		SignalRank:     4,
		Channel:        40,
		RadioID:        1,
		TxRate:         780000,
		RxRate:         468000,
		PowerSave:      true,
		WiFiMode:       5,
	}
}

func wiredClient() Clients.Client {
	return Clients.Client{
		MacAddress:     "AA-BB-CC-00-00-32",
		Name:           "Laptop Office",
		HostName:       "laptop-01",
		Vendor:         "Test Vendor",
		DeviceType:     "laptop",
		DeviceCategory: "laptop",
		OSName:         "Linux",
		Model:          "Test Laptop v1.0",
		IP:             "192.0.2.32",
		ConnectDevType: "switch",
		HealthScore:    -1,
		Port:           2,
		PortName:       "Port 2",
		TrafficDown:    8774,
		TrafficUp:      90464,
		DownPacket:     81,
		UpPacket:       85,
		Uptime:         1031,
		LastSeen:       1792840139834,
		AuthStatus:     1,
		Guest:          true,
		Active:         true,
	}
}

// gatherClientSample returns the sample of the given client metric that carries
// exactly the given labels, or nil when the series does not exist.
func gatherClientSample(t *testing.T, metricName string, want map[string]string) *dto.Metric {
	t.Helper()

	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("cannot gather metrics: %v", err)
	}

	for _, family := range families {
		if family.GetName() != metricName {
			continue
		}
		for _, metric := range family.GetMetric() {
			labels := make(map[string]string, len(metric.GetLabel()))
			for _, label := range metric.GetLabel() {
				labels[label.GetName()] = label.GetValue()
			}
			if len(labels) != len(want) {
				continue
			}
			match := true
			for name, value := range want {
				if labels[name] != value {
					match = false
					break
				}
			}
			if match {
				return metric
			}
		}
	}

	return nil
}

func requireClientSample(t *testing.T, metricName string, want map[string]string) float64 {
	t.Helper()

	metric := gatherClientSample(t, metricName, want)
	if metric == nil {
		t.Fatalf("no %s sample with labels %v", metricName, want)
	}
	return metric.GetGauge().GetValue()
}

func requireNoClientSample(t *testing.T, metricName string, want map[string]string) {
	t.Helper()

	if metric := gatherClientSample(t, metricName, want); metric != nil {
		t.Fatalf("%s has a sample with labels %v, want no series", metricName, want)
	}
}

// client_last_seen is consumed by Grafana date/time units and by
// time() - client_last_seen, both of which expect Unix seconds. The controller
// reports milliseconds, so the exporter has to scale it down (as it already
// does for device_last_seen).
func TestExposeClientMetrics_LastSeenIsUnixSeconds(t *testing.T) {
	tests := []struct {
		name       string
		client     Clients.Client
		wantSecond float64
	}{
		{
			name:       "wireless client, milliseconds from the client list",
			client:     wirelessClient(),
			wantSecond: 1792840148.946,
		},
		{
			name:       "wired client, milliseconds from the client list",
			client:     wiredClient(),
			wantSecond: 1792840139.834,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ExposeClientMetrics([]Clients.Client{tt.client})

			got := requireClientSample(t, "client_last_seen", getClientIdentityLabels(&tt.client))
			if math.Abs(got-tt.wantSecond) > 1e-6 {
				t.Fatalf("client_last_seen = %v, want %v (Unix seconds)", got, tt.wantSecond)
			}
		})
	}
}

// TestExposeClientMetrics_WirelessClient covers the full series set of a
// wireless client: every numeric series keyed by mac/name, the wireless-only
// values, and client_info carrying the descriptive labels.
func TestExposeClientMetrics_WirelessClient(t *testing.T) {
	client := wirelessClient()
	ExposeClientMetrics([]Clients.Client{client})
	labels := getClientIdentityLabels(&client)

	values := map[string]float64{
		"client_uptime_seconds":           170886,
		"client_last_seen":                1792840148.946,
		"client_traffic_down_bytes_total": 4277272606,
		"client_traffic_up_bytes_total":   4068719423,
		"client_packets_down_total":       12100447,
		"client_packets_up_total":         12426241,
		"client_signal_dbm":               -64,
		"client_snr_db":                   31,
		"client_signal_level_percent":     65,
		"client_tx_rate_bps":              780000,
		"client_rx_rate_bps":              468000,
		"client_health_score":             87,
	}
	for metricName, want := range values {
		got := requireClientSample(t, metricName, labels)
		if math.Abs(got-want) > 1e-6 {
			t.Errorf("%s = %v, want %v", metricName, got, want)
		}
	}

	if got := requireClientSample(t, "client_info", getClientInfoLabels(&client)); got != 1 {
		t.Errorf("client_info = %v, want 1", got)
	}
}

// TestExposeClientMetrics_WiredClient covers the reduced series set: a wired
// client has no signal, SNR, signal level or rates at all, and the controller
// reports no health score for it, so neither must be exposed as a 0.
func TestExposeClientMetrics_WiredClient(t *testing.T) {
	client := wiredClient()
	ExposeClientMetrics([]Clients.Client{client})
	labels := getClientIdentityLabels(&client)

	values := map[string]float64{
		"client_uptime_seconds":           1031,
		"client_traffic_down_bytes_total": 8774,
		"client_traffic_up_bytes_total":   90464,
		"client_packets_down_total":       81,
		"client_packets_up_total":         85,
	}
	for metricName, want := range values {
		got := requireClientSample(t, metricName, labels)
		if math.Abs(got-want) > 1e-6 {
			t.Errorf("%s = %v, want %v", metricName, got, want)
		}
	}

	for _, metricName := range []string{
		"client_signal_dbm",
		"client_snr_db",
		"client_signal_level_percent",
		"client_tx_rate_bps",
		"client_rx_rate_bps",
		"client_health_score",
	} {
		requireNoClientSample(t, metricName, labels)
	}

	if got := requireClientSample(t, "client_info", getClientInfoLabels(&client)); got != 1 {
		t.Errorf("client_info = %v, want 1", got)
	}
}

// TestExposeClientMetrics_SkipsUnknownHealthScore locks the -1 handling for a
// wireless client as well: the controller uses -1 for "no score", not for a
// failing client.
func TestExposeClientMetrics_SkipsUnknownHealthScore(t *testing.T) {
	client := wirelessClient()
	client.HealthScore = -1

	ExposeClientMetrics([]Clients.Client{client})
	requireNoClientSample(t, "client_health_score", getClientIdentityLabels(&client))

	client.HealthScore = 42
	ExposeClientMetrics([]Clients.Client{client})
	if got := requireClientSample(t, "client_health_score", getClientIdentityLabels(&client)); got != 42 {
		t.Errorf("client_health_score = %v, want 42", got)
	}
}

// TestCollectClientMetrics_Gating locks the opt-in behaviour: with the flag
// disabled the clients endpoint is not requested at all (no extra load and no
// rate-limit budget on the controller) and no client series appear.
func TestCollectClientMetrics_Gating(t *testing.T) {
	originalGetClients := getClients
	t.Cleanup(func() { getClients = originalGetClients })

	client := wirelessClient()
	client.MacAddress = "AA-BB-CC-00-00-41"
	client.Name = "Gating Client"
	clients := []Clients.Client{client}

	calls := 0
	getClients = func() (*[]Clients.Client, error) {
		calls++
		return &clients, nil
	}

	collectClientMetrics(false)
	if calls != 0 {
		t.Fatalf("clients endpoint requested %d times while disabled, want 0", calls)
	}
	requireNoClientSample(t, "client_info", getClientInfoLabels(&client))
	requireNoClientSample(t, "client_uptime_seconds", getClientIdentityLabels(&client))

	collectClientMetrics(true)
	if calls != 1 {
		t.Fatalf("clients endpoint requested %d times while enabled, want 1", calls)
	}
	if got := requireClientSample(t, "client_info", getClientInfoLabels(&client)); got != 1 {
		t.Errorf("client_info = %v, want 1", got)
	}
}

// TestExposeClientMetrics_DropsStaleSeries locks the per-scrape reset: a client
// that disconnects (and therefore disappears from the list) must lose all of
// its series, and a client that switches from wireless to wired must lose its
// wireless-only series.
func TestExposeClientMetrics_DropsStaleSeries(t *testing.T) {
	gone := wirelessClient()
	gone.MacAddress = "AA-BB-CC-00-00-51"
	gone.Name = "Client A"

	remaining := wirelessClient()
	remaining.MacAddress = "AA-BB-CC-00-00-52"
	remaining.Name = "Client B"

	ExposeClientMetrics([]Clients.Client{gone})
	requireClientSample(t, "client_info", getClientInfoLabels(&gone))
	requireClientSample(t, "client_uptime_seconds", getClientIdentityLabels(&gone))

	ExposeClientMetrics([]Clients.Client{remaining})
	requireNoClientSample(t, "client_info", getClientInfoLabels(&gone))
	requireNoClientSample(t, "client_uptime_seconds", getClientIdentityLabels(&gone))
	requireNoClientSample(t, "client_signal_dbm", getClientIdentityLabels(&gone))
	requireClientSample(t, "client_info", getClientInfoLabels(&remaining))

	// The same MAC can also stay connected while moving from Wi-Fi to a cable.
	remaining.Wireless = false
	ExposeClientMetrics([]Clients.Client{remaining})
	requireClientSample(t, "client_uptime_seconds", getClientIdentityLabels(&remaining))
	requireNoClientSample(t, "client_signal_dbm", getClientIdentityLabels(&remaining))
	requireNoClientSample(t, "client_snr_db", getClientIdentityLabels(&remaining))
}
