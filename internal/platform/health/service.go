package health

import "context"

type CheckResult struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type Report struct {
	Status string        `json:"status"`
	Checks []CheckResult `json:"checks"`
}

type Service struct {
	checks []Checker
}

func NewService(checks ...Checker) *Service {
	return &Service{
		checks: checks,
	}
}

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
