# rag

Сервис RAG: индексирует документы, хранит chunks и embeddings в PostgreSQL с `pgvector`, ищет релевантный контекст и вызывает `ai-inference` для генерации ответа.

## Порты и доступ

| Протокол | Порт |
|---|---:|
| HTTP | `8083` |
| gRPC | `9093` |

В глобальном compose RAG вызывается gateway по внутренней сети. Прямой HTTP через Traefik включается в dev-режиме из корня репозитория:

```bash
# из корня репозитория
make compose:up:dev
```

После этого:

- `http://localhost/rag/docs`
- `http://localhost/rag/health`

Для локального запуска отдельного RAG со своей pgvector-БД, MinIO и миграциями:

```bash
cd services/rag
make compose:up
```

После этого доступны прямые порты `8083`, `9093`, `5435`, `9010` и `9011`.
Knowledge и ai-inference в этом compose ожидаются на `host.docker.internal:9092`
и `host.docker.internal:9094`.

## Основные маршруты

- `POST /rag/api/v1/documents/{document_uuid}/index`
- `DELETE /rag/api/v1/documents/{document_uuid}/index`
- `POST /rag/api/v1/query`

## Конфигурация

Основные переменные:

- `PORT`, `GRPC_PORT`
- `DATABASE_DSN`
- `S3_ENDPOINT`, `S3_BUCKET`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `S3_USE_SSL`
- `KNOWLEDGE_GRPC_ADDR`, `AI_INFERENCE_GRPC_ADDR`
- `RAG_CHUNK_SIZE`, `RAG_CHUNK_OVERLAP`, `RAG_DEFAULT_TOP_K`
- `RAG_DEFAULT_EMBEDDING_MODEL_ALIAS`, `RAG_DEFAULT_GENERATION_MODEL_ALIAS`
- `RAG_INDEXED_EMBEDDING_DIMENSION`

По умолчанию используются `embed.default`, `chat.default`, chunk size `800`, overlap `100`, top-k `3`.

Embeddings хранятся в `pgvector` без фиксированной размерности колонки. Поиск
фильтруется по `embedding_model` и `embedding_dimension`, поэтому в одной таблице
можно держать разные embedding-модели. При старте RAG создает partial HNSW index
для размерности `RAG_INDEXED_EMBEDDING_DIMENSION` (по умолчанию `768`) и использует
его для запросов с такой же размерностью. Для других размерностей поиск остается
корректным, но без этого HNSW index.

## Миграции

Миграции находятся в `services/rag/migrations`. Локальный compose сам применяет их при старте. Если нужно выполнить миграции вручную, локальный Makefile по умолчанию ждет pgvector-БД на `localhost:5435`:

```bash
cd services/rag
make compose:up
make migrate:up
make migrate:status
```

## Разработка

```bash
cd services/rag
make api:gen
make grpc:stubs
make lint
make test
make build
make compose:config
```

Локальный запуск:

```bash
go run ./cmd/rag
```
