package ccm

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

var (
	Metrics = prometheus.NewRegistry()
	publish = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "ccm",
		Subsystem: "node",
		Name:      "publish_volume_success",
		Help:      "node publish volume success for cluster config maps",
	}, []string{"name"})
	publishTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "ccm",
		Subsystem: "node",
		Name:      "publish_volume_duration",
		Help:      "node publish volume duration for cluster config maps",
	}, []string{"name"})
	publishErr = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "ccm",
		Subsystem: "node",
		Name:      "publish_volume_error",
		Help:      "node publish volume errors for cluster config maps",
	}, []string{"name", "reason"})

	unpublish = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "ccm",
		Subsystem: "node",
		Name:      "unpublish_volume_success",
		Help:      "node unpublish volume success for cluster config maps",
	}, []string{"name"})
	unpublishTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "ccm",
		Subsystem: "node",
		Name:      "unpublish_volume_duration",
		Help:      "node unpublish volume duration for cluster config maps",
	}, []string{"name"})
	unpublishErr = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "ccm",
		Subsystem: "node",
		Name:      "unpublish_volume_error",
		Help:      "node unpublish volume errors for cluster config maps",
	}, []string{"name", "reason"})

	cleanup = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "ccm",
		Subsystem: "node",
		Name:      "cleanup_volume_success",
		Help:      "successful cleanup operations for data and metadata kept for cluster config maps",
	}, []string{"volumeId"})
	cleanupTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "ccm",
		Subsystem: "node",
		Name:      "cleanup_volume_duration",
		Help:      "time spent cleaning up data and metadata for cluster config maps",
	}, []string{})
	cleanupErr = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "ccm",
		Subsystem: "node",
		Name:      "cleanup_volume_error",
		Help:      "failed cleanup operations for data and metadata kept for cluster config maps",
	}, []string{"volumeId", "reason"})
)

func init() {
	Metrics.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	Metrics.MustRegister(collectors.NewGoCollector())
	Metrics.MustRegister(publish, publishTime, publishErr)
	Metrics.MustRegister(unpublish, unpublishTime, unpublishErr)
	Metrics.MustRegister(cleanup, cleanupTime, cleanupErr)
}
