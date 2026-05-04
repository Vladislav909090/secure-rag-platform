# iam

Identity and Access Management сервис.

## Порты

| Протокол | Порт по умолчанию |
|----------|--------------------|
| HTTP     | 8081               |
| gRPC     | 9091               |

## Контракт API

- Proto-файл: `services/iam/api/v1/iam.proto`
- Генерация: `make proto:gen:iam`
- Transport-стабы: `make grpc:stubs:iam`

## Базовые маршруты

- `GET /health`
- `GET /docs` (Swagger UI, OpenAPI встроен в HTML)

Через Traefik с хоста:

- `GET http://localhost/iam/health`
- `GET http://localhost/iam/docs`

Примечание: в `docker-compose` сервис использует только `expose`, прямой URL `http://localhost:8081` не публикуется.

## Запуск локально

Из корня репозитория:

```bash
go run ./services/iam/cmd/iam
```

Или из каталога сервиса:

```bash
cd services/iam
go run ./cmd/iam
```

## Проверки

```bash
make test:iam
make lint:iam
make build:iam
```

## Миграции

- Каталог миграций: `services/iam/migrations`
- DSN по умолчанию для make: `postgres://iam:iam@localhost:5433/iam?sslmode=disable`

```bash
make migrate:status:iam
make migrate:up:iam
make migrate:down:iam
```
