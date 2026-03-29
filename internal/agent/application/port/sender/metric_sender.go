package sender

import "github.com/a-aleesshin/metrics/internal/agent/application/dto"

type MetricSender interface {
	Send(dto dto.MetricDTO) error
}
