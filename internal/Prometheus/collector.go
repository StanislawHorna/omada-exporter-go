package Prometheus

import (
	"errors"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"omada_exporter_go/internal"
	"omada_exporter_go/internal/Log"
	"omada_exporter_go/internal/Omada/Model/AccessPoint"
	"omada_exporter_go/internal/Omada/Model/Clients"
	"omada_exporter_go/internal/Omada/Model/Devices"
	"omada_exporter_go/internal/Omada/Model/Gateway"
	"omada_exporter_go/internal/Omada/Model/Interface"
	"omada_exporter_go/internal/Omada/Model/Switch"
)

var (
	omada_scrape_duration = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "omada_scrape_duration",
			Help: "The amount of time it took to scrape the Omada controller in seconds",
		},
	)
)

// CollectMetrics scrapes the controller and exposes whatever it could collect.
// A failing device class does not abort the scrape (the other classes still
// report), but the failures are joined into the returned error so the caller can
// log them instead of silently serving a half-empty metric set.
func CollectMetrics() error {
	start := time.Now()
	deviceList, err := Devices.Get()
	if err != nil {
		return err
	}

	var (
		omadaDevices []Interface.Device
		errs         []error
	)

	switches, err := Switch.Get(*deviceList)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to get switches: %w", err))
	} else {
		for i := range *switches {
			(*switches)[i].AppendGeneralProperties(deviceList)
		}
		if err := Interface.AppendDevicesSlice(&omadaDevices, *switches); err != nil {
			errs = append(errs, err)
		}
	}

	gateways, err := Gateway.Get(*deviceList)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to get gateways: %w", err))
	} else {
		for i := range *gateways {
			(*gateways)[i].AppendGeneralProperties(deviceList)
		}
		if err := Interface.AppendDevicesSlice(&omadaDevices, *gateways); err != nil {
			errs = append(errs, err)
		}
	}

	aps, err := AccessPoint.Get(*deviceList)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to get access points: %w", err))
	} else {
		for i := range *aps {
			(*aps)[i].AppendGeneralProperties(deviceList)
		}
		if err := Interface.AppendDevicesSlice(&omadaDevices, *aps); err != nil {
			errs = append(errs, err)
		}
	}

	// Client statistics are opt-in, so the extra request is only made when the
	// operator asked for it.
	collectClientMetrics(internal.GetConfig().Prometheus.EnableClientMetrics)

	omada_scrape_duration.Set(time.Since(start).Seconds())

	ExposeDeviceMetrics(omadaDevices)
	ExposePortMetrics(omadaDevices)
	ExposeRadioMetrics(omadaDevices)

	return errors.Join(errs...)
}

// getClients is the collector's client source. It is a variable so the gating
// test can prove that a disabled ENABLE_CLIENT_METRICS never touches the
// clients endpoint, without a controller.
var getClients = Clients.Get

// collectClientMetrics fetches and exposes the optional per-client series. The
// enabled flag is passed in rather than read from the config so the scrape path
// and its tests can exercise the disabled branch without a config singleton.
func collectClientMetrics(enabled bool) {
	if !enabled {
		return
	}

	clients, err := getClients()
	if err != nil {
		// Client statistics are an optional extra: a failing request must not
		// cost the device, port and radio metrics of this scrape, so it is
		// logged and skipped instead of returned.
		Log.Warn("Failed to get clients, this scrape reports no client metrics: %s", err)
		return
	}

	ExposeClientMetrics(*clients)
}
