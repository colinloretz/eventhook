package api

import (
	"net/http"

	"github.com/eventhook/eventhook/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type sourceRequest struct {
	Name       string         `json:"name" binding:"required"`
	Slug       string         `json:"slug" binding:"required"`
	Secret     *string        `json:"secret"`
	SourceType string         `json:"source_type" binding:"required,oneof=inbound outbound internal"`
	Metadata   map[string]any `json:"metadata"`
}

func (s *Server) handleListSources(c *gin.Context) {
	sources, err := s.store.ListSources(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if sources == nil {
		sources = []*store.Source{}
	}
	c.JSON(http.StatusOK, gin.H{"data": sources})
}

func (s *Server) handleCreateSource(c *gin.Context) {
	var req sourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Metadata == nil {
		req.Metadata = map[string]any{}
	}

	src := &store.Source{
		Name:       req.Name,
		Slug:       req.Slug,
		Secret:     req.Secret,
		SourceType: req.SourceType,
		Metadata:   req.Metadata,
	}
	if err := s.store.CreateSource(c.Request.Context(), src); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, src)
}

func (s *Server) handleGetSource(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	src, err := s.store.GetSource(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, src)
}

func (s *Server) handleUpdateSource(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	src, err := s.store.GetSource(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	var req sourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	src.Name = req.Name
	src.Slug = req.Slug
	src.Secret = req.Secret
	src.SourceType = req.SourceType
	if req.Metadata != nil {
		src.Metadata = req.Metadata
	}

	if err := s.store.UpdateSource(c.Request.Context(), src); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, src)
}

func (s *Server) handleDeleteSource(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := s.store.DeleteSource(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
