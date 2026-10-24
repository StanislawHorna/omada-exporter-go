package Devices

import (
	"omada_exporter_go/internal/Log"
	"omada_exporter_go/internal/Omada/HttpClient/ApiClient"
	"omada_exporter_go/internal/Omada/HttpClient/WebClient"
)

func Get() (*[]Device, error) {
	Log.Debug("Fetching generic devices data")
	client := ApiClient.GetInstance()

	result, err := ApiClient.Get[Device](client, path_OpenApiDevicesList, nil, nil, true)
	if err != nil {
		return nil, Log.Error(err, "Failed to get devices data")
	}
	webApiData, err := requestWebApiData()
	if err != nil {
		Log.Warn("Failed to get Devices List from WebAPI")
		return result, nil
	}

	for i := range *result {
		for _, webDevice := range webApiData {
			if (*result)[i].MacAddress == webDevice.MacAddress {
				if err := (*result)[i].merge(webDevice); err != nil {
					Log.Error(err, "Failed to merge WebAPI data into device %s", (*result)[i].MacAddress)
				}
			}
		}
	}
	return result, nil
}

func requestWebApiData() ([]webApiDevice, error) {
	client := WebClient.GetInstance()

	result, err := WebClient.GetList[webApiDevice](client, path_WebApiDevicesList, nil, nil, true)
	if err != nil {
		return nil, err
	}
	if len((*result)) == 0 {
		Log.Warn("No Devices returned via WebAPI")
		return nil, nil
	}
	return *result, nil
}
