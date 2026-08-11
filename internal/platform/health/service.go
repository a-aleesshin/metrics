// Package health агрегирует проверки здоровья компонентов в единый отчёт.
package health

import "context"

// CheckResult — результат одной проверки: имя, статус и текст ошибки при сбое.
type CheckResult struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// Report — сводный отчёт: общий статус и результаты всех проверок.
type Report struct {
	Status string        `json:"status"`
	Checks []CheckResult `json:"checks"`
}

// Service последовательно выполняет зарегистрированные проверки здоровья.
type Service struct {
	checks []Checker
}

// NewService создаёт сервис health-проверок из переданных Checker.
func NewService(checks ...Checker) *Service {
	return &Service{
		checks: checks,
	}
}

// Check выполняет все проверки и возвращает отчёт;
// общий статус — "unhealthy", если хотя бы одна проверка упала.
func (service *Service) Check(ctx context.Context) Report {
	status := "ok"
	results := make([]CheckResult, 0, len(service.checks))

	for _, check := range service.checks {
		result := CheckResult{Name: check.Name()}

		if err := check.Check(ctx); err != nil {
			result.Status = "error"
			result.Error = err.Error()
			status = "unhealthy"
		} else {
			result.Status = "ok"
		}

		results = append(results, result)
	}

	return Report{Status: status, Checks: results}
}
