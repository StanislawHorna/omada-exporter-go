package Gateway

import (
	"fmt"

	"omada_exporter_go/internal/Omada/Enum"
)

const path_OpenApiGateway = "/openapi/v1/{omadaID}/sites/{siteID}/gateways/{gatewayMac}"

// Implements Interface.Port
type GatewayPort struct {
	// OpenAPI fields
	Port          int             `json:"port"`
	LinkSpeed     Enum.LinkSpeed  `json:"linkSpeed"`
	DuplexMode    Enum.DuplexMode `json:"duplexMode"`
	MirrorEnabled bool            `json:"mirrorEnabled"`
	MirroredPorts []int           `json:"mirroredPorts"`
	MirrorMode    Enum.MirrorMode `json:"mirrorMode"`
	PortMode      int8            `json:"pvid"`

	// WebAPI fields
	PortName        string
	PortDescription string
	Mode            Enum.GatewayPortMode
	IP              string
	Protocol        string
	Poe             bool
	LinkStatus      Enum.LinkStatus
	InternetState   Enum.GatewayInternetState
	Online          Enum.RouterUpstreamState
	ReceiveBytes    int64
	TransmitBytes   int64
	ReceivePackets  int64
	TransmitPackets int64
	Latency         float64
	Loss            float64
	IPv4Config      webApiWanPortIpv4Config
	IPv6Config      webApiWanPortIpv6Config
}

func (gp GatewayPort) GetID() string {
	return fmt.Sprintf("%d", gp.Port)
}
func (gp GatewayPort) GetRxBytes() float64 {
	return float64(gp.ReceiveBytes)
}
func (gp GatewayPort) GetTxBytes() float64 {
	return float64(gp.TransmitBytes)
}
func (gp GatewayPort) GetPortName() string {
	if gp.Mode != Enum.GatewayPortMode_WAN {
		return gp.PortName
	}
	return gp.PortDescription
}
func (gp GatewayPort) GetPortSpeed() float64 {
	return float64(gp.LinkSpeed.Int())
}
func (gp GatewayPort) GetPortDuplex() float64 {
	return float64(gp.DuplexMode.Int())
}

func (gp GatewayPort) GetPortIP() string {
	if gp.Mode != Enum.GatewayPortMode_WAN {
		return Enum.NotApplicable_String
	}
	return gp.IP
}
func (gp GatewayPort) GetPortMode() string {
	return gp.Mode.String()
}
func (gp GatewayPort) GetPortProtocol() string {
	if gp.Mode != Enum.GatewayPortMode_WAN {
		return Enum.NotApplicable_String
	}
	return gp.Protocol
}
func (gp GatewayPort) GetInternetState() float64 {
	return float64(gp.InternetState.Int())
}
func (gp GatewayPort) GetUpstreamState() float64 {
	return float64(gp.Online.Int())
}
func (gp GatewayPort) GetInternetLatency() float64 {
	if gp.Mode != Enum.GatewayPortMode_WAN {
		return -1
	}
	return gp.Latency
}
func (gp GatewayPort) GetInternetLoss() float64 {
	if gp.Mode != Enum.GatewayPortMode_WAN {
		return -1
	}
	return gp.Loss
}
func (gp GatewayPort) GetPoeSupported() float64 {
	// Gateway ports do not report Power over Ethernet
	return Enum.NotApplicable_Float
}
func (gp GatewayPort) GetPoePowerWatts() float64 {
	return Enum.NotApplicable_Float
}
func (gp GatewayPort) GetPoeVoltageVolts() float64 {
	return Enum.NotApplicable_Float
}
func (gp GatewayPort) GetPoeCurrentAmperes() float64 {
	return Enum.NotApplicable_Float
}

func (gp *GatewayPort) merge(toMerge webApiGatewayPort) error {
	if gp.Port != toMerge.Port {
		return fmt.Errorf("cannot merge GatewayPort with different port numbers: %d != %d", gp.Port, toMerge.Port)
	}
	gp.PortName = toMerge.PortName
	gp.PortDescription = toMerge.PortDesc
	gp.Mode = toMerge.Mode
	gp.IP = toMerge.IP
	gp.Protocol = toMerge.Protocol
	gp.Poe = toMerge.Poe
	gp.LinkStatus = toMerge.LinkStatus
	gp.InternetState = toMerge.InternetState
	gp.Online = toMerge.Online
	gp.LinkSpeed = toMerge.LinkSpeed
	gp.DuplexMode = toMerge.Duplex
	gp.ReceivePackets = toMerge.RxPackets
	gp.TransmitPackets = toMerge.TxPackets
	gp.Latency = toMerge.Latency
	gp.Loss = toMerge.Loss
	gp.IPv4Config = toMerge.WanPortIpv4Config
	gp.IPv6Config = toMerge.WanPortIpv6Config

	gp.ReceiveBytes = toMerge.Rx
	gp.TransmitBytes = toMerge.Tx

	return nil
}
