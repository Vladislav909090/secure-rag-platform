# iam

Сервис идентификации и доступа: пользователи, роли, атрибуты subject context, JWT и пользовательские сессии в Redis.

## Порты и доступ

| Протокол | Порт |
|---|---:|
| HTTP | `8081` |
| gRPC | `9091` |

В глобальном compose IAM доступен другим сервисам по внутренней сети. Прямой HTTP через Traefik включается в dev-режиме из корня репозитория:

```bash
# из корня репозитория
make compose:up:dev
```

После этого:

- `http://localhost/iam/docs`
- `http://localhost/iam/health`

Для локального запуска отдельного IAM со своей PostgreSQL, Redis и миграциями:

```bash
cd services/iam
make compose:up
```

После этого доступны прямые порты `8081`, `9091`, `5433` и `6380`.

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

Миграции находятся в `services/iam/migrations`. Локальный compose сам применяет их при старте. Если нужно выполнить миграции вручную, локальный Makefile по умолчанию ждет БД на `localhost:5433`:

```bash
cd services/iam
make compose:up
make migrate:up
make migrate:status
```

## Разработка

```bash
cd services/iam
make api:gen
make grpc:stubs
make lint
make test
make build
make compose:config
```

Локальный запуск:

```bash
go run ./cmd/iam
```
