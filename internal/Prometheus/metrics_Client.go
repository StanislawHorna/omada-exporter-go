package Prometheus

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"omada_exporter_go/internal/Omada/Model/Clients"
	"omada_exporter_go/internal/Prometheus/Utils"
)

const (
	label_clientMac      string = "mac"
	label_clientName     string = "name"
	label_clientHostName string = "hostName"
	label_clientIP       string = "ip"
	label_clientType     string = "deviceType"
	label_clientVendor   string = "vendor"
	label_clientOSName   string = "osName"
	label_clientDevType  string = "connectDevType"
	label_clientAPName   string = "apName"
	label_clientAPMac    string = "apMac"
	label_clientSSID     string = "ssid"
	label_clientPortName string = "portName"
	label_clientGuest    string = "guest"
	label_clientBlocked  string = "blocked"
	label_clientManager  string = "manager"

	// clientHealthUnknown is the health score the controller reports when it
	// has none for the client, which is the normal case for wired clients.
	clientHealthUnknown int = -1
)

// clientIdentityLabels keys every numeric client series. Only the MAC address
// and the display name are used: the descriptive fields change often (host
// name, IP, SSID, AP, ...) and every distinct label set creates a new series,
// so they are carried by client_info only.
var clientIdentityLabels = []string{label_clientMac, label_clientName}

var clientInfoLabels = []string{
	label_clientMac,
	label_clientName,
	label_clientHostName,
	label_clientIP,
	label_clientType,
	label_clientVendor,
	label_clientOSName,
	label_clientDevType,
	label_clientAPName,
	label_clientAPMac,
	label_clientSSID,
	label_clientPortName,
	label_clientGuest,
	label_clientBlocked,
	label_clientManager,
}

var (
	client_info = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "client_info",
			Help: "Information about the client, always 1",
		},
		clientInfoLabels,
	)
	client_uptime_seconds = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "client_uptime_seconds",
			Help: "How long the client has been connected in seconds",
		},
		clientIdentityLabels,
	)
	client_last_seen = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "client_last_seen",
			Help: "The last time the client was seen, as a Unix timestamp in seconds",
		},
		clientIdentityLabels,
	)
	client_traffic_down_bytes_total = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "client_traffic_down_bytes_total",
			Help: "Total number of bytes downloaded by the client",
		},
		clientIdentityLabels,
	)
	client_traffic_up_bytes_total = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "client_traffic_up_bytes_total",
			Help: "Total number of bytes uploaded by the client",
		},
		clientIdentityLabels,
	)
	client_packets_down_total = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "client_packets_down_total",
			Help: "Total number of packets received by the client",
		},
		clientIdentityLabels,
	)
	client_packets_up_total = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "client_packets_up_total",
			Help: "Total number of packets sent by the client",
		},
		clientIdentityLabels,
	)
	client_signal_dbm = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "client_signal_dbm",
			Help: "The client signal strength in dBm (wireless clients only)",
		},
		clientIdentityLabels,
	)
	client_snr_db = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "client_snr_db",
			Help: "The client signal-to-noise ratio in dB (wireless clients only)",
		},
		clientIdentityLabels,
	)
	client_signal_level_percent = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "client_signal_level_percent",
			Help: "The client signal level in percent (0 - 100, wireless clients only)",
		},
		clientIdentityLabels,
	)
	client_tx_rate_bps = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "client_tx_rate_bps",
			Help: "The client transmit rate in bits per second (wireless clients only)",
		},
		clientIdentityLabels,
	)
	client_rx_rate_bps = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "client_rx_rate_bps",
			Help: "The client receive rate in bits per second (wireless clients only)",
		},
		clientIdentityLabels,
	)
	client_health_score = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "client_health_score",
			Help: "The client health score reported by the controller (0 - 100, omitted when the controller reports -1)",
		},
		clientIdentityLabels,
	)
)

// clientMetricVecs lists every per-client vector for the reset in
// ExposeClientMetrics. Keeping the list next to the vectors makes a new series
// impossible to forget in the reset.
var clientMetricVecs = []*prometheus.GaugeVec{
	client_info,
	client_uptime_seconds,
	client_last_seen,
	client_traffic_down_bytes_total,
	client_traffic_up_bytes_total,
	client_packets_down_total,
	client_packets_up_total,
	client_signal_dbm,
	client_snr_db,
	client_signal_level_percent,
	client_tx_rate_bps,
	client_rx_rate_bps,
	client_health_score,
}

// ExposeClientMetrics exposes the client list of the current scrape.
//
// Unlike the device, port and radio metrics, the client vectors are reset once
// per scrape before they are set. Devices are a stable inventory whose series
// are updated in place, but clients associate and disconnect constantly: a
// client that left the site has to disappear from /metrics instead of keeping
// the values of its last scrape forever (or until the exporter restarts).
// Resetting wholesale is simpler and safer than tracking which MACs vanished,
// and it also drops series whose descriptive labels changed.
func ExposeClientMetrics(clients []Clients.Client) {
	for _, vec := range clientMetricVecs {
		vec.DeletePartialMatch(prometheus.Labels{})
	}

	for i := range clients {
		exposeClientMetrics(&clients[i])
	}
}

func exposeClientMetrics(client *Clients.Client) {
	labels := getClientIdentityLabels(client)

	client_uptime_seconds.With(labels).Set(float64(client.Uptime))
	// The controller reports lastSeen in milliseconds; the metric is in Unix
	// seconds (see Utils.UnixSeconds for why the scale is detected).
	client_last_seen.With(labels).Set(Utils.UnixSeconds(client.LastSeen))

	client_traffic_down_bytes_total.With(labels).Set(float64(client.TrafficDown))
	client_traffic_up_bytes_total.With(labels).Set(float64(client.TrafficUp))
	client_packets_down_total.With(labels).Set(float64(client.DownPacket))
	client_packets_up_total.With(labels).Set(float64(client.UpPacket))

	if client.HealthScore != clientHealthUnknown {
		client_health_score.With(labels).Set(float64(client.HealthScore))
	}

	// The controller omits the wireless fields for wired clients (they decode
	// to zero values, see Clients.Client), so they are only reported for
	// wireless clients instead of surfacing a fabricated 0 dBm.
	if client.Wireless {
		client_signal_dbm.With(labels).Set(float64(client.RSSI))
		client_snr_db.With(labels).Set(float64(client.SNR))
		client_signal_level_percent.With(labels).Set(float64(client.SignalLevel))
		client_tx_rate_bps.With(labels).Set(float64(client.TxRate))
		client_rx_rate_bps.With(labels).Set(float64(client.RxRate))
	}

	client_info.With(getClientInfoLabels(client)).Set(1)
}

func getClientIdentityLabels(client *Clients.Client) prometheus.Labels {
	return prometheus.Labels{
		label_clientMac:  client.MacAddress,
		label_clientName: client.Name,
	}
}

func getClientInfoLabels(client *Clients.Client) prometheus.Labels {
	labels := getClientIdentityLabels(client)
	labels[label_clientHostName] = client.HostName
	labels[label_clientIP] = client.IP
	labels[label_clientType] = client.DeviceType
	labels[label_clientVendor] = client.Vendor
	labels[label_clientOSName] = client.OSName
	labels[label_clientDevType] = client.ConnectDevType
	labels[label_clientAPName] = client.APName
	labels[label_clientAPMac] = client.APMac
	labels[label_clientSSID] = client.SSID
	labels[label_clientPortName] = client.PortName
	labels[label_clientGuest] = strconv.FormatBool(client.Guest)
	labels[label_clientBlocked] = strconv.FormatBool(client.Blocked)
	labels[label_clientManager] = strconv.FormatBool(client.Manager)
	return labels
}
