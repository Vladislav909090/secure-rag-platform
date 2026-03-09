# knowledge

Сервис базы знаний.

## Порты

| Протокол | Порт по умолчанию |
|----------|--------------------|
| HTTP     | 8082               |
| gRPC     | 9092               |

## Контракт API

- Proto-файл: `services/knowledge/api/v1/knowledge.proto`
- Генерация: `make proto:gen:knowledge`
- Transport-стабы: `make grpc:stubs:knowledge`

## Базовые маршруты

- `GET /v1/knowledge/health`
- `GET /docs` (Swagger UI, OpenAPI встроен в HTML)

Через Traefik с хоста:

- `GET http://localhost/v1/knowledge/health`
- `GET http://localhost/knowledge/docs`

Примечание: в `docker-compose` сервис использует только `expose`, прямой URL `http://localhost:8082` не публикуется.

## Запуск локально

Из корня репозитория:

```bash
go run ./services/knowledge/cmd/knowledge
```

Или из каталога сервиса:

```bash
cd services/knowledge
go run ./cmd/knowledge
```

## Проверки

```bash
make test:knowledge
make lint:knowledge
make build:knowledge
```

## Миграции

- Каталог миграций: `services/knowledge/migrations`
- DSN по умолчанию для make: `postgres://knowledge:knowledge@localhost:5434/knowledge?sslmode=disable`

```bash
make migrate:status:knowledge
make migrate:up:knowledge
make migrate:down:knowledge
make migrate:create:knowledge MIGRATION_NAME=add_knowledge_table
```
