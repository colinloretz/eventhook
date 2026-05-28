package worker

import (
	"context"
	"sync"
	"time"

	"github.com/eventhook/eventhook/internal/store"
	"github.com/rs/zerolog/log"
)

const pollInterval = 2 * time.Second

type Pool struct {
	store       store.Store
	workerCount int
	jobs        chan *store.Delivery
	wg          sync.WaitGroup
}

func NewPool(st store.Store, workerCount int) *Pool {
	return &Pool{
		store:       st,
		workerCount: workerCount,
		jobs:        make(chan *store.Delivery, workerCount*2),
	}
}

func (p *Pool) Start(ctx context.Context) {
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.runWorker(ctx)
	}
	go p.runScheduler(ctx)
}

func (p *Pool) Stop() {
	close(p.jobs)
	p.wg.Wait()
}

func (p *Pool) runWorker(ctx context.Context) {
	defer p.wg.Done()
	for d := range p.jobs {
		if err := DeliverOne(ctx, p.store, d); err != nil {
			log.Error().Err(err).Str("delivery_id", d.ID.String()).Msg("delivery failed")
		}
	}
}

func (p *Pool) runScheduler(ctx context.Context) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.poll(ctx)
		}
	}
}

func (p *Pool) poll(ctx context.Context) {
	deliveries, err := p.store.PollPendingDeliveries(ctx, p.workerCount)
	if err != nil {
		log.Error().Err(err).Msg("poll deliveries")
		return
	}
	for _, d := range deliveries {
		select {
		case p.jobs <- d:
		case <-ctx.Done():
			return
		}
	}
}
