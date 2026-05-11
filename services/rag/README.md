# rag

Сервис RAG: индексирует документы, хранит chunks и embeddings в PostgreSQL с `pgvector`, ищет релевантный контекст и вызывает `ai-inference` для генерации ответа.

## Порты и доступ

| Протокол | Порт |
|---|---:|
| HTTP | `8083` |
| gRPC | `9093` |

В обычном compose RAG вызывается gateway по внутренней сети. Прямой HTTP через Traefik включается в dev-режиме:

```bash
make compose:up DEV=1
```

После этого:

- `http://localhost/rag/docs`
- `http://localhost/rag/health`

## Основные маршруты

- `POST /rag/api/v1/documents/{document_uuid}/index`
- `POST /rag/api/v1/query`

Удаление индекса документа есть в gRPC-контракте, но HTTP-аннотация для него сейчас не задана.

## Конфигурация

Основные переменные:

- `PORT`, `GRPC_PORT`
- `DATABASE_DSN`
- `S3_ENDPOINT`, `S3_BUCKET`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `S3_USE_SSL`
- `KNOWLEDGE_GRPC_ADDR`, `AI_INFERENCE_GRPC_ADDR`
- `RAG_CHUNK_SIZE`, `RAG_CHUNK_OVERLAP`, `RAG_DEFAULT_TOP_K`
- `RAG_DEFAULT_EMBEDDING_MODEL_ALIAS`, `RAG_DEFAULT_GENERATION_MODEL_ALIAS`

По умолчанию используются `embed.default`, `chat.default`, chunk size `800`, overlap `100`, top-k `3`.

## Миграции

Миграции находятся в `services/rag/migrations`. Makefile по умолчанию ждет pgvector-БД на `localhost:5435`:

```bash
make compose:up DEV=1
make migrate:up:rag
make migrate:status:rag
```

## Разработка

```bash
make proto:gen:rag
make grpc:stubs:rag
make lint:rag
make test:rag
make build:rag
```

Локальный запуск:

```bash
cd services/rag
go run ./cmd/rag
```
