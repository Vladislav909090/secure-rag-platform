# rag

Retrieval-Augmented Generation сервис.

## Порты

| Протокол | Порт по умолчанию |
|----------|--------------------|
| HTTP     | 8083               |
| gRPC     | 9093               |

## Контракт API

- Proto-файл: `services/rag/api/v1/rag.proto`
- Генерация: `make proto:gen:rag`
- Transport-стабы: `make grpc:stubs:rag`

## Базовые маршруты

- `GET /health`
- `GET /docs` (Swagger UI, OpenAPI встроен в HTML)

Через Traefik с хоста:

- `GET http://localhost/rag/health`
- `GET http://localhost/rag/docs`

Примечание: в `docker-compose` сервис использует только `expose`, прямой URL `http://localhost:8083` не публикуется.

## Запуск локально

Из корня репозитория:

```bash
go run ./services/rag/cmd/rag
```

Или из каталога сервиса:

```bash
cd services/rag
go run ./cmd/rag
```

## Проверки

```bash
make test:rag
make lint:rag
make build:rag
```

## Миграции

- Каталог миграций: `services/rag/migrations`
- DSN по умолчанию для make: `postgres://rag:rag@localhost:5435/rag?sslmode=disable`

```bash
make migrate:status:rag
make migrate:up:rag
make migrate:down:rag
make migrate:create:rag MIGRATION_NAME=add_rag_table
```
