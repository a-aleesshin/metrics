package audit

import (
	"context"
	"time"

	sharedlogger "github.com/a-aleesshin/metrics/internal/shared/port/logger"
)

type Event struct {
	TS        int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

type Observer interface {
	Notify(ctx context.Context, event Event) error
}

type Publisher struct {
	observers []Observer
	logger    sharedlogger.Logger
	now       func() time.Time
}

func NewPublisher(logger sharedlogger.Logger, observers ...Observer) *Publisher {
	return &Publisher{
		observers: observers,
		logger:    logger,
		now:       time.Now,
	}
}

func (p *Publisher) Register(observer Observer) {
	if observer == nil {
		return
	}

	p.observers = append(p.observers, observer)
}

func (p *Publisher) Publish(ctx context.Context, metricNames []string, ipAddress string) {
	if len(p.observers) == 0 {
		return
	}

	event := Event{
		TS:        p.now().Unix(),
		Metrics:   metricNames,
		IPAddress: ipAddress,
	}

	for _, observer := range p.observers {
		if err := observer.Notify(ctx, event); err != nil && p.logger != nil {
			p.logger.Error("audit notify failed", sharedlogger.Err(err))
		}
	}
}
