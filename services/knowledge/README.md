# knowledge

Сервис базы знаний. Хранит метаданные документов в PostgreSQL, файлы кладет в S3-совместимое хранилище MinIO/S3.

## Порты и доступ

| Протокол | Порт |
|---|---:|
| HTTP | `8082` |
| gRPC | `9092` |

В обычном compose сервис используется gateway и RAG по внутренней сети. Прямой HTTP через Traefik включается в dev-режиме:

```bash
make compose:up DEV=1
```

После этого:

- `http://localhost/knowledge/docs`
- `http://localhost/knowledge/api/health`

## Основные маршруты

- `GET /knowledge/api/v1/documents`
- `POST /knowledge/api/v1/documents`
- `GET /knowledge/api/v1/documents/{document_uuid}`
- `GET /knowledge/api/v1/documents/{document_uuid}/file`
- `PATCH /knowledge/api/v1/documents/{document_uuid}`
- `PATCH /knowledge/api/v1/documents/{document_uuid}/attributes`
- `DELETE /knowledge/api/v1/documents/{document_uuid}`
- `POST /knowledge/api/v1/documents/{document_uuid}/restore`
- `POST /knowledge/api/v1/documents/{document_uuid}/reindex`

Загрузка файла по HTTP реализована отдельным handler-ом, потому что gRPC-контракт использует streaming.

## Конфигурация

Основные переменные:

- `PORT`, `GRPC_PORT`
- `DATABASE_DSN`
- `S3_ENDPOINT`, `S3_BUCKET`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `S3_USE_SSL`
- `MAX_FILE_SIZE` для ограничения размера загружаемого файла

## Миграции

Миграции находятся в `services/knowledge/migrations`. Makefile по умолчанию ждет БД на `localhost:5434`:

```bash
make compose:up DEV=1
make migrate:up:knowledge
make migrate:status:knowledge
```

## Разработка

```bash
make api:gen
make grpc:stubs:knowledge
make lint:knowledge
make test:knowledge
make build:knowledge
```

Локальный запуск:

```bash
cd services/knowledge
go run ./cmd/knowledge
```
