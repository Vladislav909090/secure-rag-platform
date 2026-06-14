# ai-inference

Внутренний gRPC-сервис для доступа к generation и embedding моделям через короткие алиасы вроде `chat.default` и `embed.default`.

Сейчас поддерживается OpenAI-compatible API: сервис вызывает `/chat/completions` для генерации и `/embeddings` для эмбеддингов.

## Порт и доступ

| Протокол | Порт |
|---|---:|
| HTTP (grpc-gateway) | `8084` |
| gRPC | `9094` |

В обычном compose сервис доступен только внутри сети приложения. Прямой HTTP через Traefik включается в dev-режиме:

```bash
make compose:up DEV=1
```

После этого:

- `http://localhost/ai-inference/health`
- `http://localhost/ai-inference/docs`

## HTTP API

- `GET /ai-inference/health`
- `POST /ai-inference/api/v1/generate`
- `POST /ai-inference/api/v1/models/list`
- `POST /ai-inference/api/v1/embed`
- `POST /ai-inference/api/v1/embed/batch`

## Конфиг моделей

По умолчанию сервис читает `config/models.json`. Для локальной разработки:

```bash
cd services/ai-inference
cp ./config/models.example.json ./config/models.json
```

Минимальный формат:

```json
{
  "chat.default": {
    "task": "generation",
    "provider": "openai_compat",
    "model": "model-name",
    "base_url": "https://example.com/v1",
    "api_key": "token-if-needed"
  },
  "embed.default": {
    "task": "embedding",
    "provider": "openai_compat",
    "model": "embedding-model-name",
    "base_url": "http://localhost:11434/v1"
  }
}
```

`provider` можно не указывать: будет использован `openai_compat`. Для generation можно добавить `generation_defaults`: `temperature`, `top_p`, `max_tokens`, `presence_penalty`, `frequency_penalty`.

При старте сервис валидирует алиасы и делает короткий health-check по каждому провайдеру.

## Конфигурация

- `HTTP_PORT`
- `GRPC_PORT`
- `AI_INFERENCE_PROVIDER_TIMEOUT`

## Разработка

```bash
make api:gen
make grpc:stubs:ai-inference
make lint:ai-inference
make test:ai-inference
make build:ai-inference
```

Локальный запуск:

```bash
cd services/ai-inference
go run ./cmd/ai-inference
```
