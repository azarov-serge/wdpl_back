## Шаг 2: Фича Users (управление пользователями)

### Цель шага

- **Отделить учётные записи (auth)** от **профиля пользователя (users)**.
- Дать фронтенду удобный способ:
  - получать данные о текущем пользователе (профиль, имя, аватар и т.п.),
  - редактировать профиль,
  - в будущем — видеть чужие профили (для команд, организаций, событий).
- Сохранить принцип Feature-First: `internal/features/users` отвечает только за пользовательские данные и не знает о паролях/JWT.

### Схема БД и размещение таблиц

- **Схема**: для профилей пользователей можно использовать `public` (по умолчанию) или отдельную `core`. Для простоты шага 2 используем **`public`**.
- Связь с auth:
  - Таблица `auth.users` уже хранит идентификатор и e‑mail (логин).
  - Фича `users` хранит дополнительные данные профиля и **ссылается на `auth.users.id`**.

#### Таблица `public.user_profiles`

Назначение: расширяем технический юзер из `auth.users` бизнес‑полями.

- **`user_id UUID PRIMARY KEY`**
  - Также `FOREIGN KEY` на `auth.users(id) ON DELETE CASCADE`.
  - PK = FK подчёркивает, что у каждого пользователя максимум один профиль.
- **`display_name TEXT NOT NULL`**
  - Отображаемое имя (можно инициализировать из e‑mail до первого редактирования).
- **`avatar_url TEXT NULL`**
  - URL до аватара (в будущем — связь с файловым хранилищем).
- **`bio TEXT NULL`**
  - Краткое описание/о себе.
- **`locale TEXT NOT NULL DEFAULT 'en'`**
  - Язык интерфейса (например: `en`, `ru`).
- **`timezone TEXT NOT NULL DEFAULT 'UTC'`**
  - Часовой пояс (например: `Europe/Moscow`).
- **`created_at TIMESTAMPTZ NOT NULL DEFAULT now()`**
- **`updated_at TIMESTAMPTZ NOT NULL DEFAULT now()`**

Рекомендуемые индексы:

- PK по `user_id` уже даёт поиск по пользователю.
- При необходимости позже можно добавить индексы по `display_name`, если будет поиск по имени.

### Структура фичи `internal/features/users`

По аналогии с `auth`:

- `internal/features/users/domain.go`
  - `UserProfile` — доменная модель профиля.
  - Возможные вспомогательные типы (ролевая модель, настройки уведомлений и т.п. — позже).
- `internal/features/users/repository.go`
  - Интерфейс `ProfileRepository` для работы с `user_profiles`.
- `internal/features/users/repository_postgres.go`
  - Реализация `ProfileRepository` на PostgreSQL (`public.user_profiles`).
- `internal/features/users/service.go`
  - Бизнес‑логика:
    - получить профиль текущего пользователя,
    - создать профиль при первой авторизации,
    - обновить профиль.
- `internal/features/users/handler.go`
  - HTTP‑эндпоинты на Fiber.
- `internal/features/users/dto.go`
  - DTO для запросов/ответов (JSON‑структуры).

### API контракты для фронтенда (первый шаг)

Базовый набор, достаточно для MVP:

- **`GET /api/users/me`**
  - Требует авторизации (access‑токен).
  - Используется фронтендом для получения профиля текущего пользователя.
  - Выход (пример):
    - `userID`
    - `email` (из `auth.users`, по `user_id`)
    - `displayName`
    - `avatarURL`
    - `locale`
    - `timezone`

- **`PUT /api/users/me`**
  - Требует авторизации (access‑токен).
  - Изменяет поля профиля (но не e‑mail и не пароль).
  - Вход:
    - `displayName?`
    - `avatarURL?`
    - `bio?`
    - `locale?`
    - `timezone?`
  - Логика:
    - Берём `userID` из JWT (auth‑middleware).
    - Обновляем только переданные поля.
    - Возвращаем обновлённый профиль.

В дальнейшем (но не в шаге 2) можно добавить:

- `GET /api/users/:id` — просмотр чужих профилей (доступен по правилам авторизации).
- `GET /api/users` — поиск/листинг пользователей (для админки).
  - Пагинация: `page`, `pageSize`.
  - Фильтры:
    - `q` — строка поиска по `display_name`/email.
    - `role` — фильтр по роли (через auth‑интерфейс).
    - `isActive` — флаг активности пользователя (берётся из `auth.users.is_active`).
  - **Важно**: по умолчанию (если `isActive` не передан) возвращать только **активных** пользователей (`is_active = true`), чтобы не “поднимать” заблокированные/удалённые аккаунты в админке без явного запроса.

### Интеграция с фичей Auth

- **Auth отвечает за:**
  - регистрацию, логин, refresh, logout;
  - таблицы `auth.users` и `auth.refresh_tokens`;
  - JWT‑токены и middleware `auth`.
- **Users отвечает за:**
  - чтение/запись `public.user_profiles`;
  - бизнес‑правила профиля (какие поля можно менять, валидация значений и т.п.).

Взаимодействие:

- После успешного логина/регистрации (фича `auth`):
  - при необходимости можно вызывать метод сервиса `users` типа `EnsureProfileExists(userID)`, который:
    - создаёт профиль с дефолтными значениями, если его ещё нет,
    - ничего не делает, если профиль уже есть.
  - Это взаимодействие лучше оформить через **интерфейс**, чтобы `auth` знал только о "профильном сервисе", а не о конкретной реализации.
- В HTTP‑обработчиках `users` мы извлекаем `userID` из клеймов JWT (через общий middleware) и не лезем в базу `auth.users` напрямую.

### DTO и валидация

В `internal/features/users/dto.go`:

- `UserProfileResponse`:
  - `UserID string`
  - `Email string`
  - `DisplayName string`
  - `AvatarURL *string`
  - `Bio *string`
  - `Locale string`
  - `Timezone string`
- `UpdateProfileRequest`:
  - `DisplayName *string` `validate:"omitempty,min=2,max=100"`
  - `AvatarURL *string` (дополнительно можно валидировать URL на фронте/бэке)
  - `Bio *string` `validate:"omitempty,max=500"`
  - `Locale *string` `validate:"omitempty"`
  - `Timezone *string` `validate:"omitempty"`

В хендлерах использовать `github.com/go-playground/validator/v10`, как и в `auth`.

### TDD для фичи Users

Рекомендуемый порядок (минимум для шага 2):

1. **Unit‑тесты сервиса (`internal/features/users/service_test.go`)**:
   - `TestGetProfile_CreatesDefaultIfNotExists` — при первом запросе создаётся профиль с дефолтами.
   - `TestGetProfile_ReturnsExisting` — при наличии профиля просто возвращаем его.
   - `TestUpdateProfile_UpdatesOnlyProvidedFields` — проверка частичного обновления.
2. **Интеграционные тесты HTTP (`internal/features/users/handler_test.go`)**:
   - `GET /api/users/me` с валидным access‑токеном.
   - `PUT /api/users/me` c валидным телом и токеном.

Тесты должны помогать держать границы ответственности:

- `auth` → кто пользователь (идентификация и права).
- `users` → какие у него данные профиля и как их менять.

