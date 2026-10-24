package Devices

import (
	"fmt"

	"omada_exporter_go/internal/Omada/Enum"
)

const path_OpenApiDevicesList = "/openapi/v1/{omadaID}/sites/{siteID}/devices"

type Device struct {
	// OpenAPI fields
	MacAddress string            `json:"mac"`
	Name       string            `json:"name"`
	Type       Enum.DeviceType   `json:"type"`
	Subtype    string            `json:"subtype"`
	Model      string            `json:"model"`
	IP         string            `json:"ip"`
	Uptime     string            `json:"uptime"`
	LastSeen   float64           `json:"lastSeen"`
	Status     Enum.DeviceStatus `json:"status"`
	CpuUsage   int               `json:"cpuUtil"`
	RamUsage   int               `json:"memUtil"`
	TagName    string            `json:"tagName"`

	// WebAPI fields
	ClientNum     int64
	Version       string
	LatestVersion string
	UpgradeNeeded Enum.UpgradeNeeded
}

func (d *Device) GetStatus() string {
	return d.Status.String()
}

func (d *Device) merge(toMerge webApiDevice) error {
	if d.MacAddress != toMerge.MacAddress {
		return fmt.Errorf("cannot merge Devices with different mac addresses")
	}
	d.ClientNum = toMerge.ClientNum
	d.Version = toMerge.Version
	d.LatestVersion = toMerge.LatestVersion
	d.UpgradeNeeded = toMerge.UpgradeNeeded
	return nil
}
