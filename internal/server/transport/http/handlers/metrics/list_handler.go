package metrics

import (
	"context"
	"html/template"
	"net/http"

	"github.com/a-aleesshin/metrics/internal/server/application/dto"
)

// ListMetricsUseCase — usecase получения списка всех метрик.
type ListMetricsUseCase interface {
	Execute(ctx context.Context) (dto.ListMetricsResult, error)
}

const listPage = `
<!doctype html>
<html>
	<body>
		<h1>Metrics</h1>

		<ul>{{range .Items}}
			<li>{{.Type}} {{.Name}} = {{.Value}}</li>{{end}}
		</ul>
	</body>
</html>`

var listPageTemplate = template.Must(template.New("metrics").Parse(listPage))

// ListMetricsHandler отдаёт HTML-страницу со списком всех метрик.
type ListMetricsHandler struct {
	listMetric ListMetricsUseCase
}

// NewListMetricsHandler создаёт хендлер HTML-списка метрик.
func NewListMetricsHandler(listMetric ListMetricsUseCase) *ListMetricsHandler {
	return &ListMetricsHandler{listMetric: listMetric}
}

// List обрабатывает GET /: возвращает 200 и HTML-страницу со списком метрик,
// 500 — при ошибке usecase.
func (h *ListMetricsHandler) List(w http.ResponseWriter, r *http.Request) {
	result, err := h.listMetric.Execute(r.Context())

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	_ = listPageTemplate.Execute(w, result)
}
