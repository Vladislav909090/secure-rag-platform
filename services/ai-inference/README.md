# ai-inference

Внутренний gRPC-сервис для доступа к generation и embedding моделям через короткие алиасы вроде `chat.default` и `embed.default`.

Сейчас поддерживается OpenAI-compatible API: сервис вызывает `/chat/completions` для генерации и `/embeddings` для эмбеддингов.

## Порт и доступ

| Протокол | Порт |
|---|---:|
| HTTP (grpc-gateway) | `8084` |
| gRPC | `9094` |

В глобальном compose сервис доступен только внутри сети приложения. Прямой HTTP через Traefik включается в dev-режиме из корня репозитория:

```bash
# из корня репозитория
make compose:up:dev
```

После этого:

- `http://localhost/ai-inference/health`
- `http://localhost/ai-inference/docs`

Для локального запуска отдельного ai-inference из каталога сервиса:

```bash
cd services/ai-inference
make compose:up
```

После этого доступны прямые порты `8084` и `9094`.

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
Для локального режима без проверки живости внешних моделей можно выставить:

```bash
AI_INFERENCE_SKIP_PROVIDER_HEALTHCHECK=true
```

В этом режиме startup и `/health` не вызывают embedding/LLM провайдеры, но обычные `generate`/`embed` запросы остаются реальными. Для полностью моковых ответов:

```bash
AI_INFERENCE_MOCK_RESPONSES=true
```

Mock-режим также отключает live health-check и явно пишет warning в логи при старте.

## Конфигурация

- `HTTP_PORT`
- `GRPC_PORT`
- `AI_INFERENCE_PROVIDER_TIMEOUT`
- `AI_INFERENCE_SKIP_PROVIDER_HEALTHCHECK`
- `AI_INFERENCE_MOCK_RESPONSES`

## Разработка

```bash
cd services/ai-inference
make api:gen
make grpc:stubs
make lint
make test
make build
make compose:config
```

Локальный запуск:

```bash
go run ./cmd/ai-inference
```
