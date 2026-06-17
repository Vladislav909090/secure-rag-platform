# knowledge

Сервис базы знаний. Хранит метаданные документов в PostgreSQL, файлы кладет в S3-совместимое хранилище MinIO/S3.

## Порты и доступ

| Протокол | Порт |
|---|---:|
| HTTP | `8082` |
| gRPC | `9092` |

В глобальном compose сервис используется gateway и RAG по внутренней сети. Прямой HTTP через Traefik включается в dev-режиме:

```bash
make compose:up:dev
```

После этого:

- `http://localhost/knowledge/docs`
- `http://localhost/knowledge/api/health`

Для isolated-запуска Knowledge со своей PostgreSQL, MinIO и миграциями:

```bash
cd services/knowledge
make compose:up
```

После этого доступны прямые порты `8082`, `9092`, `5434`, `9000` и `9001`.

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

Миграции находятся в `services/knowledge/migrations`. Локальный Makefile по умолчанию ждет БД на `localhost:5434`:

```bash
cd services/knowledge
make compose:up
make migrate:up
make migrate:status
```

## Разработка

```bash
cd services/knowledge
make api:gen
make grpc:stubs
make lint
make test
make build
make compose:config
```

Локальный запуск:

```bash
go run ./cmd/knowledge
```
