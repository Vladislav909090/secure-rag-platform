# api

Общий модуль публичных контрактов платформы.

```text
proto/          # зеркало service-local protobuf-контрактов
gen/go/         # сгенерированный Go-код и grpc-gateway handlers
gen/openapiv2/  # сгенерированные Swagger/OpenAPI JSON
```

Исходные `.proto` редактируются в `services/<service>/api/v1`. Каталог
`api/proto` обновляется командой `make api:sync` или автоматически перед
`make api:gen`.

Сервисы импортируют Go-контракты из `secure-rag-platform/api/gen/go/...`.
Исходники сервисов не должны импортировать generated-код из соседних `services/*`.

Основные команды из корня репозитория:

```bash
make api:sync
make api:gen
make api:clean
```

`make api:gen` сначала синхронизирует `api/proto`, затем вызывает `protoc`
напрямую из корневого `Makefile`.

Те же публичные контракты можно обновлять из каталога конкретного сервиса:

```bash
cd services/rag
make api:gen
```

Service-local команда синхронизирует `.proto` текущего сервиса в `api/proto`
и генерирует общий `api/gen`, используя локальную копию `third_party`.
