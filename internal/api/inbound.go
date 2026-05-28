package api

import (
	"io"
	"net/http"

	"github.com/eventhook/eventhook/internal/store"
	"github.com/gin-gonic/gin"
)

func (s *Server) handleInbound(c *gin.Context) {
	slug := c.Param("source_slug")

	src, err := s.store.GetSourceBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "unknown source"})
		return
	}

	body, err := io.ReadAll(io.LimitReader(c.Request.Body, int64(s.cfg.MaxPayloadKB)*1024))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	headers := map[string]any{}
	for k, v := range c.Request.Header {
		if len(v) == 1 {
			headers[k] = v[0]
		} else {
			headers[k] = v
		}
	}

	eventType := c.GetHeader("X-Event-Type")
	if eventType == "" {
		eventType = "inbound." + slug
	}

	var payload map[string]any
	if err := c.ShouldBindJSON(&payload); err != nil {
		// Store raw body under "raw" key if not valid JSON
		payload = map[string]any{"raw": string(body)}
	}

	ev := &store.Event{
		SourceID:  &src.ID,
		EventType: eventType,
		Payload:   payload,
		Headers:   headers,
		Status:    "pending",
	}

	if idempKey := c.GetHeader("X-Idempotency-Key"); idempKey != "" {
		ev.IdempotencyKey = &idempKey
	}

	if err := s.store.CreateEvent(c.Request.Context(), ev); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	go s.fanOutDeliveries(ev)

	c.JSON(http.StatusOK, gin.H{"id": ev.ID, "status": "accepted"})
}
