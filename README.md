# go-musthave-metrics-tpl

Шаблон репозитория для трека «Сервер сбора метрик и алертинга».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**

## Профилирование памяти

Бенчмарки компонентов лежат в `internal/server/bench`

Базовый профиль (`profiles/base.pprof`) показал три источника лишних аллокаций:

- `gzip.NewWriter` в compress-middleware создавался на каждый ответ — это ~88%
  всего alloc_space: writer держит ~800 КБ внутренних буферов flate, которые
  выделялись и выбрасывались на каждом запросе ради сжатия десятков байт JSON;
- hash-middleware на каждый запрос читал тело через `io.ReadAll` и заводил
  новый `bytes.Buffer` под буферизацию ответа;
- list-хендлер парсил HTML-шаблон страницы метрик на каждый запрос
  (~3000 allocs/op у `BenchmarkHTTP_List`).

Оптимизации: `sync.Pool` для `gzip.Writer`/`gzip.Reader` с `Reset()` вместо
создания новых, пул буферов в hash-middleware, разовый парсинг шаблона
при старте пакета.

До и после (на операцию):

| Бенчмарк | До | После |
|---|---|---|
| HTTP_UpdateJSON | 826016 B/op, 94 allocs | 13169 B/op, 72 allocs |
| HTTP_UpdatesBatch | 831849 B/op, 271 allocs | 29606 B/op, 252 allocs |
| HTTP_List | 912381 B/op, 3013 allocs | 83117 B/op, 2853 allocs |

По изменениям профилей аллокации сократились на ~75 ГБ за запуск (−87% alloc_space):
`flate.NewWriter` + `initDeflate` дали −74.7 ГБ, `io.ReadAll` в hash-middleware
−530 МБ, путь `template.Execute` в list-хендлере −3.5 ГБ.

Итог сравнения профилей (`pprof -top -diff_base=profiles/base.pprof profiles/result.pprof`):

```
File: bench.test
Type: alloc_space
Time: 2026-08-08 18:22:27 MSK
Showing nodes accounting for -75212.84MB, 87.09% of 86358.25MB total
Dropped 215 nodes (cum <= 431.79MB)
      flat  flat%   sum%        cum   cum%
-61708.15MB 71.46% 71.46% -75075.46MB 86.93%  compress/flate.NewWriter (inline)
-13006.53MB 15.06% 86.52% -13006.53MB 15.06%  compress/flate.(*compressor).initDeflate (inline)
 -530.30MB  0.61% 87.13%  -530.30MB  0.61%  io.ReadAll
    -457MB  0.53% 87.66%     -457MB  0.53%  compress/flate.(*huffmanEncoder).generate
  438.14MB  0.51% 87.15%   438.14MB  0.51%  bufio.NewReaderSize (inline)
   43.01MB  0.05% 87.10% -52689.04MB 61.01%  github.com/a-aleesshin/metrics/internal/server/bench.newBenchServer.RequestLogger.func2.1
       8MB 0.0093% 87.09% -24947.04MB 28.89%  github.com/a-aleesshin/metrics/internal/server/transport/http/handlers/metrics.(*UpdateJsonHandler).UpdateJSON
       4MB 0.0046% 87.09% -24492.22MB 28.36%  github.com/a-aleesshin/metrics/internal/server/transport/http/handlers/metrics.(*ValueJsonHandler).ValueJSON
   -3.50MB 0.0041% 87.09%   467.11MB  0.54%  net/http/httptest.NewRequestWithContext
      -1MB 0.0012% 87.09% -3516.91MB  4.07%  text/template.(*Template).execute
    0.50MB 0.00058% 87.09% -3577.95MB  4.14%  github.com/a-aleesshin/metrics/internal/server/transport/http/handlers/metrics.(*ListMetricsHandler).List
         0     0% 87.09% -22336.77MB 25.87%  compress/gzip.(*Writer).Close
         0     0% 87.09% -75080.96MB 86.94%  compress/gzip.(*Writer).Write
         0     0% 87.09% -49629.32MB 57.47%  encoding/json.(*Encoder).Encode
         0     0% 87.09% -75067.32MB 86.93%  github.com/a-aleesshin/metrics/internal/server/transport/http/middleware.CompressResponse.func1
         0     0% 87.09% -3561.44MB  4.12%  html/template.(*Template).Execute
```
### Агент

Бенчмарки агента лежат в `internal/agent/bench`:

По базовому профилю ~94% alloc_space давал `gzip.NewWriter`, создаваемый
на каждую отправку в `sendGzippedJSON`. Оптимизация `sync.Pool` для
`gzip.Writer` и буфера сжатого тела.

До и после (B/op на операцию):

| Бенчмарк | До | После |
|---|---|---|
| ReportMetrics | 823522 B/op, 113 allocs | 9973 B/op, 91 allocs |
| SendBatch | 821269 B/op, 84 allocs | 6839 B/op, 62 allocs |

Итог сравнения профилей:

```
File: bench.test
Type: alloc_space
Time: 2026-08-08 23:43:12 MSK
Showing nodes accounting for -49388.74MB, 93.62% of 52755.33MB total
Dropped 76 nodes (cum <= 263.78MB)
      flat  flat%   sum%        cum   cum%
-40487.69MB 76.75% 76.75% -49248.53MB 93.35%  compress/flate.NewWriter (inline)
-8527.17MB 16.16% 92.91% -8527.17MB 16.16%  compress/flate.(*compressor).initDeflate (inline)
    -455MB  0.86% 93.77%     -455MB  0.86%  compress/flate.(*huffmanEncoder).generate
   83.62MB  0.16% 93.61% -49485.33MB 93.80%  github.com/a-aleesshin/metrics/internal/agent/infra/http.(*MetricSender).SendBatch
   -2.50MB 0.0047% 93.62% -49625.02MB 94.07%  github.com/a-aleesshin/metrics/internal/agent/infra/http.(*MetricSender).sendGzippedJSON
         0     0% 93.62%  -496.02MB  0.94%  compress/gzip.(*Writer).Close
         0     0% 93.62% -49251.53MB 93.36%  compress/gzip.(*Writer).Write
         0     0% 93.62% -23183.15MB 43.94%  github.com/a-aleesshin/metrics/internal/agent/application/usecase.(*ReportMetricsUseCase).Execute
         0     0% 93.62% -23246.70MB 44.07%  github.com/a-aleesshin/metrics/internal/agent/application/usecase.(*ReportMetricsUseCase).SendMetrics
```
## Инструменты проекта

### Статический анализ (cmd/staticlint, cmd/multichecker)

`cmd/staticlint` — singlechecker с собственным анализатором `osexit`:
запрет прямого вызова `os.Exit` в `main` пакета `main`.

```
go run ./cmd/staticlint ./...
```

`cmd/multichecker` — расширенный набор: стандартные анализаторы
`x/tools/passes`, все `SA` и выборочные `S`/`ST`/`QF` из staticcheck,
публичные `bodyclose` и `nilerr`, плюс собственные `osexit` и `nopanic`
(сообщает об использовании встроенной функции `panic`). Подробное описание
каждого анализатора — в godoc пакета.

```
go run ./cmd/multichecker ./...
```

### Генератор Reset-методов (cmd/reset)

Утилита сканирует пакеты модуля и для структур, помеченных комментарием
`// generate:reset`, генерирует методы `Reset()` в файл `reset.gen.go` пакета
(примитивы — к нулю, слайсы — `s[:0]`, мапы — `clear`, вложенные структуры —
вызов их `Reset()`). Используется пулом объектов `internal/platform/pool`:
`Put` сбрасывает объект перед возвратом, поэтому из `Get` всегда приходит
чистый объект.

```
go run ./cmd/reset
```

### Сборка с информацией о версии

Сервер и агент при старте печатают версию, дату и коммит сборки:

```
Build version: <buildVersion>
Build date: <buildDate>
Build commit: <buildCommit>
```

Значения задаются на этапе сборки через `-ldflags -X` для глобальных
переменных `main.buildVersion`, `main.buildDate` и `main.buildCommit`
пакетов `cmd/server` и `cmd/agent`:

```
go build -ldflags "-X main.buildVersion=v1.0.0 -X main.buildDate=$(date +%Y-%m-%d) -X main.buildCommit=$(git rev-parse --short HEAD)" -o server ./cmd/server
```

```
go build -ldflags "-X main.buildVersion=v1.0.0 -X main.buildDate=$(date +%Y-%m-%d) -X main.buildCommit=$(git rev-parse --short HEAD)" -o agent ./cmd/agent
```

Если собрать без `-ldflags` (например, обычным `go build ./cmd/server`),
вместо незаданных значений будет напечатано `N/A`.

### gRPC-транспорт метрик

Помимо HTTP, агент может отправлять метрики батчами по gRPC (протокол —
`api/proto/metrics.proto`, сгенерированный код — `internal/proto`).
Сервис `Metrics.UpdateMetrics` на сервере использует тот же usecase
батчевого обновления, что и `POST /updates`.

Запуск сервера с gRPC и проверкой доверенной подсети (interceptor читает
IP агента из метаданных `x-real-ip`; запросы вне подсети отклоняются:

```
./server -grpc-address 127.0.0.1:3200 -t 192.168.0.0/24
```

Агент переключается на gRPC, когда задан адрес:

```
./agent -grpc-address 127.0.0.1:3200
```

Параметры настраиваются флагами (`-grpc-address`, `-t`), переменными
окружения (`GRPC_ADDRESS`, `TRUSTED_SUBNET`) и JSON-конфигом
(`grpc_address`, `trusted_subnet`) с обычным приоритетом источников.

Перегенерация кода из proto (нужны protoc, protoc-gen-go, protoc-gen-go-grpc):

```
protoc --proto_path=api/proto --go_out=internal/proto --go_opt=paths=source_relative --go-grpc_out=internal/proto --go-grpc_opt=paths=source_relative api/proto/metrics.proto
```
