package api

import (
	"net/http"

	"github.com/eventhook/eventhook/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type endpointRequest struct {
	URL         string         `json:"url" binding:"required,url"`
	Description *string        `json:"description"`
	Secret      string         `json:"secret" binding:"required"`
	Enabled     *bool          `json:"enabled"`
	EventTypes  []string       `json:"event_types"`
	Metadata    map[string]any `json:"metadata"`
}

func (s *Server) handleListEndpoints(c *gin.Context) {
	endpoints, err := s.store.ListEndpoints(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if endpoints == nil {
		endpoints = []*store.Endpoint{}
	}
	c.JSON(http.StatusOK, gin.H{"data": endpoints})
}

func (s *Server) handleCreateEndpoint(c *gin.Context) {
	var req endpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	if req.Metadata == nil {
		req.Metadata = map[string]any{}
	}

	ep := &store.Endpoint{
		URL:         req.URL,
		Description: req.Description,
		Secret:      req.Secret,
		Enabled:     enabled,
		EventTypes:  req.EventTypes,
		Metadata:    req.Metadata,
	}
	if err := s.store.CreateEndpoint(c.Request.Context(), ep); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, ep)
}

func (s *Server) handleGetEndpoint(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	ep, err := s.store.GetEndpoint(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, ep)
}

func (s *Server) handleUpdateEndpoint(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	ep, err := s.store.GetEndpoint(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	var req endpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ep.URL = req.URL
	ep.Description = req.Description
	ep.Secret = req.Secret
	if req.Enabled != nil {
		ep.Enabled = *req.Enabled
	}
	ep.EventTypes = req.EventTypes
	if req.Metadata != nil {
		ep.Metadata = req.Metadata
	}

	if err := s.store.UpdateEndpoint(c.Request.Context(), ep); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ep)
}

func (s *Server) handleDeleteEndpoint(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := s.store.DeleteEndpoint(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
