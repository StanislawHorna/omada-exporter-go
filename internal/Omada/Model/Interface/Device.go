package Interface

import (
	"fmt"
)

type Device interface {
	// Getters for device properties
	GetType() string
	GetMacAddress() string
	GetName() string
	GetIP() string
	GetFirmware() string
	GetModel() string
	GetHardwareVersion() string

	// Getters for device resource consumption
	GetCpuUsage() float64
	GetMemUsage() float64

	GetTemperature() float64 // Returns -1 if temperature is not available
	GetLastSeen() float64    // Returns the last seen timestamp as reported by the controller, in milliseconds (see Prometheus/Utils.UnixSeconds)
	GetClientsCount() float64
	GetUpgradeNeededStatus() float64

	// Getters for object associated with device
	GetPorts() []Port
	GetRadios() []Radio
}

func AppendDevicesSlice[T Device](devices *[]Device, newDevices []T) error {
	if devices == nil {
		return fmt.Errorf("devices is nil")
	}

	for i := range newDevices {
		*devices = append(*devices, newDevices[i])
	}

	return nil
}
