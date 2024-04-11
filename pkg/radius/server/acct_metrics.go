package server

import (
	"fmt"
	"radius-server/pkg/exporter"
)

var ocservMetricsMeta = exporter.MetricsMetaData{
	"ocserv": {
		"user_session": {
			"status":          {"id", "remote_address", "username", "nasidentifier", "framed_address"},
			"uptimes":         {"id", "remote_address", "username", "nasidentifier", "framed_address"},
			"output_bytes":    {"id", "remote_address", "username", "nasidentifier", "framed_address"},
			"input_bytes":     {"id", "remote_address", "username", "nasidentifier", "framed_address"},
			"terminate_cause": {"id", "remote_address", "username", "nasidentifier", "framed_address"},
		},
	},
}

func radiusExporter(metricsAddr string) *exporter.Exporter {
	fmt.Println("metrics server on:", metricsAddr)
	exp := exporter.New(ocservMetricsMeta)
	go exp.Start(metricsAddr)
	return exp
}
