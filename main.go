/*
Copyright 2026 Thomas Helander

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"fmt"
	"net/http"
	"os"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/thelande/liquidctl-exporter/pkg/collector"

	"github.com/prometheus/client_golang/prometheus"
	versioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/promslog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	"github.com/prometheus/exporter-toolkit/web/kingpinflag"
)

const (
	exporterName  = "liquidctl-exporter"
	exporterTitle = "Prometheus exporter for liquidctl"
)

var (
	metricsPath = kingpin.Flag(
		"web.telemetry-path",
		"Path under which to expose metrics.",
	).Default("/metrics").String()
	toolkitFlags = kingpinflag.AddFlags(kingpin.CommandLine, ":9530")
)

func main() {
	promslogConfig := &promslog.Config{}
	flag.AddFlags(kingpin.CommandLine, promslogConfig)
	kingpin.CommandLine.UsageWriter(os.Stdout)
	kingpin.HelpFlag.Short('h')
	kingpin.Version(version.Print(exporterName))
	kingpin.Parse()

	logger := promslog.New(promslogConfig)
	logger.Info(fmt.Sprintf("Starting %s", exporterName), "version", version.Info())
	logger.Info("Build context", "build_context", version.BuildContext())

	collector := collector.NewCollector(logger)

	// Uncomment the following three lines and comment out prometheus.MustRegister(...)
	// to exclude the go metrics. Make sure to swap line 89 and 90 as well.
	// registry := prometheus.NewRegistry()
	// registry.MustRegister(versioncollector.NewCollector(exporterName))
	// registry.MustRegister(collector)
	prometheus.MustRegister(collector)
	prometheus.MustRegister(versioncollector.NewCollector(exporterName))

	landingConfig := web.LandingConfig{
		Name:        exporterTitle,
		Description: "Prometheus Exporter for liquidctl",
		Version:     version.Info(),
		Links: []web.LandingLinks{
			{
				Address: *metricsPath,
				Text:    "Metrics",
			},
		},
		Profiling: "false",
	}
	landingPage, err := web.NewLandingPage(landingConfig)
	if err != nil {
		logger.Error(fmt.Sprintf("%s", err))
		os.Exit(1)
	}

	// http.Handle(*metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	http.Handle(*metricsPath, promhttp.Handler())
	http.Handle("/", landingPage)

	srv := &http.Server{}
	if err := web.ListenAndServe(srv, toolkitFlags, logger); err != nil {
		logger.Error("HTTP listener stopped", "error", err)
		os.Exit(1)
	}
}
