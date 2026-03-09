# secure-rag-platform

Минималистичное монорепо из 4 Go-сервисов на `gRPC + grpc-gateway`.

## Архитектура

```
services/
  gateway/
  iam/
  knowledge/
  rag/
deploy/
  compose/              # docker-compose для локального запуска
  traefik/              # конфиг Traefik
third_party/google/api/ # внешние proto-зависимости
tools/grpcstubgen/      # генератор transport-стабов
```

Каждый сервис полностью автономен: свой `go.mod`, свой `api/v1/*.proto`, свой `migrations/`.

## Структура отдельного сервиса

```
services/<svc>/
  api/v1/<svc>.proto
  cmd/<svc>/main.go
  internal/
    app/
    closer/
    config/
    docs/
    transport/grpc/
    repository/
    usecase/
  gen/v1/
  gen/openapiv2/v1/
  migrations/
```

## Порты сервисов

| Сервис    | HTTP | gRPC | PostgreSQL |
|-----------|------|------|------------|
| gateway   | 8080 | 9090 |    5432    |
| iam       | 8081 | 9091 |    5433    |
| knowledge | 8082 | 9092 |    5434    |
| rag       | 8083 | 9093 |    5435    |


- Единственный публичный вход: Traefik (`http://localhost`, dashboard: `http://localhost:8090`).
- HTTP-порты сервисов и порты PostgreSQL не публикуются на хост, используются только через `expose` внутри сети `traefik-net`.
- Прямые адреса вида `http://localhost:8081` и `localhost:5433` недоступны с хоста.

Внешние маршруты через Traefik:

- Gateway: `GET /docs`, `GET /v1/gateway/health`
- IAM: `GET /iam/docs`, `GET /v1/iam/health`
- Knowledge: `GET /knowledge/docs`, `GET /v1/knowledge/health`
- RAG: `GET /rag/docs`, `GET /v1/rag/health`

## Быстрый старт

```bash
# Установить protoc-плагины (один раз)
make proto:tools

# Сгенерировать код из proto
make proto:gen

# Сгенерировать stub хендлеров gRPC ручек
make grpc:stubs

# Запустить все сервисы
make compose:up

# Применить все миграции
make migrate:up

# Остановить
make compose:down
```

## Основные команды

```bash
# Проверки
make lint
make test
make build

# Генерация
make proto:deps
make proto:gen
make grpc:stubs

# Миграции
make migrate:status
make migrate:up
make migrate:down
make migrate:create:gateway MIGRATION_NAME=add_users_table
make migrate:create:iam MIGRATION_NAME=add_tokens_table
make migrate:create:knowledge MIGRATION_NAME=add_documents_table
make migrate:create:rag MIGRATION_NAME=add_jobs_table
```

## Примечания

- `gen/` и `third_party/google/` заполняются генерацией (`make proto:*`).
- `grpcstubgen` не удаляет устаревшие файлы автоматически: лишние стабы удаляются вручную.
- Папки-каркасы, где пока нет кода (`repository`, `usecase`), удерживаются в git через `doc.go`.
- Для `iam`, `knowledge`, `rag` префиксы `/iam`, `/knowledge`, `/rag` и `/v1/iam`, `/v1/knowledge`, `/v1/rag` обрабатываются Traefik middleware `StripPrefix`.