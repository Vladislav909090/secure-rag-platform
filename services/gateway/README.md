# gateway

Gateway-сервис, единая точка входа для API.

## Порты

| Протокол | Порт по умолчанию |
|----------|--------------------|
| HTTP     | 8080               |
| gRPC     | 9090               |

## Контракт API

- Proto-файл: `services/gateway/api/v1/gateway.proto`
- Генерация: `make proto:gen:gateway`
- Transport-стабы: `make grpc:stubs:gateway`

## Базовые маршруты

- `GET /v1/gateway/health`
- `GET /docs` (Swagger UI, OpenAPI встроен в HTML)

Через Traefik с хоста:

- `GET http://localhost/v1/gateway/health`
- `GET http://localhost/docs`

Примечание: в `docker-compose` сервисы работают через `expose`, поэтому прямой URL `http://localhost:8080` не используется.

## Запуск локально

Из корня репозитория:

```bash
go run ./services/gateway/cmd/gateway
```

Или из каталога сервиса:

```bash
cd services/gateway
go run ./cmd/gateway
```

## Проверки

```bash
make test:gateway
make lint:gateway
make build:gateway
```

## Миграции

- Каталог миграций: `services/gateway/migrations`
- DSN по умолчанию для make: `postgres://gateway:gateway@localhost:5432/gateway?sslmode=disable`

```bash
make migrate:status:gateway
make migrate:up:gateway
make migrate:down:gateway
make migrate:create:gateway MIGRATION_NAME=add_gateway_table
```
