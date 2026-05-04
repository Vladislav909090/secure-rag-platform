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
- `POST /rag/api/v1/documents/{document_uuid}/index`
- `DELETE /rag/api/v1/documents/{document_uuid}/index`
- `POST /rag/api/v1/query`

Через Traefik с хоста:

- `GET http://localhost/rag/health`
- `GET http://localhost/rag/docs`

Примечание: в `docker-compose` сервис использует только `expose`, прямой URL `http://localhost:8083` не публикуется.

## Основные переменные окружения

- `DATABASE_DSN` (или `DB_DSN`)
- `S3_ENDPOINT`, `S3_BUCKET`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `S3_USE_SSL`
- `KNOWLEDGE_GRPC_ADDR`, `AI_INFERENCE_GRPC_ADDR`
- `RAG_CHUNK_SIZE`, `RAG_CHUNK_OVERLAP`, `RAG_DEFAULT_TOP_K`
- `RAG_DEFAULT_EMBEDDING_MODEL_ALIAS`, `RAG_DEFAULT_GENERATION_MODEL_ALIAS`

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
```
