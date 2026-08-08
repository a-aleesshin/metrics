// Package audit реализует паттерн «Наблюдатель» для аудита запросов на изменение метрик:
// Publisher рассылает события Event зарегистрированным приёмникам Observer.
package audit

import (
	"context"
	"time"

	sharedlogger "github.com/a-aleesshin/metrics/internal/shared/port/logger"
)

// Event — событие аудита: unix-время, имена изменённых метрик и IP-адрес клиента.
type Event struct {
	TS        int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

// Observer — приёмник событий аудита. Notify доставляет событие; ошибка доставки
// логируется издателем и не прерывает рассылку остальным приёмникам.
type Observer interface {
	Notify(ctx context.Context, event Event) error
}

// Publisher рассылает события аудита зарегистрированным наблюдателям.
type Publisher struct {
	observers []Observer
	logger    sharedlogger.Logger
	now       func() time.Time
}

// NewPublisher создаёт издателя с заданным логгером и начальным набором наблюдателей.
func NewPublisher(logger sharedlogger.Logger, observers ...Observer) *Publisher {
	return &Publisher{
		observers: observers,
		logger:    logger,
		now:       time.Now,
	}
}

// Register добавляет наблюдателя; nil игнорируется.
func (p *Publisher) Register(observer Observer) {
	if observer == nil {
		return
	}

	p.observers = append(p.observers, observer)
}

// Publish формирует событие с текущим временем и рассылает его всем наблюдателям.
// Ошибки доставки логируются и не прерывают рассылку.
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
