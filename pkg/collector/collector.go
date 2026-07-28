package collector

import (
	"fmt"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/thelande/liquidctl-exporter/pkg/liquidctl"
)

const namespace = "liquidctl"

var baseLabels = []string{"bus", "address"}
var infoLabels = append(baseLabels, []string{"description", "vendor_id", "product_id", "release_number", "serial_number", "driver"}...)
var attributeLabels = append(baseLabels, []string{"key", "value", "unit"}...)
var metricLabels = append(baseLabels, []string{"key"}...)

var (
	upDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "up"),
		"Was the exporter able to run liquidctl and collect metrics?",
		nil,
		nil,
	)

	infoDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "meta", "info"),
		"Information about the device.",
		infoLabels,
		nil,
	)

	attributesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "attributes", "info"),
		"The device attributes.",
		attributeLabels,
		nil,
	)

	fanSpeedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "fan_speed", "rpm"),
		"The speed of the fans connected to the controller.",
		metricLabels,
		nil,
	)

	temperatureDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "temperature", "celsius"),
		"The temperature of the sensors connected to the controller.",
		metricLabels,
		nil,
	)
)

type Collector struct {
	Logger    *slog.Logger
	LiquidCtl *liquidctl.LiquidCtl
}

func NewCollector(Logger *slog.Logger) Collector {
	return Collector{
		Logger:    Logger,
		LiquidCtl: liquidctl.NewLiquidCtl(),
	}
}

func (c Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- upDesc
	ch <- infoDesc
	ch <- attributesDesc
	ch <- fanSpeedDesc
	ch <- temperatureDesc
}

func (c Collector) Collect(ch chan<- prometheus.Metric) {
	devices, err := c.LiquidCtl.ListDevices()
	if err != nil {
		ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, 0)
		c.Logger.Warn("failed to collect devices from liquidctl", "err", err)
		return
	}

	// Made it this far, so we must have talked to liquidctl okay.
	ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, 1)

	for i := range devices {
		device := devices[i]
		c.Logger.Debug(
			"collecting metrics for device",
			"bus", device.Bus,
			"address", device.Address,
			"description", device.Description,
		)

		ch <- prometheus.MustNewConstMetric(
			infoDesc,
			prometheus.GaugeValue,
			1,
			device.Bus,
			device.Address,
			device.Description,
			fmt.Sprintf("0x%x", device.VendorID),
			fmt.Sprintf("0x%x", device.ProductID),
			fmt.Sprintf("0x%x", device.ReleaseNumber),
			device.SerialNumber,
			device.Driver,
		)

		// Collect detailed attributes
		attributes, err := c.LiquidCtl.GetAttributes(device.Bus, device.Address)
		if err != nil {
			c.Logger.Warn("failed to collect device attributes", "bus", device.Bus, "address", device.Address, "err", err)
			continue
		}

		for j := range attributes.Status {
			status := attributes.Status[j]
			ch <- prometheus.MustNewConstMetric(
				attributesDesc,
				prometheus.GaugeValue,
				1,
				device.Bus,
				device.Address,
				status.Key,
				status.Value.ToString(),
				status.Unit,
			)
		}

		// Collect the sensors
		details, err := c.LiquidCtl.GetDevice(device.Bus, device.Address)
		if err != nil {
			c.Logger.Warn("failed to collect device details", "bus", device.Bus, "address", device.Address, "err", err)
			continue
		}

		for j := range details.Status {
			status := details.Status[j]
			var desc *prometheus.Desc
			var value float64
			labels := []string{device.Bus, device.Address, status.Key}

			switch status.Unit {
			case "°C":
				desc = temperatureDesc
				value = *status.Value.Float
			case "rpm":
				desc = fanSpeedDesc
				value = float64(*status.Value.Integer)
			default:
				c.Logger.Warn("unknown unit found", "bus", device.Bus, "address", device.Address, "key", status.Key, "unit", status.Unit, "value", status.Value.ToString())
				continue
			}

			ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, value, labels...)
		}
	}
}
