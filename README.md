# secure-rag-platform

Монорепозиторий RAG-платформы на Go. Внутри 5 сервисов: публичный `gateway`, IAM, база знаний, RAG-индексация/поиск и внутренний ai-inference для работы с LLM и embedding-моделями.

## Что внутри

```text
services/
  gateway/       # публичный API и оркестрация IAM/Knowledge/RAG
  iam/           # пользователи, роли, сессии, JWT
  knowledge/     # метаданные документов + файлы в S3/MinIO
  rag/           # чанкинг, embeddings, pgvector, генерация ответа
  ai-inference/  # gRPC + HTTP (grpc-gateway) для OpenAI-compatible моделей
deploy/
  compose/       # локальный docker compose
  traefik/       # маршрутизация HTTP API
  opa/           # политики доступа для gateway
tools/
  grpcstubgen/   # генератор transport/grpc стабов по proto
third_party/     # внешние proto-зависимости google/api
```

## Как это запускается

Обычный локальный запуск идет через Docker Compose:

```bash
make compose:up
```

После старта основной вход в систему:

- `http://localhost/gateway/docs`
- `http://localhost/gateway/health`
- Traefik dashboard: `http://localhost:8090`

По умолчанию с хоста опубликованы только Traefik `80` и dashboard `8090`. Базы, Redis, MinIO и внутренние HTTP-порты сервисов остаются внутри docker-сетей.

Для разработки с доступом к инфраструктуре с хоста:

```bash
make compose:up DEV=1
```

В этом режиме дополнительно публикуются:

| Компонент | Host | Для чего |
|---|---:|---|
| `iam-db` | `5433` | PostgreSQL IAM |
| `knowledge-db` | `5434` | PostgreSQL Knowledge |
| `rag-db` | `5435` | PostgreSQL + pgvector для RAG |
| `iam-redis` | `6380` | Redis для сессий IAM |
| `knowledge-minio` | `9001` | MinIO Console |

В `DEV=1` через Traefik также доступны прямые сервисные docs/health:

- `http://localhost/iam/docs`, `http://localhost/iam/health`
- `http://localhost/knowledge/docs`, `http://localhost/knowledge/api/health`
- `http://localhost/rag/docs`, `http://localhost/rag/health`
- `http://localhost/ai-inference/docs`, `http://localhost/ai-inference/health`

## Порты сервисов

| Сервис | HTTP внутри compose | gRPC внутри compose |
|---|---:|---:|
| `gateway` | `8080` | `9090` |
| `iam` | `8081` | `9091` |
| `knowledge` | `8082` | `9092` |
| `rag` | `8083` | `9093` |
| `ai-inference` | `8084` | `9094` |

Прямые адреса вроде `http://localhost:8081` в compose не используются: сервисы публикуются через Traefik или общаются друг с другом по внутренним сетям.

## Частые команды

```bash
# генерация proto-кода и grpc-gateway
make proto:gen

# генерация transport/grpc стабов
make grpc:stubs

# миграции всех сервисных БД
make migrate:up
make migrate:status

# проверки
make test
make lint
make build

# остановка compose
make compose:down

# пересборка и пересоздание compose-контейнеров с нуля
make compose:recreate
```

Перед первой генерацией может понадобиться установить `protoc` и Go-плагины:

```bash
make proto:tools
make proto:deps
```

## Модели для ai-inference

`ai-inference` читает конфиг из `services/ai-inference/config/models.json`. В репозитории должен лежать только шаблон:

```bash
cp services/ai-inference/config/models.example.json services/ai-inference/config/models.json
```

Файл `models.json` локальный: там могут быть URL моделей и токены. Compose монтирует его в контейнер как `/app/config/models.json`.

## Миграции

Миграции есть у `iam`, `knowledge` и `rag`. Команды `make migrate:*` рассчитаны на опубликованные dev-порты БД, поэтому перед ними обычно запускают:

```bash
make compose:up DEV=1
```

Затем:

```bash
make migrate:up
```

## Полезные замечания

- `gateway` является основным публичным API и проксирует бизнес-операции в IAM, Knowledge и RAG.
- `iam`, `knowledge` и `rag` можно открыть напрямую через Traefik только в `DEV=1`; обычный сценарий идет через `gateway`.
- `knowledge` хранит файлы в MinIO/S3, а метаданные в PostgreSQL.
- `rag` использует PostgreSQL с `pgvector`, читает файлы через Knowledge/MinIO и ходит в `ai-inference` за embeddings и генерацией.
- `third_party/google/api` восстанавливается командой `make proto:deps`, если нужных proto-файлов нет.
