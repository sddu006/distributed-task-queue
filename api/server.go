package api

import (
	"fmt"
	"net/http"
	"task-queue/queue"
	"task-queue/store"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	store       *store.Store
	queue       *queue.Queue
	router      *gin.Engine
	rateLimiter *RateLimiter
}

func NewServer(s *store.Store, q *queue.Queue) *Server {
	server := &Server{
		store:       s,
		queue:       q,
		router:      gin.Default(),
		rateLimiter: NewRateLimiter(1, 5), // 1 token/sec, max 5 tokens
	}
	server.setupRoutes()
	return server
}

func (s *Server) setupRoutes() {
	s.router.Use(s.rateLimiter.Middleware())
	s.router.POST("/jobs", s.submitJob)
	s.router.GET("/jobs", s.listJobs)
	s.router.GET("/jobs/:id", s.getJob)
	s.router.GET("/dashboard", s.serveDashboard)
}

func (s *Server) Run(port string) {
	fmt.Printf("API Server running on port %s\n", port)
	s.router.Run(":" + port)
}

func (s *Server) submitJob(c *gin.Context) {
	var request struct {
		Type    string `json:"type"`
		Payload string `json:"payload"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	if request.Type == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "job type is required",
		})
		return
	}

	job := &store.Job{
		ID:        fmt.Sprintf("job_%d", time.Now().UnixNano()),
		Type:      request.Type,
		Payload:   request.Payload,
		Status:    store.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Retries:   0,
	}

	s.store.Save(job)
	s.queue.Push(job)

	c.JSON(http.StatusCreated, job)
}

func (s *Server) listJobs(c *gin.Context) {
	jobs := s.store.GetAll()
	c.JSON(http.StatusOK, jobs)
}

func (s *Server) getJob(c *gin.Context) {
	id := c.Param("id")
	job, exists := s.store.Get(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "job not found",
		})
		return
	}
	c.JSON(http.StatusOK, job)
}

func (s *Server) serveDashboard(c *gin.Context) {
	c.File("dashboard/index.html")
}
