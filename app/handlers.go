package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// Task represents a single todo item tracked by the API.
type Task struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

var (
	tasks   []Task
	tasksMu sync.RWMutex
)

// HealthHandler answers liveness checks with a simple healthy status.
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

// ReadyHandler answers readiness checks; here it always reports ready.
func ReadyHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

// GetTasks returns every task currently stored in memory.
func GetTasks(c *gin.Context) {
	tasksMu.RLock()
	n := len(tasks)
	out := make([]Task, n)
	copy(out, tasks)
	tasksMu.RUnlock()
	SetTaskCount(n)
	c.JSON(http.StatusOK, out)
}

// CreateTask parses a JSON body with a title, assigns a random ID, appends the task, and returns it.
func CreateTask(c *gin.Context) {
	var body struct {
		Title string `json:"title"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	// validate first, generate ID only after
	if body.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}

	id := fmt.Sprintf("%016x%016x", rand.Uint64(), rand.Uint64())
	task := Task{ID: id, Title: body.Title, Done: false}

	tasksMu.Lock()
	tasks = append(tasks, task)
	SetTaskCount(len(tasks))
	tasksMu.Unlock()

	c.JSON(http.StatusCreated, task)
}

// DeleteTask removes the task whose ID matches the :id path parameter.
func DeleteTask(c *gin.Context) {
	id := c.Param("id")

	tasksMu.Lock()
	defer tasksMu.Unlock()
	for i := range tasks {
		if tasks[i].ID == id {
			tasks = append(tasks[:i], tasks[i+1:]...)
			SetTaskCount(len(tasks))
			c.Status(http.StatusNoContent)
			return
		}
	}
	SetTaskCount(len(tasks))
	c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
}
