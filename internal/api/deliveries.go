package api

import (
	"net/http"
	"time"

	"github.com/eventhook/eventhook/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *Server) handleListDeliveries(c *gin.Context) {
	f := store.ListDeliveriesFilter{}

	if v := c.Query("event_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event_id"})
			return
		}
		f.EventID = &id
	}
	if v := c.Query("endpoint_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid endpoint_id"})
			return
		}
		f.EndpointID = &id
	}
	if v := c.Query("status"); v != "" {
		f.Status = &v
	}

	deliveries, err := s.store.ListDeliveries(c.Request.Context(), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if deliveries == nil {
		deliveries = []*store.Delivery{}
	}
	c.JSON(http.StatusOK, gin.H{"data": deliveries})
}

func (s *Server) handleGetDelivery(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	d, err := s.store.GetDelivery(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	attempts, err := s.store.ListDeliveryAttempts(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if attempts == nil {
		attempts = []*store.DeliveryAttempt{}
	}

	c.JSON(http.StatusOK, gin.H{"delivery": d, "attempts": attempts})
}

func (s *Server) handleRetryDelivery(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	d, err := s.store.GetDelivery(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	now := time.Now()
	d.Status = "retrying"
	d.NextAttempt = now

	if err := s.store.UpdateDelivery(c.Request.Context(), d); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, d)
}
