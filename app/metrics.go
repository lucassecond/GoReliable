package main

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// goreliable_tasks_total reflects how many tasks currently live in the in-memory store.
var tasksTotalGauge = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "goreliable_tasks_total",
	Help: "Current number of tasks stored in memory.",
})

// goreliable_http_requests_total counts each HTTP request, broken down by method and path.
var httpRequestsTotalCounter = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "goreliable_http_requests_total",
		Help: "Total HTTP requests received, labeled by method and path.",
	},
	[]string{"method", "path"},
)

// SetTaskCount sets the task gauge to n so scrapers see the live backlog size.
func SetTaskCount(n int) {
	tasksTotalGauge.Set(float64(n))
}

// RecordRequest increments the labeled HTTP request counter for one observed request.
func RecordRequest(method, path string) {
	httpRequestsTotalCounter.WithLabelValues(method, path).Inc()
}

// MetricsHandler exposes Prometheus metrics in text exposition format for scrapers.
func MetricsHandler(c *gin.Context) {
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}
