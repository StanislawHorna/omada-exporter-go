package Clients

// path_OpenApiClientsList is the site-wide client list. One row per client that
// the controller currently associates with the site; the list only contains
// clients that are (or were very recently) connected, so it is fetched with
// pagination and replaces the previous client series on every scrape.
const path_OpenApiClientsList = "/openapi/v1/{omadaID}/sites/{siteID}/clients"

// Client is one row of the site-wide client list.
//
// All values are kept exactly as the controller reports them: the raw unit is
// documented per field and scaling belongs to the Prometheus layer (as with
// Devices.LastSeen), so the model stays a faithful representation of the API.
//
// The wireless-only fields at the bottom are omitted by the controller for
// wired clients and therefore decode to their zero values. They are only
// exposed when Wireless is true, so a missing value never surfaces as a
// measurement. Pointers are not needed because no wireless value is legitimately
// zero in a way that would have to be told apart from "absent": rssi 0 dBm is
// not a usable signal and the metric layer gates all of them on Wireless.
type Client struct {
	MacAddress     string   `json:"mac"`
	Name           string   `json:"name"`
	HostName       string   `json:"hostName"`
	Vendor         string   `json:"vendor"`
	DeviceType     string   `json:"deviceType"`
	DeviceCategory string   `json:"deviceCategory"`
	OSName         string   `json:"osName"`
	Model          string   `json:"model"`
	IP             string   `json:"ip"`
	IPv6List       []string `json:"ipv6List"`
	ConnectType    int      `json:"connectType"`
	ConnectDevType string   `json:"connectDevType"`
	Wireless       bool     `json:"wireless"`
	// HealthScore is -1 when the controller has no score for the client, which
	// is the normal case for wired clients.
	HealthScore int    `json:"healthScore"`
	APName      string `json:"apName"`
	APMac       string `json:"apMac"`
	Vid         int    `json:"vid"`
	Port        int    `json:"port"`
	PortName    string `json:"portName"`
	Activity    int    `json:"activity"`

	TrafficDown int64 `json:"trafficDown"` // bytes
	TrafficUp   int64 `json:"trafficUp"`   // bytes
	DownPacket  int64 `json:"downPacket"`
	UpPacket    int64 `json:"upPacket"`

	Uptime int64 `json:"uptime"` // seconds
	// LastSeen is a millisecond epoch, see Prometheus/Utils.UnixSeconds.
	LastSeen float64 `json:"lastSeen"`

	AuthStatus int  `json:"authStatus"`
	Blocked    bool `json:"blocked"`
	Guest      bool `json:"guest"`
	Active     bool `json:"active"`
	Manager    bool `json:"manager"`

	// Wireless-only fields, see the struct comment.
	SSID        string `json:"ssid"`
	RSSI        int    `json:"rssi"`        // dBm
	SNR         int    `json:"snr"`         // dB
	SignalLevel int    `json:"signalLevel"` // percent
	SignalRank  int    `json:"signalRank"`
	Channel     int    `json:"channel"`
	RadioID     int    `json:"radioId"`
	TxRate      int64  `json:"txRate"` // bits per second
	RxRate      int64  `json:"rxRate"` // bits per second
	PowerSave   bool   `json:"powerSave"`
	WiFiMode    int    `json:"wifiMode"`
}
