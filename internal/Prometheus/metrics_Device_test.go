package Prometheus

import (
	"math"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"

	"omada_exporter_go/internal/Omada/Model/Interface"
)

// fakeDevice is the smallest possible Interface.Device: the device identity is
// what the metric is keyed by, everything else only feeds unrelated series.
type fakeDevice struct {
	deviceType string
	macAddress string
	lastSeen   float64
	ports      []Interface.Port
}

func (d fakeDevice) GetType() string                 { return d.deviceType }
func (d fakeDevice) GetMacAddress() string           { return d.macAddress }
func (d fakeDevice) GetName() string                 { return "test-device" }
func (d fakeDevice) GetIP() string                   { return "192.0.2.10" }
func (d fakeDevice) GetFirmware() string             { return "1.0.0" }
func (d fakeDevice) GetModel() string                { return "TEST-MODEL v1.0" }
func (d fakeDevice) GetHardwareVersion() string      { return "1.0" }
func (d fakeDevice) GetCpuUsage() float64            { return 1 }
func (d fakeDevice) GetMemUsage() float64            { return 53 }
func (d fakeDevice) GetTemperature() float64         { return -1 }
func (d fakeDevice) GetLastSeen() float64            { return d.lastSeen }
func (d fakeDevice) GetClientsCount() float64        { return 3 }
func (d fakeDevice) GetUpgradeNeededStatus() float64 { return 0 }
func (d fakeDevice) GetPorts() []Interface.Port      { return d.ports }
func (d fakeDevice) GetRadios() []Interface.Radio    { return nil }

// gatherLastSeen returns the gathered device_last_seen family plus the sample
// that carries exactly the given labels.
func gatherLastSeen(t *testing.T, want map[string]string) (*dto.MetricFamily, *dto.Metric) {
	t.Helper()

	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("cannot gather metrics: %v", err)
	}

	for _, family := range families {
		if family.GetName() != "device_last_seen" {
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
				return family, metric
			}
		}
	}

	t.Fatalf("no device_last_seen sample with labels %v", want)
	return nil, nil
}

// device_last_seen is consumed by Grafana date/time units, which interpret the
// number as a Unix epoch in seconds. The controller reports milliseconds, so the
// exporter has to scale it down; otherwise a panel shows 1.79e+12 or, when the
// number is read as seconds, a year around 58,700. The timestamps below are from
// a real scrape, the identifiers are anonymised.
func TestExposeDeviceMetrics_LastSeenIsUnixSeconds(t *testing.T) {
	tests := []struct {
		name       string
		device     fakeDevice
		wantSecond float64
	}{
		{
			name:       "access point, milliseconds from the device list",
			device:     fakeDevice{deviceType: "AccessPoint", macAddress: "AA-BB-CC-00-00-01", lastSeen: 1792833739702},
			wantSecond: 1792833739.702,
		},
		{
			name:       "switch, milliseconds from the device list",
			device:     fakeDevice{deviceType: "Switch", macAddress: "AA-BB-CC-00-00-02", lastSeen: 1792833715370},
			wantSecond: 1792833715.370,
		},
		{
			name:       "gateway, already reported in seconds",
			device:     fakeDevice{deviceType: "Gateway", macAddress: "AA-BB-CC-00-00-03", lastSeen: 1792833700},
			wantSecond: 1792833700,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ExposeDeviceMetrics([]Interface.Device{tt.device})

			_, metric := gatherLastSeen(t, map[string]string{
				label_deviceType: tt.device.deviceType,
				label_macAddress: tt.device.macAddress,
			})

			got := metric.GetGauge().GetValue()
			if math.Abs(got-tt.wantSecond) > 1e-6 {
				t.Fatalf("device_last_seen = %v, want %v (Unix seconds)", got, tt.wantSecond)
			}
		})
	}
}

// TestDeviceLastSeenHelp locks the documented unit: the value is only useful if
// its scale is written down, a bare "Unix timestamp" reads like seconds and is
// what made the original millisecond leak hard to spot.
func TestDeviceLastSeenHelp(t *testing.T) {
	family, _ := gatherLastSeen(t, map[string]string{
		label_deviceType: "AccessPoint",
		label_macAddress: "AA-BB-CC-00-00-01",
	})

	if help := family.GetHelp(); !strings.Contains(help, "Unix timestamp in seconds") {
		t.Fatalf("device_last_seen help = %q, want it to mention the unit", help)
	}
}
