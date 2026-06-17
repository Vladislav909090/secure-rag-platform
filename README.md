# secure-rag-platform

Монорепозиторий RAG-платформы на Go. Внутри 5 сервисов: публичный `gateway`, IAM, база знаний, RAG-индексация/поиск и внутренний `ai-inference` для работы с LLM и embedding-моделями.

## Что внутри

```text
api/            # зеркало proto-контрактов и сгенерированный Go/OpenAPI-код
services/
  gateway/       # публичный API и оркестрация IAM/Knowledge/RAG
  iam/           # пользователи, роли, сессии, JWT
  knowledge/     # метаданные документов + файлы в S3/MinIO
  rag/           # чанкинг, embeddings, pgvector, генерация ответа
  ai-inference/  # OpenAI-compatible generation/embedding adapter
deploy/
  compose/
    docker-compose.yml          # подключает базовые compose-файлы сервисов
    docker-compose.prod.yml     # prod-like сети, Traefik и связи сервисов
    docker-compose.dev.yml      # dev-only routes к внутренним сервисам
  traefik/       # маршрутизация HTTP API
  opa/           # политики доступа для gateway
```

Каждый сервис дополнительно содержит свой `Makefile`, `README.md`, `compose.yaml`, `compose.local.yaml`, `Dockerfile`, `third_party/` и `tools/`. Это позволяет запускать типовые команды из каталога сервиса, не держа всю механику только в корне.

## Глобальный запуск

По умолчанию корневые compose-команды используют dev-режим: `deploy/compose/docker-compose.yml`, `deploy/compose/docker-compose.prod.yml` и `deploy/compose/docker-compose.dev.yml`.

```bash
make compose:up
```

Для пересборки и пересоздания контейнеров:

```bash
make compose:recreate
```

Prod-like режим использует только `deploy/compose/docker-compose.yml` и `deploy/compose/docker-compose.prod.yml`:

```bash
make compose:up:prod
```

При старте compose одноразовые сервисы `iam-migrate`, `knowledge-migrate` и `rag-migrate` применяют SQL-миграции через `goose`, а основные сервисы ждут успешного завершения своей миграции. Это позволяет поднимать платформу на пустых Docker volumes без ручного предварительного `make migrate:up`.

Основной вход:

- `http://localhost/gateway/docs`
- `http://localhost/gateway/health`
- Traefik dashboard: `http://localhost:8090`

В dev-режиме через Traefik также доступны прямые сервисные docs/health:

- `http://localhost/iam/docs`, `http://localhost/iam/health`
- `http://localhost/knowledge/docs`, `http://localhost/knowledge/api/health`
- `http://localhost/rag/docs`, `http://localhost/rag/health`
- `http://localhost/ai-inference/docs`, `http://localhost/ai-inference/health`

## Порты сервисов

| Сервис | HTTP | gRPC |
|---|---:|---:|
| `gateway` | `8080` | `9090` |
| `iam` | `8081` | `9091` |
| `knowledge` | `8082` | `9092` |
| `rag` | `8083` | `9093` |
| `ai-inference` | `8084` | `9094` |

В глобальном compose сервисы публикуются через Traefik или общаются по внутренним docker-сетям. Прямые host-порты сервисов и инфраструктуры находятся только в service-local `compose.local.yaml`.

`deploy/compose/docker-compose.yml` подключает сервисные `services/<service>/compose.yaml` через Docker Compose `include`. В базовых compose-файлах лежит общая логика сервиса: build, environment, миграции, базы, MinIO/Redis/OPA и приватные сети. `deploy/compose/docker-compose.prod.yml` добавляет platform wiring: Traefik, общую `app-net`, межсервисные DNS-адреса и production-like настройки. `deploy/compose/docker-compose.dev.yml` добавляет dev-only Traefik routes к внутренним сервисам.

## Частые корневые команды

```bash
# генерация общих proto-контрактов, grpc-gateway и OpenAPI
make api:gen

# копирование service-local proto-контрактов в api/proto без генерации
make api:sync

# очистка сгенерированных API-файлов
make api:clean

# генерация transport/grpc стабов во всех сервисах
make grpc:stubs

# миграции всех сервисных БД
make migrate:validate
make migrate:up
make migrate:status

# проверки всех сервисов
make test
make lint
make build

# compose
make compose:config
make compose:config:prod
make compose:down
make compose:recreate
```

Перед первой генерацией может понадобиться установить `protoc` и Go-плагины:

```bash
make proto:tools
make proto:deps
```

## Команды из сервиса

В каждом `services/<service>` есть локальный `Makefile`:

```bash
cd services/rag

make test
make lint
make build
make proto:gen
make api:gen
make grpc:stubs
make compose:up
```

У `iam`, `knowledge` и `rag` там же есть локальные миграционные команды:

```bash
make migrate:validate
make migrate:up
make migrate:status
make migrate:create MIGRATION_NAME=add_new_table
```

`proto:gen` генерирует service-local файлы в `services/<service>/gen`. Рабочий код сервисов продолжает импортировать публичные Go-контракты из `secure-rag-platform/api/gen/go/...`, поэтому для обновления публичного контракта используется `api:gen`.

## Модели для ai-inference

`ai-inference` читает конфиг из `services/ai-inference/config/models.json`. В репозитории должен лежать только шаблон:

```bash
cp services/ai-inference/config/models.example.json services/ai-inference/config/models.json
```

Файл `models.json` локальный: там могут быть URL моделей и токены. Он игнорируется Git и не копируется в Docker image. Compose монтирует его в контейнер как `/app/config/models.json`, а внутри образа остается только безопасный `models.example.json`.

## Контракты

- Источник публичного контракта лежит рядом с сервисом: `services/<service>/api/v1/*.proto`.
- `api/proto` является зеркалом service-local контрактов для общей генерации.
- `api/gen/go` и `api/gen/openapiv2` хранят сгенерированные Go/OpenAPI файлы.
- Сервисы импортируют Go-контракты из `secure-rag-platform/api/gen/go/...`.
- Корневой `make api:gen` синхронизирует все service-local proto-файлы и генерирует весь общий API.
- Корневой `make api:gen` использует service-local include path из `PROTO_THIRD_PARTY` (по умолчанию `services/gateway/third_party`).
- Service-local `make api:gen` синхронизирует proto текущего сервиса и тоже генерирует общий API, используя локальную копию `third_party`.
- `services/<service>/third_party/google/api` восстанавливается командой `make proto:deps`.
- `proto:deps` качает googleapis с закрепленного `GOOGLEAPIS_REF`; при обновлении googleapis меняйте этот ref явно.

## Service-local compose

У каждого сервиса есть два compose-файла:

- `services/<service>/compose.yaml` — базовая конфигурация сервиса без host-портов;
- `services/<service>/compose.local.yaml` — standalone-режим с `ports`, `host.docker.internal` и локальными overrides.

Пример standalone-запуска:

```bash
cd services/gateway
make compose:up
```

Локальный набор зависимостей:

- `iam`: PostgreSQL, Redis, миграции;
- `knowledge`: PostgreSQL, MinIO, миграции;
- `rag`: pgvector, MinIO, миграции; Knowledge и ai-inference ожидаются по `host.docker.internal`;
- `gateway`: gateway и OPA; внешние сервисы ожидаются по `host.docker.internal`;
- `ai-inference`: только сервис и монтирование `config/models.json`.

Важно: Dockerfile сервисов пока используют общий `api` модуль, поэтому service-local compose собирает image с build context `../..`.
