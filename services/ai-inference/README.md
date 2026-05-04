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

Конфиг задаётся JSON-файлом (по умолчанию `services/ai-inference/config/models.json`).

В репозитории есть один шаблон и один локальный файл конфигурации:

- `services/ai-inference/config/models.example.json` — шаблон без токена
- `services/ai-inference/config/models.json` — локальный файл с токеном и адресами моделей, он не должен попадать в git

Формат файла:

```json
{
  "chat.default": {
    "task": "generation",
    "provider": "openai_compat",
    "model": "Qwen/Qwen2.5-3B-Instruct",
    "base_url": "https://glade-untrue-stooge.ngrok-free.dev/v1",
    "api_key": "<PUT_YOUR_BEARER_TOKEN_HERE>",
    "generation_defaults": {
      "temperature": 0.3,
      "top_p": 1,
      "max_tokens": 1024
    }
  },
  "embed.default": {
    "task": "embedding",
    "provider": "openai_compat",
    "model": "nomic-embed-text:latest",
    "base_url": "http://localhost:11434/v1"
  }
}
```

`provider` можно не указывать — по умолчанию используется `openai_compat`.
Требование: сейчас поддерживается только `openai_compat`.

## Запуск локально

Из корня репозитория:

```bash
cp ./services/ai-inference/config/models.example.json ./services/ai-inference/config/models.json
go run ./services/ai-inference/cmd/ai-inference --config ./services/ai-inference/config/models.json
```

Или из каталога сервиса:

```bash
cd services/ai-inference
cp ./config/models.example.json ./config/models.json
go run ./cmd/ai-inference --config ./config/models.json
```

Для Docker используется тот же `models.json`: compose монтирует его в контейнер как `/app/config/models.json`.

При старте сервис выполняет простой health-check по всем алиасам (короткий `chat/completions` для generation и `embeddings` для embedding).

## Проверки

```bash
make test:ai-inference
make lint:ai-inference
make build:ai-inference
```
