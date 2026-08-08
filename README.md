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