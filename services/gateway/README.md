# gateway

API Gateway — единая точка входа для внешних клиентов.

## Порт

| Протокол | Порт |
|----------|------|
| HTTP     | 8080 |

## API

API определяется через protobuf-контракты в `libs/proto/proto/gateway/v1/gateway.proto` с аннотациями `google.api.http`.

После генерации кода (см. `libs/proto/README.md`), grpc-gateway runtime будет обслуживать HTTP API на базе сгенерированных stubs.

## Сборка

```bash
cd services/gateway
go build ./cmd/gateway
```

## Тесты

```bash
go test ./...
```

## Линтинг

```bash
golangci-lint run ./...
```
