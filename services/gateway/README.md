# gateway

Публичная точка входа в платформу. Gateway принимает HTTP-запросы, проверяет доступ через IAM/OPA, вызывает Knowledge для документов и RAG для индексации и ответов.

## Порты и доступ

| Протокол | Порт |
|---|---:|
| HTTP | `8080` |
| gRPC | `9090` |

В compose сервис не публикует `8080` напрямую. С хоста используется Traefik:

- `http://localhost/gateway/docs`
- `http://localhost/gateway/health`

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

## Разработка

```bash
make proto:gen:gateway
make grpc:stubs:gateway
make lint:gateway
make test:gateway
make build:gateway
```

Локальный запуск из каталога сервиса:

```bash
cd services/gateway
go run ./cmd/gateway
```
