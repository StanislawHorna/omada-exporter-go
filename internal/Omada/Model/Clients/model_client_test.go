package Clients

import (
	"encoding/json"
	"testing"
)

// The payload below is copied from a live controller's
// GET /openapi/v1/{omadacId}/sites/{siteId}/clients response. Structure, field
// names, units and the measured values are untouched; MAC addresses, names,
// IPs, SSIDs and vendor/model strings are replaced with documentation values
// (RFC 5737 / locally administered MACs).
//
// Row 0 is a wireless client: it carries the wireless-only fields (ssid, rssi,
// snr, signalLevel, signalRank, channel, radioId, txRate, rxRate, powerSave,
// wifiMode). Row 1 is a wired client: the controller omits all of them and adds
// port/portName instead. Both report healthScore -1 because the controller has
// no score for them.
const clientListFixture = `{
  "currentPage": 1,
  "currentSize": 1000,
  "totalRows": 2,
  "data": [
    {
      "active": true,
      "activity": 2,
      "apMac": "AA-BB-CC-00-00-20",
      "apName": "AP Office",
      "authStatus": 1,
      "blocked": false,
      "channel": 40,
      "connectDevType": "ap",
      "connectType": 1,
      "deviceCategory": "phone",
      "deviceType": "phone",
      "downPacket": 12100447,
      "guest": false,
      "healthScore": -1,
      "hostName": "phone-01",
      "ip": "192.0.2.31",
      "ipv6List": ["2001:db8::10"],
      "lastSeen": 1792840148946,
      "mac": "AA-BB-CC-00-00-31",
      "manager": false,
      "model": "Test Phone v1.0",
      "name": "Phone Living Room",
      "osName": "Android",
      "powerSave": true,
      "radioId": 1,
      "rssi": -64,
      "rxRate": 468000,
      "signalLevel": 65,
      "signalRank": 4,
      "snr": 31,
      "ssid": "Home",
      "trafficDown": 4277272606,
      "trafficUp": 4068719423,
      "txRate": 780000,
      "upPacket": 12426241,
      "uptime": 170886,
      "vendor": "Test Vendor",
      "vid": 0,
      "wifiMode": 5,
      "wireless": true
    },
    {
      "active": true,
      "activity": 0,
      "apMac": "",
      "apName": "",
      "authStatus": 1,
      "blocked": false,
      "connectDevType": "switch",
      "connectType": 0,
      "deviceCategory": "laptop",
      "deviceType": "laptop",
      "downPacket": 81,
      "guest": true,
      "healthScore": -1,
      "hostName": "laptop-01",
      "ip": "192.0.2.32",
      "ipv6List": [],
      "lastSeen": 1792840139834,
      "mac": "AA-BB-CC-00-00-32",
      "manager": false,
      "model": "Test Laptop v1.0",
      "name": "Laptop Office",
      "osName": "Linux",
      "port": 2,
      "portName": "Port 2",
      "trafficDown": 8774,
      "trafficUp": 90464,
      "upPacket": 85,
      "uptime": 1031,
      "vendor": "Test Vendor",
      "vid": 0,
      "wireless": false
    }
  ]
}`

// TestClientListParsing locks the parsing of the client list for both client
// kinds: the wireless-only fields must stay absent for wired clients and the
// raw units (bytes, bps, seconds, milliseconds) must survive into the model,
// because the metric layer converts them.
func TestClientListParsing(t *testing.T) {
	var grid struct {
		Data []Client `json:"data"`
	}
	if err := json.Unmarshal([]byte(clientListFixture), &grid); err != nil {
		t.Fatalf("cannot unmarshal client list fixture: %v", err)
	}
	if len(grid.Data) != 2 {
		t.Fatalf("got %d clients, want 2", len(grid.Data))
	}

	wireless := grid.Data[0]
	if !wireless.Wireless {
		t.Error("wireless client: Wireless = false, want true")
	}
	if wireless.MacAddress != "AA-BB-CC-00-00-31" || wireless.Name != "Phone Living Room" {
		t.Errorf("wireless identity = %q/%q", wireless.MacAddress, wireless.Name)
	}
	if wireless.HostName != "phone-01" || wireless.IP != "192.0.2.31" || wireless.OSName != "Android" {
		t.Errorf("wireless descriptive fields = %q/%q/%q", wireless.HostName, wireless.IP, wireless.OSName)
	}
	if len(wireless.IPv6List) != 1 || wireless.IPv6List[0] != "2001:db8::10" {
		t.Errorf("wireless ipv6List = %v", wireless.IPv6List)
	}
	if wireless.APName != "AP Office" || wireless.APMac != "AA-BB-CC-00-00-20" {
		t.Errorf("wireless AP = %q/%q", wireless.APName, wireless.APMac)
	}
	// Units as reported by the controller (see the metric layer for the
	// conversion of lastSeen).
	if wireless.Uptime != 170886 {
		t.Errorf("wireless uptime = %d, want 170886 seconds", wireless.Uptime)
	}
	if wireless.LastSeen != 1792840148946 {
		t.Errorf("wireless lastSeen = %v, want the raw millisecond value 1792840148946", wireless.LastSeen)
	}
	if wireless.TrafficDown != 4277272606 || wireless.TrafficUp != 4068719423 {
		t.Errorf("wireless traffic = %d/%d, want the raw byte counts", wireless.TrafficDown, wireless.TrafficUp)
	}
	if wireless.DownPacket != 12100447 || wireless.UpPacket != 12426241 {
		t.Errorf("wireless packets = %d/%d", wireless.DownPacket, wireless.UpPacket)
	}
	if wireless.SSID != "Home" || wireless.RSSI != -64 || wireless.SNR != 31 || wireless.SignalLevel != 65 {
		t.Errorf("wireless signal = %q/%d/%d/%d", wireless.SSID, wireless.RSSI, wireless.SNR, wireless.SignalLevel)
	}
	if wireless.TxRate != 780000 || wireless.RxRate != 468000 {
		t.Errorf("wireless rates = %d/%d, want the raw bps values", wireless.TxRate, wireless.RxRate)
	}
	if wireless.Channel != 40 || wireless.RadioID != 1 || wireless.WiFiMode != 5 || !wireless.PowerSave {
		t.Errorf("wireless radio = channel %d, radioId %d, wifiMode %d, powerSave %t",
			wireless.Channel, wireless.RadioID, wireless.WiFiMode, wireless.PowerSave)
	}
	if wireless.Guest || wireless.Blocked || wireless.Manager || !wireless.Active {
		t.Errorf("wireless flags = guest %t, blocked %t, manager %t, active %t",
			wireless.Guest, wireless.Blocked, wireless.Manager, wireless.Active)
	}

	wired := grid.Data[1]
	if wired.Wireless {
		t.Error("wired client: Wireless = true, want false")
	}
	if wired.MacAddress != "AA-BB-CC-00-00-32" || wired.Port != 2 || wired.PortName != "Port 2" {
		t.Errorf("wired identity = %q, port %d/%q", wired.MacAddress, wired.Port, wired.PortName)
	}
	if !wired.Guest {
		t.Error("wired client: Guest = false, want true")
	}
	// The controller omits the wireless fields for wired clients; they have to
	// stay zero so the metric layer can gate them on Wireless.
	if wired.RSSI != 0 || wired.SNR != 0 || wired.SignalLevel != 0 || wired.TxRate != 0 || wired.RxRate != 0 || wired.SSID != "" {
		t.Errorf("wired client has wireless values: rssi %d, snr %d, signalLevel %d, txRate %d, rxRate %d, ssid %q",
			wired.RSSI, wired.SNR, wired.SignalLevel, wired.TxRate, wired.RxRate, wired.SSID)
	}
	if wired.LastSeen != 1792840139834 {
		t.Errorf("wired lastSeen = %v, want the raw millisecond value 1792840139834", wired.LastSeen)
	}
}
