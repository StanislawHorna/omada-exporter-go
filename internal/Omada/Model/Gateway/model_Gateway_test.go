package Gateway

import (
	"encoding/json"
	"testing"
)

// Sanitised copy of a GatewayInfo payload as returned by
// GET /openapi/v1/{omadacId}/sites/{siteId}/gateways/{gatewayMac}. Values are
// documentation placeholders; the shape and the lastSeen scale are real.
const gatewayFixture = `{
  "deviceType": "gateway",
  "name": "Gateway",
  "mac": "AA-BB-CC-00-00-03",
  "showModel": "ER707-M2 v1.0",
  "firmwareVersion": "1.2.3 Build 20240822 Rel.52946",
  "ip": "192.0.2.1",
  "uptime": "10day(s) 01h 02m 03s",
  "temp": 51,
  "cpuUtil": 5,
  "memUtil": 13,
  "ipv6List": ["2001:db8::1"],
  "lastSeen": 1792833700123,
  "portConfigs": []
}`

func TestGatewayLastSeenParsing(t *testing.T) {
	var gw Gateway
	if err := json.Unmarshal([]byte(gatewayFixture), &gw); err != nil {
		t.Fatalf("cannot unmarshal gateway fixture: %v", err)
	}

	if gw.MacAddress != "AA-BB-CC-00-00-03" {
		t.Errorf("mac = %q", gw.MacAddress)
	}
	if gw.GetLastSeen() != 1792833700123 {
		t.Errorf("lastSeen = %v, want the raw millisecond value 1792833700123", gw.GetLastSeen())
	}
}
