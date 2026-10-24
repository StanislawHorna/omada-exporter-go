package Switch

import (
	"encoding/json"
	"math"
	"testing"

	"omada_exporter_go/internal/Omada/Enum"
)

// The payload below is copied from a live controller's
// GET /openapi/v1/{omadacId}/sites/{siteId}/switches/ports/poe-info response.
// Structure, field names and the measured power/voltage/current values are
// untouched; MAC addresses, switch and port names are replaced with
// documentation values (locally administered MACs).
//
// Port 1 is powered (7.9 W, 52.9 V, 150.0 mA), port 2 supports PoE but delivers
// nothing, and port 9 does not support PoE at all - the controller omits
// power, voltage and current for it entirely.
const poeInfoFixture = `{
  "currentPage": 1,
  "currentSize": 100,
  "totalRows": 3,
  "data": [
    {
      "connectedStatus": 1,
      "current": 150.0,
      "disable": false,
      "duplex": 0,
      "linkSpeed": 0,
      "operation": "switching",
      "pdClass": "--",
      "poe": 1,
      "poeDisplayType": 4,
      "poeStatus": 7.9,
      "port": 1,
      "portName": "Port 1",
      "portStatus": {
        "duplex": 0,
        "linkSpeed": 0,
        "linkStatus": 0,
        "poe": true,
        "poePower": 7.9,
        "port": 1,
        "rx": 0,
        "rxRate": 0,
        "stpDiscarding": false,
        "tx": 0,
        "txRate": 0
      },
      "power": 7.9,
      "switchMac": "AA-BB-CC-00-00-02",
      "switchName": "Switch Basement",
      "switchSupportPoe": 1,
      "supportPoe": true,
      "type": 1,
      "voltage": 52.9
    },
    {
      "connectedStatus": 0,
      "current": 0.0,
      "disable": false,
      "duplex": 0,
      "linkSpeed": 0,
      "operation": "switching",
      "pdClass": "--",
      "poe": 1,
      "poeDisplayType": 4,
      "port": 2,
      "portName": "Port 2",
      "portStatus": {
        "duplex": 0,
        "linkSpeed": 0,
        "linkStatus": 0,
        "poe": false,
        "poePower": 0.0,
        "port": 2,
        "rx": 0,
        "rxRate": 0,
        "stpDiscarding": false,
        "tx": 0,
        "txRate": 0
      },
      "power": 0.0,
      "switchMac": "AA-BB-CC-00-00-02",
      "switchName": "Switch Basement",
      "switchSupportPoe": 1,
      "supportPoe": true,
      "type": 1,
      "voltage": 0.0
    },
    {
      "connectedStatus": 0,
      "disable": false,
      "duplex": 0,
      "linkSpeed": 0,
      "operation": "switching",
      "pdClass": "--",
      "poe": 1,
      "poeDisplayType": -1,
      "port": 9,
      "portName": "SFP+ 1",
      "portStatus": {
        "duplex": 0,
        "linkSpeed": 3,
        "linkStatus": 0,
        "poe": false,
        "port": 9,
        "rx": 0,
        "rxRate": 0,
        "stpDiscarding": false,
        "tx": 0,
        "txRate": 0
      },
      "switchMac": "AA-BB-CC-00-00-02",
      "switchName": "Switch Basement",
      "switchSupportPoe": 1,
      "supportPoe": false,
      "type": 2
    }
  ]
}`

func parsePoeInfoFixture(t *testing.T) []openApiSwitchPoePort {
	t.Helper()

	var grid struct {
		TotalRows int                    `json:"totalRows"`
		Data      []openApiSwitchPoePort `json:"data"`
	}
	if err := json.Unmarshal([]byte(poeInfoFixture), &grid); err != nil {
		t.Fatalf("cannot unmarshal poe-info fixture: %v", err)
	}
	if len(grid.Data) != grid.TotalRows {
		t.Fatalf("got %d rows, want totalRows %d", len(grid.Data), grid.TotalRows)
	}

	return grid.Data
}

// TestPoeInfoParsing locks the parsing of the poe-info report. The units are
// the point of the test: the controller reports the current in milliamperes, so
// the raw value must stay 150.0 here and only be scaled when it is merged into
// a port.
func TestPoeInfoParsing(t *testing.T) {
	rows := parsePoeInfoFixture(t)

	powered := rows[0]
	if powered.Port != 1 || powered.SwitchMac != "AA-BB-CC-00-00-02" {
		t.Errorf("row 0 identifies port %d of %q, want port 1 of AA-BB-CC-00-00-02", powered.Port, powered.SwitchMac)
	}
	if !powered.SupportPoe {
		t.Error("row 0 supportPoe = false, want true")
	}
	if powered.Power != 7.9 || powered.Voltage != 52.9 {
		t.Errorf("row 0 power/voltage = %v/%v, want 7.9 W / 52.9 V", powered.Power, powered.Voltage)
	}
	if powered.Current != 150.0 {
		t.Errorf("row 0 current = %v, want the raw 150.0 mA from the controller", powered.Current)
	}

	unsupported := rows[2]
	if unsupported.SupportPoe {
		t.Error("row 2 supportPoe = true, want false for a port without PoE support")
	}
	if unsupported.Power != 0 || unsupported.Voltage != 0 || unsupported.Current != 0 {
		t.Errorf("row 2 reports power/voltage/current %v/%v/%v, want 0 (the controller omits them)",
			unsupported.Power, unsupported.Voltage, unsupported.Current)
	}
}

// TestMergePoeInfo covers what the metrics depend on: the rows are matched to
// ports by (switchMac, port), a powered port reports watts/volts/amperes, a
// port that does not support PoE reports those values as not applicable, and a
// port without a row is not touched.
func TestMergePoeInfo(t *testing.T) {
	switches := []Switch{
		{
			MacAddress: "AA-BB-CC-00-00-02",
			PortList: []SwitchPort{
				{Port: 1},
				{Port: 2},
				{Port: 9},
				{Port: 10},
			},
		},
		{
			// Same port numbers, different switch: must stay without PoE.
			MacAddress: "AA-BB-CC-00-00-03",
			PortList:   []SwitchPort{{Port: 1}},
		},
	}

	mergePoeInfo(switches, parsePoeInfoFixture(t))

	powered := switches[0].PortList[0]
	if !powered.PoeSupported {
		t.Error("port 1 PoeSupported = false, want true")
	}
	if powered.GetPoePowerWatts() != 7.9 {
		t.Errorf("port 1 power = %v W, want 7.9", powered.GetPoePowerWatts())
	}
	if powered.GetPoeVoltageVolts() != 52.9 {
		t.Errorf("port 1 voltage = %v V, want 52.9", powered.GetPoeVoltageVolts())
	}
	// 150.0 mA from the controller is 0.150 A. The unit is verifiable against
	// the other two values: 52.9 V * 0.150 A = 7.935 W ~= 7.9 W.
	if math.Abs(powered.GetPoeCurrentAmperes()-0.150) > 1e-9 {
		t.Errorf("port 1 current = %v A, want 0.150 (converted from 150.0 mA)", powered.GetPoeCurrentAmperes())
	}
	if want := powered.GetPoeVoltageVolts() * powered.GetPoeCurrentAmperes(); math.Abs(want-powered.GetPoePowerWatts()) > 0.1 {
		t.Errorf("port 1 reports %v W, want ~%v W from voltage * current", powered.GetPoePowerWatts(), want)
	}

	idle := switches[0].PortList[1]
	if !idle.PoeSupported || idle.GetPoeSupported() != 1 {
		t.Error("port 2 supports PoE but is reported as unsupported")
	}
	if idle.GetPoePowerWatts() != 0 || idle.GetPoeCurrentAmperes() != 0 {
		t.Errorf("port 2 power/current = %v/%v, want 0 for a PoE capable but unpowered port",
			idle.GetPoePowerWatts(), idle.GetPoeCurrentAmperes())
	}

	unsupported := switches[0].PortList[2]
	if unsupported.PoeSupported || unsupported.GetPoeSupported() != 0 {
		t.Error("port 9 does not support PoE but is reported as supporting it")
	}
	if unsupported.GetPoePowerWatts() != Enum.NotApplicable_Float ||
		unsupported.GetPoeVoltageVolts() != Enum.NotApplicable_Float ||
		unsupported.GetPoeCurrentAmperes() != Enum.NotApplicable_Float {
		t.Errorf("port 9 PoE values = %v/%v/%v, want NotApplicable",
			unsupported.GetPoePowerWatts(), unsupported.GetPoeVoltageVolts(), unsupported.GetPoeCurrentAmperes())
	}

	withoutRow := switches[0].PortList[3]
	if withoutRow.PoeSupported || withoutRow.GetPoePowerWatts() != Enum.NotApplicable_Float {
		t.Error("port 10 has no poe-info row and must stay without PoE values")
	}

	otherSwitch := switches[1].PortList[0]
	if otherSwitch.PoeSupported || otherSwitch.GetPoePowerWatts() != Enum.NotApplicable_Float {
		t.Error("port 1 of another switch picked up the PoE row of the first switch")
	}
}
