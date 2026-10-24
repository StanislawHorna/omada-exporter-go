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
	label_deviceName      string = "deviceName"
	label_deviceModel     string = "deviceModel"
	label_IP              string = "IP"
	label_deviceFirmware  string = "deviceFirmware"
	label_HardwareVersion string = "hardwareVersion"
)

var deviceInfoLabels = []string{label_deviceName, label_deviceModel, label_IP, label_deviceFirmware, label_HardwareVersion}

var (
	cpu_usage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cpu_usage",
			Help: "The percentage of CPU usage (0 - 100)",
		},
		deviceIdentityLabels,
	)

	memory_usage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "memory_usage",
			Help: "The percentage of memory usage (0 - 100)",
		},
		deviceIdentityLabels,
	)

	temperature = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "temperature",
			Help: "The device temperature in degrees Celsius",
		},
		deviceIdentityLabels,
	)

	device_info = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "device_info",
			Help: "Information about the device",
		},
		append(deviceIdentityLabels, deviceInfoLabels...),
	)

	device_last_seen = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "device_last_seen",
			Help: "The last time the device was seen, as a Unix timestamp in seconds",
		},
		deviceIdentityLabels,
	)

	device_upgrade_needed = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "device_upgrade_needed",
			Help: fmt.Sprintf("The bool value whether there are any upgrades to install (%s)",
				Enum.GetUpgradeNeededPossibleValues(),
			),
		},
		deviceIdentityLabels,
	)

	device_clients_count = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "device_clients_count",
			Help: "Number of clients connected to device",
		},
		deviceIdentityLabels,
	)
)

func ExposeDeviceMetrics(devices []Interface.Device) {
	for _, d := range devices {
		identityLabels := getDeviceIdentityLabels(d)

		cpu_usage.With(identityLabels).Set(d.GetCpuUsage())
		memory_usage.With(identityLabels).Set(d.GetMemUsage())
		// The controller reports lastSeen in milliseconds; the metric is in Unix
		// seconds (see Utils.UnixSeconds for why the scale is detected).
		device_last_seen.With(identityLabels).Set(Utils.UnixSeconds(d.GetLastSeen()))

		device_clients_count.With(identityLabels).Set(d.GetClientsCount())
		device_upgrade_needed.With(identityLabels).Set(d.GetUpgradeNeededStatus())

		setDeviceTemperature(d, identityLabels)
		setDeviceInfo(d, identityLabels)
	}

}

func setDeviceTemperature(device Interface.Device, labels prometheus.Labels) {
	temp := device.GetTemperature()
	if temp > Enum.NotApplicable_Float {
		temperature.With(labels).Set(temp)
	} else {
		temperature.Delete(labels)
	}
}

func setDeviceInfo(device Interface.Device, labels prometheus.Labels) {
	// Delete all info metrics to avoid duplicates created due to changed labels
	// new set of labels always creates new series, but old one is not deleted,
	// even if it was not set in the current iteration
	device_info.DeletePartialMatch(labels)
	device_info.With(Utils.AppendMaps(map[string]string{
		label_deviceName:      device.GetName(),
		label_deviceModel:     device.GetModel(),
		label_IP:              device.GetIP(),
		label_deviceFirmware:  device.GetFirmware(),
		label_HardwareVersion: device.GetHardwareVersion(),
	}, labels)).Set(1)
}
