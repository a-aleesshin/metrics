package metrics

import (
	"html/template"
	"net/http"

	"github.com/a-aleesshin/metrics/internal/server/application/dto"
)

type ListMetricsUseCase interface {
	Execute() (dto.ListMetricsResult, error)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	result, err := h.listMetric.Execute()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	const page = `
<!doctype html>
<html>
	<body>
		<h1>Metrics</h1>

		<ul>{{range .Items}}
			<li>{{.Type}} {{.Name}} = {{.Value}}</li>{{end}}
		</ul>
	</body>
</html>`

	tpl := template.Must(template.New("metrics").Parse(page))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	_ = tpl.Execute(w, result)
}
