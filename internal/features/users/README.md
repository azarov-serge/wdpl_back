# Фича Users (профили пользователей)

## Статус

**Готово.** Реализованы домен, репозиторий, сервис, HTTP-хендлеры и роуты.

- ✅ `domain.go` — модель `UserProfile`
- ✅ `repository.go` — интерфейс `ProfileRepository`
- ✅ `repository_postgres.go` — реализация для PostgreSQL
- ✅ `service.go` — бизнес-логика (GetOrCreate, Update)
- ✅ `service_test.go` — unit-тесты сервиса (моки)
- ✅ `dto.go` — DTO запросов/ответов
- ✅ `handler.go` — GET/PUT /api/users/me
- ✅ `router.go` — регистрация роутов
- ✅ Роуты зарегистрированы в приложении (`internal/shared/http/router`)

Полное описание шага и API-контрактов: **`docs/steps/STEP2.md`**.

## Эндпоинты

- **GET /api/users/me** — профиль текущего пользователя (по JWT). При отсутствии профиля создаётся с дефолтами.
- **PUT /api/users/me** — обновление профиля (displayName, avatarURL, bio, locale, timezone).

Оба эндпоинта требуют авторизации (Bearer access-токен).
