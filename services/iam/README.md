# iam

Identity & Access Management сервис.

## Порт

| Протокол | Порт |
|----------|------|
| HTTP     | 8081 |

## API

API определяется через protobuf-контракты в `libs/proto/proto/iam/v1/iam.proto` с аннотациями `google.api.http`.

После генерации кода (см. `libs/proto/README.md`), gRPC сервер будет обслуживать межсервисные вызовы, а grpc-gateway — HTTP.

## Сборка

```bash
cd services/iam
go build ./cmd/iam
```

## Тесты

```bash
go test ./...
```

## Линтинг

```bash
golangci-lint run ./...
```
