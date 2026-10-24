package Switch

// path_OpenApiSwitchPoe is the site-wide Power over Ethernet port report. It is
// not scoped to a single switch: one page holds one row per port of every
// switch in the site, which is why it is fetched once per scrape (see Get) and
// merged into the switches by (switchMac, port).
const path_OpenApiSwitchPoe = "/openapi/v1/{omadaID}/sites/{siteID}/switches/ports/poe-info"

// openApiSwitchPoePort is one row of the poe-info report. Only the fields that
// feed the metrics are modelled; the payload carries many more (linkSpeed,
// duplex, portStatus, ...) which are already available per switch.
//
// The controller omits power, voltage and current entirely for ports that do
// not support PoE (`supportPoe: false`); those fields then decode to 0. The
// `poe` field is 1 even for ports without PoE support and does not mean "is
// currently delivering power", so it is not used.
type openApiSwitchPoePort struct {
	Port       int     `json:"port"`
	SwitchMac  string  `json:"switchMac"`
	SupportPoe bool    `json:"supportPoe"`
	Power      float64 `json:"power"`
	Voltage    float64 `json:"voltage"`
	Current    float64 `json:"current"` // milliamperes, see SwitchPort.applyPoeInfo
}

// poePortKey identifies a port across the per-switch endpoint and the site-wide
// PoE report. The controller uses the same MAC format (uppercase, dash
// separated) in both, so the values can be compared as they are.
type poePortKey struct {
	switchMac string
	port      int
}

// mergePoeInfo copies the PoE values of the site-wide report onto the matching
// switch ports. Ports without a row keep their zero values and are reported as
// ports without PoE support.
func mergePoeInfo(switches []Switch, poePorts []openApiSwitchPoePort) {
	rows := make(map[poePortKey]openApiSwitchPoePort, len(poePorts))
	for _, row := range poePorts {
		rows[poePortKey{switchMac: row.SwitchMac, port: row.Port}] = row
	}

	for i := range switches {
		for j := range switches[i].PortList {
			key := poePortKey{switchMac: switches[i].MacAddress, port: switches[i].PortList[j].Port}
			if row, ok := rows[key]; ok {
				switches[i].PortList[j].applyPoeInfo(row)
			}
		}
	}
}
