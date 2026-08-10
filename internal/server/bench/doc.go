// Package bench содержит бенчмарки важнейших компонентов сервера:
// полный HTTP-конвейер (middleware подписи и сжатия + хендлеры + in-memory
// хранилище) и отдельные компоненты — batch-обновление хранилища и издатель
// аудита.
//
// Снятие профиля памяти (из корня репозитория):
//
//	go test -run='^$' -bench=. -benchmem -memprofile=profiles/base.pprof ./internal/server/bench
//
// Сравнение профилей до и после оптимизаций:
//
//	go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof
//
// Разбор находок и результаты оптимизаций описаны в README, раздел
// «Профилирование памяти».
package bench
