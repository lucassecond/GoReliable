package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	// Gin router with default middleware (logger and recovery).
	r := gin.Default()

	// HTTP request duration and status metrics for Prometheus.
	r.Use(PrometheusHTTPMiddleware())

	// Liveness and readiness probes for load balancers / orchestrators.
	r.GET("/health", HealthHandler)
	r.GET("/ready", ReadyHandler)

	// Task collection routes: list, create, and delete by id.
	r.GET("/tasks", GetTasks)
	r.POST("/tasks", CreateTask)
	r.DELETE("/tasks/:id", DeleteTask)

	// Metrics endpoint (e.g. for Prometheus scrapers).
	r.GET("/metrics", MetricsHandler)

	// Start HTTP server on port 8080.
	_ = r.Run(":8080")
}
