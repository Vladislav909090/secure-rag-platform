# secure-rag-platform

Минималистичное монорепо микросервисов на Go с gRPC + grpc-gateway.

## Архитектура

```
services/       — автономные микросервисы (gateway, iam, knowledge, rag)
  <svc>/api/    — .proto-контракты сервиса (gRPC + HTTP аннотации)
  <svc>/gen/    — сгенерированный Go-код (не коммитится)
third_party/    — google/api proto-зависимости (скачиваются автоматически)
deploy/         — Docker Compose + Traefik
```

## Правила изоляции

1. **Сервис не импортирует другой сервис.** Каждый сервис полностью автономен.
2. **Общий код запрещён.** Proto-контракты живут внутри каждого сервиса.
3. **API определяется в protobuf + `google.api.http` аннотации.** Один контракт покрывает и gRPC, и HTTP.
4. **gRPC для inter-service, HTTP для клиентов** (через grpc-gateway).

## Структура сервиса

```
services/<svc>/
├── api/v1/<svc>.proto        — исходный .proto файл
├── gen/v1/                   — сгенерированный Go-код
│   ├── <svc>.pb.go           — Go messages
│   ├── <svc>_grpc.pb.go      — gRPC server/client stubs
│   └── <svc>.pb.gw.go        — HTTP reverse-proxy (grpc-gateway)
├── gen/openapiv2/            — OpenAPI v2 (Swagger) спецификация
├── cmd/<svc>/main.go         — точка входа
├── internal/                 — внутренняя логика
├── Dockerfile
├── go.mod
└── README.md
```

## Сервисы и порты

| Сервис    | HTTP порт | Описание                     |
|-----------|-----------|------------------------------|
| gateway   | 8080      | API Gateway (единая точка входа) |
| iam       | 8081      | Identity & Access Management |
| knowledge | 8082      | Knowledge Base               |
| rag       | 8083      | Retrieval-Augmented Generation |

## Быстрый старт

```bash
# Установить protoc-плагины (один раз)
make proto:tools

# Сгенерировать код из proto
make proto:gen

# Запустить все сервисы
make compose:up

# Остановить
make compose:down
```

## Makefile targets

```bash
# ── Lint ──────────────────────
make lint                # Все сервисы
make lint:gateway        # Только gateway
make lint:iam
make lint:knowledge
make lint:rag

# ── Test ──────────────────────
make test                # Все сервисы
make test:gateway
make test:iam
make test:knowledge
make test:rag

# ── Build ─────────────────────
make build               # Все сервисы
make build:gateway
make build:iam
make build:knowledge
make build:rag

# ── Proto ─────────────────────
make proto:tools         # Установить protoc-gen-go и др. плагины
make proto:deps          # Скачать google/api protos в third_party/
make proto:gen           # Генерация всех сервисов
make proto:gen:gateway   # Генерация только gateway
make proto:gen:iam
make proto:gen:knowledge
make proto:gen:rag

# ── Docker ────────────────────
make compose:up          # Запуск Docker Compose
make compose:down        # Остановка Docker Compose
```

## Генерация proto

### Требования

- **protoc** — компилятор Protocol Buffers ([установка](https://grpc.io/docs/protoc-installation/))
- **Go-плагины** — устанавливаются через `make proto:tools`:
  - `protoc-gen-go` — Go messages
  - `protoc-gen-go-grpc` — gRPC stubs
  - `protoc-gen-grpc-gateway` — HTTP reverse-proxy
  - `protoc-gen-openapiv2` — OpenAPI спецификация

### Как это работает

1. `make proto:deps` — скачивает `google/api/annotations.proto` и `http.proto` в `third_party/`
2. `make proto:gen:<svc>` — вызывает `protoc` для `services/<svc>/api/v1/<svc>.proto`
3. Сгенерированный код кладётся в `services/<svc>/gen/v1/`

### Импорт в Go-коде

```go
import (
    gatewayv1 "example.com/project/gateway/gen/v1"
)
```

Модуль сервиса (`go.mod`) уже содержит правильный путь, `gen/v1/` находится внутри сервиса — никаких `replace` не нужно.

### Клиент vs Сервер

Генерация создаёт **и серверные, и клиентские** stubs в одном пакете:

| Файл                      | Что содержит                                    | Когда нужен |
|---------------------------|-------------------------------------------------|-------------|
| `<svc>.pb.go`             | Go-структуры (messages)                          | Всегда      |
| `<svc>_grpc.pb.go`        | `Register<Svc>ServiceServer()` + клиент-интерфейс | **Сервер**: реализуйте интерфейс и регистрируйте. **Клиент**: создайте `New<Svc>ServiceClient(conn)` |
| `<svc>.pb.gw.go`          | HTTP reverse-proxy handler                       | gateway (RegisterXxxHandlerFromEndpoint) |

**Серверная сторона** (например, `services/iam`):
```go
import iamv1 "example.com/project/iam/gen/v1"

type server struct {
    iamv1.UnimplementedIAMServiceServer
}

func (s *server) Authenticate(ctx context.Context, req *iamv1.AuthenticateRequest) (*iamv1.AuthenticateResponse, error) {
    // ...
}

// Регистрация:
grpcServer := grpc.NewServer()
iamv1.RegisterIAMServiceServer(grpcServer, &server{})
```

**Клиентская сторона** (например, `services/gateway` вызывает iam по gRPC):
```go
// gateway НЕ импортирует iam напрямую — используйте gRPC-клиент.
conn, _ := grpc.Dial("iam:50051", grpc.WithInsecure())
client := iamv1.NewIAMServiceClient(conn)
resp, _ := client.Authenticate(ctx, &iamv1.AuthenticateRequest{...})
```

> **Важно**: gateway импортирует proto-файлы iam только если ему нужен клиент. В этом случае нужно добавить зависимость через `go.mod` / `replace`. Но для изоляции рекомендуется общаться через HTTP/gRPC без прямого импорта чужих proto.

## Требования

- Go 1.22+
- protoc (Protocol Buffers compiler)
- Docker & Docker Compose