package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// Gin router with default middleware (logger and recovery).
	r := gin.Default()

	// HTTP request duration and status metrics for Prometheus.
	r.Use(PrometheusHTTPMiddleware())

	// Explicit 404 handler so unknown paths run through the same handler chain as routed traffic.
	r.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "404 page not found")
	})

	// Liveness and readiness probes for load balancers / orchestrators.
	r.GET("/health", HealthHandler)
	r.GET("/ready", ReadyHandler)

	// Task collection routes: list, create, and delete by id.
	r.GET("/tasks", GetTasks)
	r.POST("/tasks", CreateTask)
	r.DELETE("/tasks/:id", DeleteTask)

	// Metrics endpoint (e.g. for Prometheus scrapers).
	r.GET("/metrics", MetricsHandler)

	// Synthetic 500 for local metrics/Grafana demos (remove or gate in production).
	r.GET("/debug/simulate-500", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "simulated server error"})
	})

	// Start HTTP server on port 8080.
	_ = r.Run(":8080")
}
