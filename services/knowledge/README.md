# knowledge

Knowledge Base сервис — управление базой знаний.

## Порт

| Протокол | Порт |
|----------|------|
| HTTP     | 8082 |

## API

API определяется через protobuf-контракты в `libs/proto/proto/knowledge/v1/knowledge.proto` с аннотациями `google.api.http`.

## Сборка

```bash
cd services/knowledge
go build ./cmd/knowledge
```

## Тесты

```bash
go test ./...
```

## Линтинг

```bash
golangci-lint run ./...
```
