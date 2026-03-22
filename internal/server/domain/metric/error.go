package metric

import "errors"

var (
	ErrNameEmpty = errors.New("name is empty")

	ErrUnsupportedMetricType = errors.New("unsupport metric type")
	ErrInvalidMetricType     = errors.New("invalid metric type")
	ErrInvalidMetricValue    = errors.New("invalid metric value")

	ErrIDEmpty = errors.New("id is empty")
)
