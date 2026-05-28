package api

import (
	"context"
	"net/http"
	"time"

	"github.com/eventhook/eventhook/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type createEventRequest struct {
	EventType      string         `json:"event_type" binding:"required"`
	Payload        map[string]any `json:"payload" binding:"required"`
	SourceID       *uuid.UUID     `json:"source_id"`
	IdempotencyKey *string        `json:"idempotency_key"`
	Headers        map[string]any `json:"headers"`
}

func (s *Server) handleCreateEvent(c *gin.Context) {
	var req createEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Headers == nil {
		req.Headers = map[string]any{}
	}

	ev := &store.Event{
		EventType:      req.EventType,
		Payload:        req.Payload,
		SourceID:       req.SourceID,
		IdempotencyKey: req.IdempotencyKey,
		Headers:        req.Headers,
		Status:         "pending",
	}

	if err := s.store.CreateEvent(c.Request.Context(), ev); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	go s.fanOutDeliveries(ev)

	c.JSON(http.StatusCreated, ev)
}

func (s *Server) handleListEvents(c *gin.Context) {
	f := store.ListEventsFilter{}

	if v := c.Query("source_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid source_id"})
			return
		}
		f.SourceID = &id
	}
	if v := c.Query("event_type"); v != "" {
		f.EventType = &v
	}
	if v := c.Query("status"); v != "" {
		f.Status = &v
	}

	events, err := s.store.ListEvents(c.Request.Context(), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if events == nil {
		events = []*store.Event{}
	}
	c.JSON(http.StatusOK, gin.H{"data": events})
}

func (s *Server) handleGetEvent(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	ev, err := s.store.GetEvent(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, ev)
}

func (s *Server) handleReplayEvent(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	orig, err := s.store.GetEvent(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	replay := &store.Event{
		SourceID:  orig.SourceID,
		EventType: orig.EventType,
		Payload:   orig.Payload,
		Headers:   orig.Headers,
		Status:    "pending",
	}
	if err := s.store.CreateEvent(c.Request.Context(), replay); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	go s.fanOutDeliveries(replay)

	c.JSON(http.StatusCreated, replay)
}

// fanOutDeliveries creates a pending delivery row for every enabled endpoint
// that subscribes to the event type.
func (s *Server) fanOutDeliveries(ev *store.Event) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	endpoints, err := s.store.ListEndpoints(ctx)
	if err != nil {
		log.Error().Err(err).Str("event_id", ev.ID.String()).Msg("fan-out: list endpoints")
		return
	}

	for _, ep := range endpoints {
		if !ep.Enabled {
			continue
		}
		if !endpointMatchesEvent(ep, ev.EventType) {
			continue
		}
		d := &store.Delivery{
			EventID:     ev.ID,
			EndpointID:  ep.ID,
			Status:      "pending",
			NextAttempt: time.Now(),
		}
		if err := s.store.CreateDelivery(ctx, d); err != nil {
			log.Error().Err(err).
				Str("event_id", ev.ID.String()).
				Str("endpoint_id", ep.ID.String()).
				Msg("fan-out: create delivery")
		}
	}
}

func endpointMatchesEvent(ep *store.Endpoint, eventType string) bool {
	if len(ep.EventTypes) == 0 {
		return true // wildcard: receive all events
	}
	for _, t := range ep.EventTypes {
		if t == eventType || t == "*" {
			return true
		}
	}
	return false
}
