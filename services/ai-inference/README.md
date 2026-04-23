# ai-inference

Внутренний gRPC-сервис для доступа к LLM и embedding-моделям через модельные алиасы.

## Порты

| Протокол | Порт по умолчанию |
|----------|--------------------|
| gRPC     | 9094               |

## Контракт API

- Proto-файл: `services/ai-inference/api/v1/ai_inference.proto`
- Генерация: `make proto:gen:ai-inference`
- Transport-стабы: `make grpc:stubs:ai-inference`

## Доступные gRPC сервисы

- `GenerationService`
  - `Generate`
  - `ListModels`
- `EmbeddingService`
  - `Embed`
  - `BatchEmbed`

## Конфиг моделей

Используется env `AI_INFERENCE_MODELS_JSON` с JSON-объектом вида:

```json
{
  "chat.default": {
    "task": "generation",
    "provider": "openai_compat",
    "model": "gpt-4o-mini",
    "base_url": "https://api.openai.com/v1",
    "api_key": "",
    "generation_defaults": {
      "temperature": 0.2,
      "top_p": 1,
      "max_tokens": 1024
    }
  },
  "embed.default": {
    "task": "embedding",
    "provider": "openai_compat",
    "model": "text-embedding-3-small",
    "base_url": "https://api.openai.com/v1",
    "api_key": ""
  }
}
```

Если env не задан, используются встроенные дефолтные алиасы.

## Запуск локально

Из корня репозитория:

```bash
go run ./services/ai-inference/cmd/ai-inference
```

Или из каталога сервиса:

```bash
cd services/ai-inference
go run ./cmd/ai-inference
```

## Проверки

```bash
make test:ai-inference
make lint:ai-inference
make build:ai-inference
```
