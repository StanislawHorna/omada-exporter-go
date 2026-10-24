package Devices

import (
	"encoding/json"
	"testing"

	"omada_exporter_go/internal/Omada/Enum"
)

// The payloads below are copied from a live controller's
// GET /openapi/v1/{omadacId}/sites/{siteId}/devices response. Structure, field
// names and the lastSeen scale are untouched; MAC addresses, names, IPs,
// serial numbers and firmware strings are replaced with documentation values
// (RFC 5737 / locally administered MACs).
const deviceListFixture = `{
  "currentPage": 1,
  "currentSize": 1000,
  "totalRows": 2,
  "data": [
    {
      "compatible": 0,
      "cpuUtil": 1,
      "detailStatus": 14,
      "deviceSeriesType": 0,
      "firmwareVersion": "3.3.1 Build 20251203 Rel. 58135",
      "ip": "192.0.2.36",
      "lastSeen": 1792833739702,
      "mac": "AA-BB-CC-00-00-01",
      "memUtil": 53,
      "model": "EAP235-Wall(EU) v1.0",
      "modelName": "EAP235-Wall",
      "modelVersion": "1.0",
      "name": "AP Office",
      "publicIp": "192.0.2.36",
      "sn": "0000000000000",
      "status": 1,
      "supportAfc": false,
      "switchConsistent": true,
      "type": "ap",
      "uptime": "1day(s) 21h 42m 8s"
    },
    {
      "compatible": 0,
      "cpuUtil": 12,
      "detailStatus": 14,
      "deviceSeriesType": 0,
      "firmwareVersion": "1.0.8 Build 20260718 Rel.11174",
      "ip": "192.0.2.122",
      "lastSeen": 1792833715370,
      "mac": "AA-BB-CC-00-00-02",
      "memUtil": 66,
      "model": "ES210GMP v1.0",
      "modelName": "ES210GMP",
      "modelVersion": "1.0",
      "name": "Switch Basement",
      "publicIp": "192.0.2.122",
      "sn": "0000000000001",
      "status": 1,
      "subtype": "es",
      "switchConsistent": true,
      "type": "switch",
      "uptime": "1day(s) 21h 41m 33s"
    }
  ]
}`

// TestDeviceListParsing locks the parsing of the endpoint that is the only
// source of lastSeen: the per-device Open API endpoints (AP, switch, gateway
// detail) do not return the field at all, it is copied from this list.
func TestDeviceListParsing(t *testing.T) {
	var grid struct {
		Data []Device `json:"data"`
	}
	if err := json.Unmarshal([]byte(deviceListFixture), &grid); err != nil {
		t.Fatalf("cannot unmarshal device list fixture: %v", err)
	}
	if len(grid.Data) != 2 {
		t.Fatalf("got %d devices, want 2", len(grid.Data))
	}

	ap := grid.Data[0]
	if ap.Type != Enum.DeviceType_AccessPoint {
		t.Errorf("AP type = %q, want %q", ap.Type, Enum.DeviceType_AccessPoint)
	}
	if ap.MacAddress != "AA-BB-CC-00-00-01" {
		t.Errorf("AP mac = %q", ap.MacAddress)
	}
	if ap.Name != "AP Office" {
		t.Errorf("AP name = %q", ap.Name)
	}
	if ap.Model != "EAP235-Wall(EU) v1.0" {
		t.Errorf("AP model = %q", ap.Model)
	}
	if ap.IP != "192.0.2.36" {
		t.Errorf("AP ip = %q", ap.IP)
	}
	if ap.CpuUsage != 1 || ap.RamUsage != 53 {
		t.Errorf("AP cpu/ram = %d/%d, want 1/53", ap.CpuUsage, ap.RamUsage)
	}
	// The controller reports lastSeen as a millisecond epoch. Keeping the raw
	// value in the model is intentional: scaling belongs to the Prometheus
	// layer, so the models stay a faithful representation of the API.
	if ap.LastSeen != 1792833739702 {
		t.Errorf("AP lastSeen = %v, want the raw millisecond value 1792833739702", ap.LastSeen)
	}

	sw := grid.Data[1]
	if sw.Type != Enum.DeviceType_Switch {
		t.Errorf("switch type = %q, want %q", sw.Type, Enum.DeviceType_Switch)
	}
	// "es" marks an Agile Series switch, which needs the /switches/es/{mac}
	// endpoint (see the first local patch).
	if sw.Subtype != "es" {
		t.Errorf("switch subtype = %q, want %q", sw.Subtype, "es")
	}
	if sw.LastSeen != 1792833715370 {
		t.Errorf("switch lastSeen = %v, want the raw millisecond value 1792833715370", sw.LastSeen)
	}
}
