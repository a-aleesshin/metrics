package metric

import "errors"

var (
	// ErrNameEmpty — имя метрики пустое.
	ErrNameEmpty = errors.New("name is empty")

	// ErrUnsupportedMetricType — тип метрики не поддерживается (не gauge и не counter).
	ErrUnsupportedMetricType = errors.New("unsupport metric type")
	// ErrInvalidMetricType — некорректный тип метрики.
	ErrInvalidMetricType = errors.New("invalid metric type")
	// ErrInvalidMetricValue — значение метрики отсутствует или не парсится.
	ErrInvalidMetricValue = errors.New("invalid metric value")

	// ErrIDEmpty — идентификатор метрики пустой.
	ErrIDEmpty = errors.New("id is empty")
)
