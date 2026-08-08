// Package bench содержит бенчмарки важнейших компонентов агента:
// сбор runtime-метрик, построение батча для отправки и полный путь
// отправки (маппинг + JSON + gzip + подпись).
//
// Снятие профиля памяти (из корня репозитория):
//
//	go test -run='^$' -bench=. -benchmem -memprofile=profiles/agent_base.pprof ./internal/agent/bench
package bench
