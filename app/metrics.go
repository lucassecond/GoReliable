package main

import (
	"net/http"
	"strconv"
	"time"

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

// http_requests_total counts HTTP requests by method, path, and terminal status code.
var httpRequestsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests processed.",
	},
	[]string{"method", "path", "status"},
)

// http_request_duration_seconds measures request latency by method and path.
var httpRequestDurationSeconds = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration in seconds.",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"method", "path"},
)

// prometheusResponseWriter captures the status code written by handlers and middleware.
type prometheusResponseWriter struct {
	gin.ResponseWriter
	statusCode    int
	headerWritten bool
}

func (w *prometheusResponseWriter) WriteHeader(statusCode int) {
	if w.headerWritten {
		return
	}
	w.headerWritten = true
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *prometheusResponseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	// Gin applies the status on first body write (WriteHeaderNow); keep our copy in sync.
	w.statusCode = w.ResponseWriter.Status()
	if !w.headerWritten {
		w.headerWritten = true
	}
	return n, err
}

func (w *prometheusResponseWriter) WriteString(s string) (int, error) {
	n, err := w.ResponseWriter.WriteString(s)
	w.statusCode = w.ResponseWriter.Status()
	if !w.headerWritten {
		w.headerWritten = true
	}
	return n, err
}

func (w *prometheusResponseWriter) WriteHeaderNow() {
	w.ResponseWriter.WriteHeaderNow()
	w.statusCode = w.ResponseWriter.Status()
	w.headerWritten = true
}

func (w *prometheusResponseWriter) Flush() {
	w.ResponseWriter.Flush()
	w.statusCode = w.ResponseWriter.Status()
	if w.ResponseWriter.Written() {
		w.headerWritten = true
	}
}

// SetTaskCount sets the task gauge to n so scrapers see the live backlog size.
func SetTaskCount(n int) {
	tasksTotalGauge.Set(float64(n))
}

// recordPrometheusHTTPMetrics increments HTTP counters and histograms (single recording path).
func recordPrometheusHTTPMetrics(method, path string, statusCode int, elapsed time.Duration) {
	statusLabel := strconv.Itoa(statusCode)
	httpRequestsTotal.WithLabelValues(method, path, statusLabel).Inc()
	httpRequestDurationSeconds.WithLabelValues(method, path).Observe(elapsed.Seconds())
}

// PrometheusHTTPMiddleware records request counts and latencies for Prometheus scraping.
func PrometheusHTTPMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		pw := &prometheusResponseWriter{
			ResponseWriter: c.Writer,
			statusCode:     http.StatusOK,
		}
		c.Writer = pw

		c.Next()

		// Gin's serveError sets status on writermem before the chain runs, so prefer the real status
		// from the Gin response writer (covers 404/405 and any path that never calls WriteHeader).
		status := pw.ResponseWriter.Status()
		if status == 0 {
			status = pw.statusCode
		}
		recordPrometheusHTTPMetrics(c.Request.Method, c.Request.URL.Path, status, time.Since(start))
	}
}

// MetricsHandler exposes Prometheus metrics in text exposition format for scrapers.
func MetricsHandler(c *gin.Context) {
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}
