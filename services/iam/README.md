# iam

Сервис идентификации и доступа: пользователи, роли, атрибуты subject context, JWT и пользовательские сессии в Redis.

## Порты и доступ

| Протокол | Порт |
|---|---:|
| HTTP | `8081` |
| gRPC | `9091` |

В обычном compose IAM доступен другим сервисам по внутренней сети. Прямой HTTP через Traefik включается в dev-режиме:

```bash
make compose:up DEV=1
```

После этого:

- `http://localhost/iam/docs`
- `http://localhost/iam/health`

## Что реализовано

- auth: login, refresh, logout, logout-all, me;
- пользователи: список, получение, создание, обновление;
- роли: список, чтение и изменение ролей пользователя;
- attributes: чтение, замена, patch, удаление ключа;
- sessions: список и отзыв сессий;
- internal API для gateway: subject context и проверка access token.

IAM хранит несколько активных refresh-сессий на пользователя. Новый login создает
новую строку `user_sessions`; logout/revoke отзывает конкретную сессию, а
logout-all/revoke-all отзывает все активные сессии пользователя.

## Конфигурация

Основные переменные:

- `PORT`, `GRPC_PORT`
- `DATABASE_DSN`
- `REDIS_ADDR`, `REDIS_PASSWORD`
- `JWT_SECRET`, `JWT_ISSUER`, `JWT_AUDIENCE`
- `BOOTSTRAP_ADMIN_LOGIN`, `BOOTSTRAP_ADMIN_PASSWORD`

В compose создается bootstrap-админ `superadmin / superadmin`.

## Миграции

Миграции находятся в `services/iam/migrations`. Makefile по умолчанию ждет БД на `localhost:5433`, поэтому для локального применения нужен dev-compose:

```bash
make compose:up DEV=1
make migrate:up:iam
make migrate:status:iam
```

## Разработка

```bash
make api:gen
make grpc:stubs:iam
make lint:iam
make test:iam
make build:iam
```

Локальный запуск:

```bash
cd services/iam
go run ./cmd/iam
```
