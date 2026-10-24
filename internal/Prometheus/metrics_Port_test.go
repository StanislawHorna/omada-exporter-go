package Prometheus

import (
	"math"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"

	"omada_exporter_go/internal/Omada/Enum"
	"omada_exporter_go/internal/Omada/Model/Interface"
)

// fakePort is the smallest possible Interface.Port: the PoE getters are what
// the metrics below are keyed by, everything else only feeds unrelated series.
type fakePort struct {
	portID    string
	supported float64
	power     float64
	voltage   float64
	current   float64
}

func (p fakePort) GetID() string                 { return p.portID }
func (p fakePort) GetPortName() string           { return "Port " + p.portID }
func (p fakePort) GetPortSpeed() float64         { return 1000000000 }
func (p fakePort) GetPortDuplex() float64        { return 2 }
func (p fakePort) GetPortIP() string             { return Enum.NotApplicable_String }
func (p fakePort) GetPortMode() string           { return Enum.NotApplicable_String }
func (p fakePort) GetPortProtocol() string       { return Enum.NotApplicable_String }
func (p fakePort) GetTxBytes() float64           { return 0 }
func (p fakePort) GetRxBytes() float64           { return 0 }
func (p fakePort) GetInternetState() float64     { return Enum.NotApplicable_Float }
func (p fakePort) GetInternetLatency() float64   { return Enum.NotApplicable_Float }
func (p fakePort) GetInternetLoss() float64      { return Enum.NotApplicable_Float }
func (p fakePort) GetUpstreamState() float64     { return Enum.NotApplicable_Float }
func (p fakePort) GetPoeSupported() float64      { return p.supported }
func (p fakePort) GetPoePowerWatts() float64     { return p.power }
func (p fakePort) GetPoeVoltageVolts() float64   { return p.voltage }
func (p fakePort) GetPoeCurrentAmperes() float64 { return p.current }

// poePortLabels are the labels every PoE series is keyed by.
func poePortLabels(macAddress string, portID string) map[string]string {
	return map[string]string{
		label_deviceType: "Switch",
		label_macAddress: macAddress,
		label_portID:     portID,
	}
}

// gatherPoeSample returns the sample of the given PoE metric that carries
// exactly the given labels, or nil when the series does not exist.
func gatherPoeSample(t *testing.T, metricName string, want map[string]string) *dto.Metric {
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

func requirePoeSample(t *testing.T, metricName string, want map[string]string) float64 {
	t.Helper()

	metric := gatherPoeSample(t, metricName, want)
	if metric == nil {
		t.Fatalf("no %s sample with labels %v", metricName, want)
	}
	return metric.GetGauge().GetValue()
}

func requireNoPoeSample(t *testing.T, metricName string, want map[string]string) {
	t.Helper()

	if metric := gatherPoeSample(t, metricName, want); metric != nil {
		t.Fatalf("%s has a sample with labels %v, want no series", metricName, want)
	}
}

// TestExposePortMetrics_Poe covers the three kinds of switch ports and the
// device classes without PoE: a powered port reports all four series, a PoE
// capable but unpowered port reports 0, a port without PoE support reports only
// port_poe_supported=0 (deleting the value series) and an access point port
// reports nothing at all.
func TestExposePortMetrics_Poe(t *testing.T) {
	device := fakeDevice{
		deviceType: "Switch",
		macAddress: "AA-BB-CC-00-00-10",
		ports: []Interface.Port{
			fakePort{portID: "1", supported: 1, power: 7.9, voltage: 52.9, current: 0.150},
			fakePort{portID: "2", supported: 1},
			fakePort{portID: "9", supported: 0, power: Enum.NotApplicable_Float, voltage: Enum.NotApplicable_Float, current: Enum.NotApplicable_Float},
		},
	}
	accessPoint := fakeDevice{
		deviceType: "AccessPoint",
		macAddress: "AA-BB-CC-00-00-11",
		ports: []Interface.Port{
			fakePort{portID: "1", supported: Enum.NotApplicable_Float, power: Enum.NotApplicable_Float,
				voltage: Enum.NotApplicable_Float, current: Enum.NotApplicable_Float},
		},
	}

	ExposePortMetrics([]Interface.Device{device, accessPoint})

	powered := poePortLabels(device.macAddress, "1")
	if got := requirePoeSample(t, "port_poe_supported", powered); got != 1 {
		t.Errorf("port 1 port_poe_supported = %v, want 1", got)
	}
	if got := requirePoeSample(t, "port_poe_power_watts", powered); math.Abs(got-7.9) > 1e-9 {
		t.Errorf("port 1 port_poe_power_watts = %v, want 7.9", got)
	}
	if got := requirePoeSample(t, "port_poe_voltage_volts", powered); math.Abs(got-52.9) > 1e-9 {
		t.Errorf("port 1 port_poe_voltage_volts = %v, want 52.9", got)
	}
	if got := requirePoeSample(t, "port_poe_current_amperes", powered); math.Abs(got-0.150) > 1e-9 {
		t.Errorf("port 1 port_poe_current_amperes = %v, want 0.150", got)
	}

	// A PoE capable port without a powered device is a real measurement of 0 W,
	// not a gap: the dashboards show 0 instead of "no data".
	idle := poePortLabels(device.macAddress, "2")
	if got := requirePoeSample(t, "port_poe_supported", idle); got != 1 {
		t.Errorf("port 2 port_poe_supported = %v, want 1", got)
	}
	if got := requirePoeSample(t, "port_poe_power_watts", idle); got != 0 {
		t.Errorf("port 2 port_poe_power_watts = %v, want 0", got)
	}
	if got := requirePoeSample(t, "port_poe_voltage_volts", idle); got != 0 {
		t.Errorf("port 2 port_poe_voltage_volts = %v, want 0", got)
	}
	if got := requirePoeSample(t, "port_poe_current_amperes", idle); got != 0 {
		t.Errorf("port 2 port_poe_current_amperes = %v, want 0", got)
	}

	unsupported := poePortLabels(device.macAddress, "9")
	if got := requirePoeSample(t, "port_poe_supported", unsupported); got != 0 {
		t.Errorf("port 9 port_poe_supported = %v, want 0", got)
	}
	requireNoPoeSample(t, "port_poe_power_watts", unsupported)
	requireNoPoeSample(t, "port_poe_voltage_volts", unsupported)
	requireNoPoeSample(t, "port_poe_current_amperes", unsupported)

	apPort := map[string]string{
		label_deviceType: "AccessPoint",
		label_macAddress: accessPoint.macAddress,
		label_portID:     "1",
	}
	requireNoPoeSample(t, "port_poe_supported", apPort)
	requireNoPoeSample(t, "port_poe_power_watts", apPort)
	requireNoPoeSample(t, "port_poe_voltage_volts", apPort)
	requireNoPoeSample(t, "port_poe_current_amperes", apPort)
}

// TestExposePortMetrics_PoeSeriesDropped locks the stale-series rule: when a
// port stops supporting PoE (or its device is reported without PoE data), the
// value series has to disappear instead of keeping the last delivered power
// forever.
func TestExposePortMetrics_PoeSeriesDropped(t *testing.T) {
	macAddress := "AA-BB-CC-00-00-12"
	labels := poePortLabels(macAddress, "7")

	powered := fakeDevice{
		deviceType: "Switch",
		macAddress: macAddress,
		ports:      []Interface.Port{fakePort{portID: "7", supported: 1, power: 3.4, voltage: 52.9, current: 0.066}},
	}
	ExposePortMetrics([]Interface.Device{powered})
	if got := requirePoeSample(t, "port_poe_power_watts", labels); math.Abs(got-3.4) > 1e-9 {
		t.Fatalf("port 7 port_poe_power_watts = %v, want 3.4", got)
	}

	withoutPoe := fakeDevice{
		deviceType: "Switch",
		macAddress: macAddress,
		ports: []Interface.Port{fakePort{portID: "7", supported: 0, power: Enum.NotApplicable_Float,
			voltage: Enum.NotApplicable_Float, current: Enum.NotApplicable_Float}},
	}
	ExposePortMetrics([]Interface.Device{withoutPoe})

	if got := requirePoeSample(t, "port_poe_supported", labels); got != 0 {
		t.Errorf("port 7 port_poe_supported = %v, want 0", got)
	}
	requireNoPoeSample(t, "port_poe_power_watts", labels)
	requireNoPoeSample(t, "port_poe_voltage_volts", labels)
	requireNoPoeSample(t, "port_poe_current_amperes", labels)
}

// TestPortPoeHelp locks the documented units: a bare number is useless in a
// dashboard if the gauge does not say whether it is watts, volts or amperes.
func TestPortPoeHelp(t *testing.T) {
	tests := []struct {
		metricName string
		wantUnit   string
	}{
		{metricName: "port_poe_supported", wantUnit: "0 = no, 1 = yes"},
		{metricName: "port_poe_power_watts", wantUnit: "in watts"},
		{metricName: "port_poe_voltage_volts", wantUnit: "in volts"},
		{metricName: "port_poe_current_amperes", wantUnit: "in amperes"},
	}

	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("cannot gather metrics: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.metricName, func(t *testing.T) {
			for _, family := range families {
				if family.GetName() != tt.metricName {
					continue
				}
				if help := family.GetHelp(); !strings.Contains(help, tt.wantUnit) {
					t.Fatalf("%s help = %q, want it to mention %q", tt.metricName, help, tt.wantUnit)
				}
				return
			}
			t.Fatalf("%s is not gathered, so its help text cannot be checked", tt.metricName)
		})
	}
}
