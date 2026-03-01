# rag

Retrieval-Augmented Generation — сервис генерации ответов с использованием базы знаний.

## Порт

| Протокол | Порт |
|----------|------|
| HTTP     | 8083 |

## API

API определяется через protobuf-контракты в `libs/proto/proto/rag/v1/rag.proto` с аннотациями `google.api.http`.

## Сборка

```bash
cd services/rag
go build ./cmd/rag
```

## Тесты

```bash
go test ./...
```

## Линтинг

```bash
golangci-lint run ./...
```
