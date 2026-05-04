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

- `GET /gateway/health`
- `GET /gateway/docs` (Swagger UI, OpenAPI встроен в HTML)
- `POST /gateway/api/v1/auth/login`
- `POST /gateway/api/v1/auth/refresh`
- `POST /gateway/api/v1/auth/logout`
- `GET /gateway/api/v1/auth/me`
- `GET /gateway/api/v1/documents`
- `GET /gateway/api/v1/documents/{document_uuid}`
- `GET /gateway/api/v1/documents/{document_uuid}/file`
- `DELETE /gateway/api/v1/documents/{document_uuid}`
- `POST /gateway/api/v1/rag/query`
- `POST /gateway/api/v1/admin/rag/documents/{document_uuid}/reindex`
- `POST /gateway/api/v1/admin/rag/reindex`

Через Traefik с хоста:

- `GET http://localhost/gateway/health`
- `GET http://localhost/gateway/docs`

Примечание: в `docker-compose` сервисы работают через `expose`, поэтому прямой URL `http://localhost:8080` не используется.

## Основные переменные окружения

- `IAM_GRPC_ADDR`, `KNOWLEDGE_GRPC_ADDR`, `RAG_GRPC_ADDR`
- `DISABLE_AUTH`, `DISABLE_DOC_FILTER`
- `GATEWAY_DEFAULT_TOP_K`, `GATEWAY_DEFAULT_EMBEDDING_MODEL_ALIAS`, `GATEWAY_DEFAULT_GENERATION_MODEL_ALIAS`

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

Gateway не использует собственную БД и не имеет миграций.
