# WD Planning Backend

Go‑бэкенд для WD Planning: API авторизации, профилей пользователей и событий.

## Оглавление

- [Технологии](#технологии)
- [Структура проекта](#структура-проекта)
- [Требования](#требования)
- [Установка и запуск](#установка-и-запуск)
- [Переменные окружения](#переменные-окружения)
- [API для фронтенда](#api-для-фронтенда)
- [Сборка и тесты](#сборка-и-тесты)
- [Документация](#документация)

## Технологии

- **Go 1.21+**
- **Fiber v2** — HTTP‑фреймворк
- **PostgreSQL** — БД (миграции в `internal/shared/postgres/migrations/`)
- **JWT** (golang-jwt) — access‑токены
- **bcrypt** — хэширование паролей
- **go-playground/validator** — валидация DTO

## Структура проекта

```
cmd/server/           # Точка входа (main)
internal/
  features/           # Фичи (feature-first)
    auth/             # Регистрация, логин, refresh, logout
    events/           # События и черновики (публичные + защищённые)
    users/            # Профили (GET/PUT /api/users/me)
  shared/             # Общий код
    config/           # Конфигурация из env
    logger/           # Логгер
    postgres/         # Подключение к БД и миграции
    authutils/        # JWT, bcrypt, claims
    http/             # response, handler helpers, middleware, router
```

## Требования

- Go 1.21 или новее
- PostgreSQL (для локальной разработки можно использовать `docker-compose`)

## Установка и запуск

1. Клонируйте репозиторий и перейдите в каталог:

   ```bash
   cd wdpl_back
   ```

2. Скопируйте переменные окружения и при необходимости отредактируйте:

   ```bash
   cp .env.example .env
   ```

3. Запустите PostgreSQL (если ещё не запущен). Например, через Docker:

   ```bash
   docker-compose up -d
   ```

4. Запустите сервер:

   ```bash
   go run ./cmd/server
   ```

   По умолчанию API доступен на `http://0.0.0.0:3000`. Проверка: `curl http://localhost:3000/healthz`.

## Переменные окружения

Основные переменные (подробнее в `.env.example`):

| Переменная        | Описание                          |
|-------------------|-----------------------------------|
| `SERVER_PORT`     | Порт сервера (по умолчанию 3000)  |
| `DATABASE_URL`    | URL подключения к PostgreSQL     |
| `JWT_SECRET`      | Секрет для подписи access JWT     |
| `REFRESH_SECRET`  | Секрет для refresh‑токенов        |
| `CORS_ALLOWED_ORIGINS` | Разрешённые origins для CORS (или `*`) |

## API для фронтенда

- **Базовый URL:** `http://localhost:3000/api` (или тот хост/порт, на котором запущен сервер).
- **Проверка доступности:** `GET /healthz` → `{"status":"ok"}`.

Основные группы эндпоинтов:

| Группа   | Префикс        | Описание |
|----------|----------------|----------|
| Auth     | `/api/auth`    | sign-up, sign-in, sign-out, refresh |
| Events   | `/api/events`  | Публичные события; `/api/events/drafts` — черновики (требуют авторизации) |
| Users    | `/api/users`   | GET/PUT `/api/users/me` — профиль текущего пользователя (требуют JWT) |

Заголовок авторизации: `Authorization: Bearer <accessToken>`.

Подробные контракты и шаги разработки — в `docs/steps/` и `docs/SUPABASE.md`.

## Сборка и тесты

```bash
# Сборка
go build ./cmd/server

# Запуск тестов
go test ./...

# Покрытие
go test ./... -coverprofile=coverage.out -covermode=atomic
go tool cover -func=coverage.out
```

## Документация

- **docs/ARCHITECTURE.md** — обзор архитектуры (SOLID, DRY, KISS).
- **docs/SUPABASE.md** — схема БД и домены (Supabase/PostgreSQL).
- **docs/steps/** — пошаговое описание фич (STEP1–4, REMAINING_STEPS).
- **internal/features/users/README.md** — статус фичи users.

## Docker

```bash
docker-compose up -d   # PostgreSQL (и при необходимости другие сервисы)
docker build -t wdpl_back .   # Сборка образа приложения
```
