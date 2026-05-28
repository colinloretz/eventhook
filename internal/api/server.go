package api

import (
	"fmt"
	"net/http"

	"github.com/eventhook/eventhook/internal/config"
	"github.com/eventhook/eventhook/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type Server struct {
	cfg    *config.Config
	store  store.Store
	router *gin.Engine
}

func NewServer(cfg *config.Config, st store.Store) *Server {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	s := &Server{cfg: cfg, store: st}
	s.router = gin.New()
	s.router.Use(gin.Recovery())
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.router.GET("/health", s.handleHealth)

	api := s.router.Group("/api/v1", authMiddleware(s.cfg.APIKey))

	api.POST("/events", s.handleCreateEvent)
	api.GET("/events", s.handleListEvents)
	api.GET("/events/:id", s.handleGetEvent)
	api.POST("/events/:id/replay", s.handleReplayEvent)

	api.GET("/deliveries", s.handleListDeliveries)
	api.GET("/deliveries/:id", s.handleGetDelivery)
	api.POST("/deliveries/:id/retry", s.handleRetryDelivery)

	api.GET("/endpoints", s.handleListEndpoints)
	api.POST("/endpoints", s.handleCreateEndpoint)
	api.GET("/endpoints/:id", s.handleGetEndpoint)
	api.PUT("/endpoints/:id", s.handleUpdateEndpoint)
	api.DELETE("/endpoints/:id", s.handleDeleteEndpoint)

	api.GET("/sources", s.handleListSources)
	api.POST("/sources", s.handleCreateSource)
	api.GET("/sources/:id", s.handleGetSource)
	api.PUT("/sources/:id", s.handleUpdateSource)
	api.DELETE("/sources/:id", s.handleDeleteSource)

	// Inbound webhook receiver (no auth — uses source secret for verification)
	s.router.POST("/api/v1/in/:source_slug", s.handleInbound)
}

func (s *Server) Run() error {
	addr := fmt.Sprintf(":%d", s.cfg.Port)
	log.Info().Str("addr", addr).Msg("starting server")
	return s.router.Run(addr)
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
