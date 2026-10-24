package Switch

import (
	"omada_exporter_go/internal/Log"
	"omada_exporter_go/internal/Omada/Enum"
	"omada_exporter_go/internal/Omada/HttpClient/ApiClient"
	"omada_exporter_go/internal/Omada/HttpClient/WebClient"
	"omada_exporter_go/internal/Omada/Model/Devices"
)

func Get(devices []Devices.Device) (*[]Switch, error) {
	Log.Debug("Fetching switches data")
	var allDataOpenApi []Switch

	for _, d := range devices {
		if d.Type != Enum.DeviceType_Switch {
			continue
		}

		openApiResult, err := getOpenApiData(d)
		// If OpenAPI data is not available,
		// return error because OpenAPI data is base information, completed with WebAPI data
		if err != nil {
			return nil, Log.Error(err, "Failed to get OpenAPI data for switch %s", d.MacAddress)
		}

		// getOpenApiData also reports a nil result when the controller
		// returned no data; dereferencing that below would panic.
		if openApiResult == nil || len(*openApiResult) == 0 {
			continue
		}

		webApiResult, err := getWebApiData(d)
		if err != nil {
			Log.Error(err, "Failed to get WebAPI data for switch %s", d.MacAddress)
		} else {
			for i := range (*openApiResult)[0].PortList {
				// Merge the web API data into the OpenAPI result
				for _, webPort := range *webApiResult {
					if (*openApiResult)[0].PortList[i].Port == webPort.Port {
						if err := (*openApiResult)[0].PortList[i].merge(webPort); err != nil {
							Log.Error(err, "Failed to merge port data for switch %s", d.MacAddress)
						}
						// If port is down set speed and duplex as disabled
						if (*openApiResult)[0].PortList[i].LinkStatus == Enum.LinkStatus_Down {
							(*openApiResult)[0].PortList[i].LinkSpeed = Enum.LinkSpeed_Disabled
							(*openApiResult)[0].PortList[i].DuplexMode = Enum.DuplexMode_Down
						}
						break
					}
				}
			}
		}
		allDataOpenApi = append(allDataOpenApi, *openApiResult...)

	}

	// The PoE report is site-wide (one row per port of every switch), so it is
	// requested once per scrape instead of once per switch. It only decorates
	// the ports: a failing request must never cost the switch and port metrics,
	// which is why it is logged and skipped rather than returned.
	poePorts, err := getPoeInfo()
	if err != nil {
		Log.Warn("Failed to get PoE port data, ports will report as PoE unsupported: %s", err)
	} else {
		mergePoeInfo(allDataOpenApi, poePorts)
	}

	Log.Debug("Fetched %d switches", len(allDataOpenApi))

	return &allDataOpenApi, nil
}

func getPoeInfo() ([]openApiSwitchPoePort, error) {
	client := ApiClient.GetInstance()

	result, err := ApiClient.Get[openApiSwitchPoePort](client, path_OpenApiSwitchPoe, nil, nil, true)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	return *result, nil
}

func getOpenApiData(d Devices.Device) (*[]Switch, error) {
	client := ApiClient.GetInstance()

	// Agile Series ("Easy Managed", model prefix ES) switches are only
	// served by the dedicated /switches/es/{switchMac} endpoint; the generic
	// one answers errorCode -39742. The device list reports the series in
	// `subtype` ("es" for Agile Series).
	path := path_OpenApiSwitch
	isEasyManaged := d.Subtype == "es"
	if isEasyManaged {
		path = path_OpenApiEasyManagedSwitch
	}

	result, err := ApiClient.Get[Switch](client, path, map[string]string{"switchMac": d.MacAddress}, nil, false)
	if err != nil {
		return nil, err
	}

	if result == nil || len(*result) == 0 {
		Log.Warn("No OpenAPI data found for switch %s", d.MacAddress)
		return nil, nil
	}

	// Set the device type and name based on device entry
	(*result)[0].DeviceType = Enum.DeviceType_Switch
	(*result)[0].Name = d.Name
	(*result)[0].LastSeen = d.LastSeen

	// The Agile Series overview omits cpuUtil/memUtil; the device list still
	// carries them, so use those instead of reporting 0%.
	if isEasyManaged {
		(*result)[0].CpuUsage = d.CpuUsage
		(*result)[0].RamUsage = d.RamUsage
	}

	return result, nil
}

func getWebApiData(d Devices.Device) (*[]webApiSwitchPort, error) {
	client := WebClient.GetInstance()

	result, err := WebClient.GetList[webApiSwitchPort](client, path_WebApiSwitchPort, map[string]string{"switchMac": d.MacAddress}, nil, false)
	if err != nil {
		return nil, err
	}
	if len(*result) == 0 {
		Log.Warn("No WebAPI data found for switch %s", d.MacAddress)
		return nil, nil
	}

	return result, nil

}
