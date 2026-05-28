package worker

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/eventhook/eventhook/internal/store"
)

const deliveryTimeout = 30 * time.Second

func deliver(ctx context.Context, st store.Store, d *store.Delivery, ep *store.Endpoint, ev *store.Event) error {
	body, err := json.Marshal(ev.Payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	sig := signPayload(ep.Secret, body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ep.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-EventHook-Signature", "sha256="+sig)
	req.Header.Set("X-EventHook-Event", ev.EventType)
	req.Header.Set("X-EventHook-Delivery", d.ID.String())
	req.Header.Set("X-EventHook-Event-ID", ev.ID.String())

	reqHeaders := map[string]any{}
	for k, v := range req.Header {
		reqHeaders[k] = v
	}
	reqBody := string(body)

	start := time.Now()
	client := &http.Client{Timeout: deliveryTimeout}
	resp, err := client.Do(req)
	latencyMS := int(time.Since(start).Milliseconds())

	attempt := &store.DeliveryAttempt{
		DeliveryID:     d.ID,
		Attempt:        d.AttemptCount + 1,
		RequestHeaders: reqHeaders,
		RequestBody:    &reqBody,
		LatencyMS:      &latencyMS,
	}

	now := time.Now()
	d.AttemptCount++
	d.LastAttemptedAt = &now

	if err != nil {
		timeoutStatus := "timeout"
		attempt.Status = timeoutStatus
		_ = st.CreateDeliveryAttempt(ctx, attempt)
		return scheduleRetry(ctx, st, d, ev)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	respBodyStr := string(respBody)
	respHeaders := map[string]any{}
	for k, v := range resp.Header {
		respHeaders[k] = v
	}

	attempt.ResponseStatus = &resp.StatusCode
	attempt.ResponseHeaders = respHeaders
	attempt.ResponseBody = &respBodyStr
	d.LastResponseStatus = &resp.StatusCode
	d.LastResponseBody = &respBodyStr

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		attempt.Status = "success"
		d.Status = "delivered"
		d.DeliveredAt = &now
		_ = st.UpdateEventStatus(ctx, ev.ID, "delivered")
	} else {
		attempt.Status = "failure"
		_ = scheduleRetry(ctx, st, d, ev)
	}

	_ = st.CreateDeliveryAttempt(ctx, attempt)
	return st.UpdateDelivery(ctx, d)
}

func scheduleRetry(ctx context.Context, st store.Store, d *store.Delivery, ev *store.Event) error {
	delay := NextBackoff(d.AttemptCount)
	if delay < 0 {
		d.Status = "failed"
		_ = st.UpdateEventStatus(ctx, ev.ID, "failed")
	} else {
		d.Status = "retrying"
		d.NextAttempt = time.Now().Add(delay)
	}
	return st.UpdateDelivery(ctx, d)
}

func signPayload(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

// DeliverOne loads the event and endpoint for a delivery and executes delivery.
func DeliverOne(ctx context.Context, st store.Store, d *store.Delivery) error {
	ev, err := st.GetEvent(ctx, d.EventID)
	if err != nil {
		return fmt.Errorf("get event %s: %w", d.EventID, err)
	}
	ep, err := st.GetEndpoint(ctx, d.EndpointID)
	if err != nil {
		return fmt.Errorf("get endpoint %s: %w", d.EndpointID, err)
	}
	return deliver(ctx, st, d, ep, ev)
}

