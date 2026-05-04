# secure-rag-platform

Минималистичное монорепо из 5 Go-сервисов на `gRPC + grpc-gateway`.

## Архитектура

```
services/
  gateway/
  iam/
  knowledge/
  rag/
  ai-inference/
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
| gateway   | 8080 | 9090 |      -     |
| iam       | 8081 | 9091 |    5433    |
| knowledge | 8082 | 9092 |    5434    |
| rag       | 8083 | 9093 |    5435    |
| ai-inference |  -  | 9094 |      -     |


- Единственный публичный вход: Traefik (`http://localhost`, dashboard: `http://localhost:8090`).
- HTTP-порты сервисов и порты PostgreSQL не публикуются на хост, используются только через `expose` внутри сети `traefik-net`.
- Прямые адреса вида `http://localhost:8081` и `localhost:5433` недоступны с хоста.
- `ai-inference` является внутренним gRPC-сервисом и не публикуется через Traefik.

## Инфраструктурные порты

В обычном режиме (`make compose:up`) Redis/MinIO/PostgreSQL доступны только внутри docker-сетей.

В dev-режиме (`make compose:up DEV=1`) публикуются следующие порты на хост:

| Компонент | Host порт | Container порт | Назначение |
|-----------|-----------|----------------|------------|
| iam-db | 5433 | 5432 | PostgreSQL iam |
| knowledge-db | 5434 | 5432 | PostgreSQL knowledge |
| rag-db | 5435 | 5432 | PostgreSQL rag |
| iam-redis | 6380 | 6379 | Redis iam |
| knowledge-minio | 9001 | 9001 | MinIO Console |
| ai-inference | 9094 | 9094 | gRPC ai-inference |

Примечание: S3 API MinIO работает на порту `9000` внутри контейнера (`knowledge-minio:9000`) и используется сервисом `knowledge` по внутренней сети; на хост по умолчанию не публикуется.

Внешние маршруты через Traefik:

- Gateway: `GET /gateway/docs`, `GET /gateway/v1/gateway/health`
- IAM: `GET /iam/docs`, `GET /iam/v1/iam/health`
- Knowledge: `GET /knowledge/docs`, `GET /knowledge/api/v1/knowledge/health`, `GET /knowledge/api/v1/documents`
- RAG: `GET /rag/docs`, `GET /rag/v1/rag/health`

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

# Запустить в dev-режиме (с публикацией только инфраструктурных портов: БД/MinIO)
make compose:up DEV=1

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
```

## Примечания

- `gen/` и `third_party/google/` заполняются генерацией (`make proto:*`).
- `grpcstubgen` не удаляет устаревшие файлы автоматически: лишние стабы удаляются вручную.
- Папки-каркасы, где пока нет кода (`repository`, `usecase`), удерживаются в git через `doc.go`.
- Для всех сервисов используется сервисный namespace в Traefik: `/gateway/*`, `/iam/*`, `/knowledge/*`, `/rag/*`.
- `ai-inference` не имеет HTTP API и предназначен для межсервисного gRPC-вызова.
- HTTP-доступ к сервисам оставлен только через Traefik; даже в `DEV=1` порты `8080-8083` на хост не публикуются.
