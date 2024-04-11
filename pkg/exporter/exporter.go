package exporter

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Exporter struct {
	guageVecs       map[string]*prometheus.GaugeVec
	MetricsMetaData MetricsMetaData
}

type metricName string
type metricNamespace string
type metricSubsystem string
type MetricsMetaData map[metricNamespace]map[metricSubsystem]map[metricName][]string

func New(metricsMetaData MetricsMetaData) *Exporter {
	var exporter = Exporter{
		guageVecs:       make(map[string]*prometheus.GaugeVec),
		MetricsMetaData: metricsMetaData,
	}
	for namespace, subsystems := range metricsMetaData {
		for subsystem, names := range subsystems {
			for name, lables := range names {
				exporter.guageVecs[string(name)] = prometheus.NewGaugeVec(
					prometheus.GaugeOpts{
						Namespace: string(namespace),
						Subsystem: string(subsystem),
						Name:      string(name),
						Help:      "",
					}, lables)
			}
		}
	}
	return &exporter
}
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, vec := range e.guageVecs {
		vec.Describe(ch)
	}
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	for _, vec := range e.guageVecs {
		vec.Collect(ch)
	}
}

func (e *Exporter) Start(addr string) {
	prometheus.MustRegister(e)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(addr, nil))
}

func (e *Exporter) WriteMetrics(namespace, subsystem, name string, value float64, labelValues ...string) {
	e.guageVecs[name].WithLabelValues(labelValues...).Set(value)
}

func (e *Exporter) DeleteMetrics(namespace, subsystem, name string, labels prometheus.Labels) bool {
	return e.guageVecs[name].Delete(labels)
}
