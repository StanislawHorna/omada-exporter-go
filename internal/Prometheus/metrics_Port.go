package Prometheus

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"omada_exporter_go/internal/Omada/Enum"
	"omada_exporter_go/internal/Omada/Model/Interface"
	"omada_exporter_go/internal/Prometheus/Utils"
)

const (
	label_portName     string = "portName"
	label_portIP       string = "portIP"
	label_portProtocol string = "portProtocol"
	label_portMode     string = "portMode"
)

var portInfoLabels = []string{label_portName, label_portIP, label_portProtocol, label_portMode}

var (
	port_rx_bytes_total = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "port_rx_bytes_total",
			Help: "Total number of bytes received on the port",
		},
		portIdentityLabels,
	)
	port_tx_bytes_total = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "port_tx_bytes_total",
			Help: "Total number of bytes transmitted on the port",
		},
		portIdentityLabels,
	)
	port_speed = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "port_speed",
			Help: "Speed of the port in bits per second",
		},
		portIdentityLabels,
	)
	port_duplex = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "port_duplex",
			Help: fmt.Sprintf(
				"Duplex mode of the port (%s)",
				Enum.GetDuplexPossibleValues(),
			),
		},
		portIdentityLabels,
	)
	port_info = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "port_info",
			Help: "Information about port of the device",
		},
		append(portIdentityLabels, portInfoLabels...),
	)
	port_upstream_state = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "port_upstream_state",
			Help: fmt.Sprintf("Router port upstream state (%s)",
				Enum.GetRouterUpstreamStatePossibleValues(),
			),
		},
		portIdentityLabels,
	)
	port_internet_state = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "port_internet_state",
			Help: fmt.Sprintf("Router port Internet state (%s)",
				Enum.GetInternetStatePossibleValues(),
			),
		},
		portIdentityLabels,
	)
	port_internet_latency = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "port_internet_latency",
			Help: "Router port Internet latency in milliseconds",
		},
		portIdentityLabels,
	)
	port_internet_loss = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "port_internet_loss",
			Help: "Router port Internet loss in percentage (0 - 100)",
		},
		portIdentityLabels,
	)
	port_poe_supported = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "port_poe_supported",
			Help: "Whether the port supports Power over Ethernet (0 = no, 1 = yes)",
		},
		portIdentityLabels,
	)
	port_poe_power_watts = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "port_poe_power_watts",
			Help: "Power delivered to the powered device on the port in watts",
		},
		portIdentityLabels,
	)
	port_poe_voltage_volts = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "port_poe_voltage_volts",
			Help: "Power over Ethernet voltage on the port in volts",
		},
		portIdentityLabels,
	)
	port_poe_current_amperes = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "port_poe_current_amperes",
			Help: "Power over Ethernet current on the port in amperes",
		},
		portIdentityLabels,
	)
)

func ExposePortMetrics(devices []Interface.Device) {
	for _, d := range devices {
		for _, p := range d.GetPorts() {

			labels := getPortIdentityLabels(d, p)
			port_rx_bytes_total.With(labels).Set(float64(p.GetRxBytes()))
			port_tx_bytes_total.With(labels).Set(float64(p.GetTxBytes()))

			port_speed.With(labels).Set(p.GetPortSpeed())
			port_duplex.With(labels).Set(p.GetPortDuplex())

			if p.GetInternetState() != Enum.NotApplicable_Float {
				port_internet_state.With(labels).Set(p.GetInternetState())
			}
			if p.GetUpstreamState() != Enum.NotApplicable_Float {
				port_upstream_state.With(labels).Set(p.GetUpstreamState())
			}
			if p.GetInternetLatency() != Enum.NotApplicable_Float {
				port_internet_latency.With(labels).Set(p.GetInternetLatency())
			}
			if p.GetInternetLoss() != Enum.NotApplicable_Float {
				port_internet_loss.With(labels).Set(p.GetInternetLoss())
			}

			setPortInfo(p, labels)
			setPortPoe(p, labels)
		}
	}

}

// setPortPoe exposes the Power over Ethernet values of a port. Access point and
// gateway ports do not report PoE at all and ports without PoE support have no
// values, so their series are deleted instead of kept with a stale value
// (mirroring setDeviceTemperature).
func setPortPoe(port Interface.Port, labels prometheus.Labels) {
	if supported := port.GetPoeSupported(); supported != Enum.NotApplicable_Float {
		port_poe_supported.With(labels).Set(supported)
	} else {
		port_poe_supported.Delete(labels)
	}

	setPortPoeValue(port_poe_power_watts, port.GetPoePowerWatts(), labels)
	setPortPoeValue(port_poe_voltage_volts, port.GetPoeVoltageVolts(), labels)
	setPortPoeValue(port_poe_current_amperes, port.GetPoeCurrentAmperes(), labels)
}

func setPortPoeValue(gauge *prometheus.GaugeVec, value float64, labels prometheus.Labels) {
	if value > Enum.NotApplicable_Float {
		gauge.With(labels).Set(value)
		return
	}
	gauge.Delete(labels)
}

func setPortInfo(port Interface.Port, labels prometheus.Labels) {
	// Delete all info metrics to avoid duplicates created due to changed labels
	// new set of labels always creates new series, but old one is not deleted,
	// even if it was not set in the current iteration
	port_info.DeletePartialMatch(labels)
	port_info.With(Utils.AppendMaps(labels, map[string]string{
		label_portName:     port.GetPortName(),
		label_portIP:       port.GetPortIP(),
		label_portProtocol: port.GetPortProtocol(),
		label_portMode:     port.GetPortMode(),
	},
	)).Set(1)
}
