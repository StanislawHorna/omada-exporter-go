package AccessPoint

import "omada_exporter_go/internal/Omada/Enum"

// Implements Interface.Port
type AccessPointPort struct {
	PortReceiveBytes    int64
	PortTransmitBytes   int64
	PortReceivePackets  int64
	PortTransmitPackets int64
}

func (app AccessPointPort) GetID() string {
	// Access Point Port does not have a specific port number, return 1
	return "1"
}
func (app AccessPointPort) GetRxBytes() float64 {
	return float64(app.PortReceiveBytes)
}
func (app AccessPointPort) GetTxBytes() float64 {
	return float64(app.PortTransmitBytes)
}
func (app AccessPointPort) GetPortSpeed() float64 {
	// Access Point Port does not have a specific speed, return 0
	return 0
}
func (app AccessPointPort) GetPortDuplex() float64 {
	// Access Point Port does not have a specific duplex mode, return 0
	return 0
}
func (app AccessPointPort) GetPortName() string {
	return "AP Port"
}
func (app AccessPointPort) GetPortIP() string {
	// Access Point ports do not have a specific IP address, return empty string
	return Enum.NotApplicable_String
}
func (app AccessPointPort) GetPortProtocol() string {
	// Access Point ports do not have a specific protocol, return empty string
	return Enum.NotApplicable_String
}
func (app AccessPointPort) GetPortMode() string {
	// Access Point ports do not have a specific mode, return empty string
	return Enum.NotApplicable_String
}
func (app AccessPointPort) GetInternetState() float64 {
	// Access Point do not have an internet state, return NotApplicable
	return Enum.NotApplicable_Float
}
func (app AccessPointPort) GetUpstreamState() float64 {
	// Access Point do not have an upstream state, return NotApplicable
	return Enum.NotApplicable_Float
}
func (app AccessPointPort) GetInternetLatency() float64 {
	return Enum.NotApplicable_Float
}
func (app AccessPointPort) GetInternetLoss() float64 {
	return Enum.NotApplicable_Float
}

func (app AccessPointPort) GetPoeSupported() float64 {
	// Access Point ports do not report Power over Ethernet
	return Enum.NotApplicable_Float
}
func (app AccessPointPort) GetPoePowerWatts() float64 {
	return Enum.NotApplicable_Float
}
func (app AccessPointPort) GetPoeVoltageVolts() float64 {
	return Enum.NotApplicable_Float
}
func (app AccessPointPort) GetPoeCurrentAmperes() float64 {
	return Enum.NotApplicable_Float
}
