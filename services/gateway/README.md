# gateway

Публичная точка входа в платформу. Gateway принимает HTTP-запросы, проверяет доступ через IAM/OPA, вызывает Knowledge для документов и RAG для индексации и ответов.

## Порты и доступ

| Протокол | Порт |
|---|---:|
| HTTP | `8080` |
| gRPC | `9090` |

В глобальном dev/prod compose с хоста используется Traefik:

- `http://localhost/gateway/docs`
- `http://localhost/gateway/health`

Для локального запуска отдельного gateway из каталога сервиса:

```bash
cd services/gateway
make compose:up
```

После этого доступны прямые порты `8080`, `9090` и локальный OPA на `8181`.
IAM, Knowledge и RAG в этом режиме ожидаются на `host.docker.internal`.

## Основные маршруты

- `POST /gateway/api/v1/auth/login`
- `POST /gateway/api/v1/auth/refresh`
- `POST /gateway/api/v1/auth/logout`
- `GET /gateway/api/v1/auth/me`
- `GET /gateway/api/v1/documents`
- `POST /gateway/api/v1/documents`
- `GET /gateway/api/v1/documents/{document_uuid}`
- `GET /gateway/api/v1/documents/{document_uuid}/file`
- `DELETE /gateway/api/v1/documents/{document_uuid}`
- `POST /gateway/api/v1/rag/query`
- `POST /gateway/api/v1/admin/rag/documents/{document_uuid}/reindex`
- `POST /gateway/api/v1/admin/rag/reindex`

Swagger UI живет на `/gateway/docs`.

## Конфигурация

Самые важные переменные:

- `PORT`, `GRPC_PORT`
- `IAM_GRPC_ADDR`, `KNOWLEDGE_GRPC_ADDR`, `RAG_GRPC_ADDR`
- `OPA_URL`
- `DISABLE_AUTH`, `DISABLE_DOC_FILTER`
- `GATEWAY_DEFAULT_TOP_K`
- `GATEWAY_DEFAULT_EMBEDDING_MODEL_ALIAS`, `GATEWAY_DEFAULT_GENERATION_MODEL_ALIAS`

Если фильтрация документов включена (`DISABLE_DOC_FILTER=false`), `OPA_URL` обязателен.
При недоступной или не настроенной OPA gateway возвращает ошибку вместо локального
fallback-решения.

## Разработка

```bash
cd services/gateway
make api:gen
make grpc:stubs
make lint
make test
make build
make compose:config
```

Локальный запуск из каталога сервиса:

```bash
go run ./cmd/gateway
```
