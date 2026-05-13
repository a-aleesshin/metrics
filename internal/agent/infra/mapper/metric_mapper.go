package mapper

import (
	"fmt"
	"math"
	"strconv"

	appdto "github.com/a-aleesshin/metrics/internal/agent/application/dto"
	httpdto "github.com/a-aleesshin/metrics/internal/agent/infra/dto"
)

func ToSendMetric(dto appdto.MetricDTO) (httpdto.MetricsSend, error) {
	payload := httpdto.MetricsSend{
		ID:    dto.Name,
		MType: dto.Type,
	}

	switch dto.Type {
	case "gauge":
		v, err := strconv.ParseFloat(dto.Value, 64)
		if err != nil {
			return httpdto.MetricsSend{}, fmt.Errorf("invalid gauge value %q: %w", dto.Value, err)
		}

		if math.IsNaN(v) || math.IsInf(v, 0) {
			v = 0
		}

		payload.Value = &v

	case "counter":
		d, err := strconv.ParseInt(dto.Value, 10, 64)
		if err != nil {
			return httpdto.MetricsSend{}, fmt.Errorf("invalid counter value %q: %w", dto.Value, err)
		}

		payload.Delta = &d

	default:
		return httpdto.MetricsSend{}, fmt.Errorf("unsupported metric type: %s", dto.Type)
	}

	return payload, nil
}
