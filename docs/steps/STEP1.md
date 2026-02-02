## Шаг 1: Модуль авторизации (Go Fiber + миграция с Supabase)

### Цель шага

- **Заменить Supabase Auth** на собственный модуль авторизации на Go Fiber.
- **Сохранять UX фронтенда** (контракты уровня `checkAuth`, `signIn`, `signUp`, `signOut`, `refreshToken`).
- **Подготовить базу** для дальнейших фич (users, organizations и т.д.).

### Схема БД и размещение таблиц

- **Схема**: `auth` (отдельная от `public`) для явного разграничения прав и безопасности.
  - В дальнейшем бизнес‑данные (events, schedule и т.п.) можно хранить в `public` или `core`, но **учетные записи и токены** — в `auth`.
- Миграции для авторизации:
  - На старте достаточно хранить миграции в общей директории `migrations/` (как описано в `ARCHITECTURE.md`).
  - При усложнении можно вынести в `internal/features/auth/migrations/`.

#### Таблица `auth.users`

Основная таблица пользователей (аналог `auth.users` в Supabase, но упрощённый под наш кейс):

- **`id UUID PRIMARY KEY`** — внутренний идентификатор пользователя.
- **`email TEXT UNIQUE NOT NULL`** — уникальный e‑mail (ключ логина).
- **`password_hash TEXT NOT NULL`** — хэш пароля (bcrypt, `shared/authutils/password.go`).
- **`role TEXT NOT NULL DEFAULT 'user'`** — роль пользователя (например: `user`, `admin`).
- **`is_active BOOLEAN NOT NULL DEFAULT TRUE`** — мягкая деактивация учётки.
- **`supabase_user_id UUID NULL`** — опциональная ссылка на старый `supabase.auth.users.id` для миграции.
- **`created_at TIMESTAMPTZ NOT NULL DEFAULT now()`**.
- **`updated_at TIMESTAMPTZ NOT NULL DEFAULT now()`**.

Рекомендуется повесить:

- **UNIQUE(email)**.
- Индекс по **`(supabase_user_id)`** для миграции и отладки.

#### Таблица `auth.refresh_tokens`

Для поддержки `refreshToken` и безопасного logout:

- **`id UUID PRIMARY KEY`** — идентификатор записи токена.
- **`user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE`**.
- **`token TEXT NOT NULL`** — строка refresh‑токена (можно хранить как случайный UUID или opaque‑строку).
- **`expires_at TIMESTAMPTZ NOT NULL`** — время истечения.
- **`revoked_at TIMESTAMPTZ NULL`** — время отзыва (logout / security event).
- **`user_agent TEXT NULL`** — опционально, для аудита (из HTTP заголовков).
- **`ip TEXT NULL`** — опционально, IP клиента.
- **`created_at TIMESTAMPTZ NOT NULL DEFAULT now()`**.

Рекомендуется:

- Индекс по **`(user_id)`**.
- Индекс по **`(token)`**.
- Периодический job (cron/worker) для чистки просроченных токенов.

### API контракты для фронтенда

Цель — сохранить интерфейс на фронте:

- `checkAuth(): Promise<boolean>`
- `signIn(config?): Promise<T>`
- `signUp(config?): Promise<T>`
- `signOut(): Promise<void>`
- `refreshToken(args?): Promise<void>`

Рекомендуемые HTTP‑эндпоинты (Fiber handlers в `features/auth/handler.go`):

- **`POST /api/auth/sign-up`**
  - Вход: `email`, `password`, опционально дополнительные поля профиля (если нужны сразу).
  - Выход: базовая инфа о пользователе + access/refresh токены (или только access, а refresh в cookie).

- **`POST /api/auth/sign-in`**
  - Вход: `email`, `password`.
  - Выход: пользователь + access/refresh токены.

- **`POST /api/auth/sign-out`**
  - Вход: опционально refresh‑токен (если в теле, cookie или заголовке).
  - Действие: пометка refresh‑токена как `revoked_at = now()`.

- **`POST /api/auth/refresh`**
  - Вход: refresh‑токен.
  - Действие: валидация, проверка в `auth.refresh_tokens`, выдача нового access‑токена (и при необходимости нового refresh‑токена).

Таким образом, фронтенд может маппить свои функции:

- `signUp` → `POST /api/auth/sign-up`.
- `signIn` → `POST /api/auth/sign-in`.
- `signOut` → `POST /api/auth/sign-out`.
- `refreshToken` → `POST /api/auth/refresh`.
- `checkAuth` → `POST /api/auth/refresh` (простая реализация: если refresh успешен — считаем, что пользователь залогинен; если 401 — считаем, что нет).

### Интеграция с Go Fiber и shared/authutils

- Используем **`github.com/gofiber/fiber/v2`** для HTTP‑слоя:
  - Регистрация роутов в `internal/shared/http/router.go`.
  - Подключение middleware `auth` и `authorize` из `shared/http/middleware/`.
- Используем **`golang-jwt/jwt/v5`** в `shared/authutils/jwt.go`:
  - Генерация access/refresh JWT (или только access, если refresh хранится как opaque в БД).
  - Структуры claims в `shared/authutils/claims.go` (минимум: `UserID`, `Role`, `Exp`).
- В `features/auth/service.go`:
  - `Register` — создаёт пользователя, хэширует пароль, возвращает токены.
  - `Login` — проверяет пароль, создаёт refresh‑токен в БД, возвращает токены.
  - `Refresh` — валидирует refresh‑токен, проверяет в БД, обновляет/создаёт запись, возвращает новые токены.
  - `Logout` — помечает refresh‑токен как `revoked`.

Комментарии в коде писать на русском и только там, где решение неочевидно (например, про безопасность, инварианты и особенности миграции).

### Стратегия миграции с Supabase

1. **Считать Supabase источником правды** на время миграции.
2. **Создать таблицы `auth.users` и `auth.refresh_tokens`** с миграциями в текущем проекте.
3. **Импортировать пользователей из Supabase**:
   - Временный скрипт (Go или SQL), который читает `supabase.auth.users` и вставляет в `auth.users`.
   - Сохранять `supabase_user_id` для трассировки.
   - Если возможно, перенести существующие password_hash (если формат совместим) — иначе форсировать reset пароля.
4. **Переключить фронтенд** с Supabase SDK на наши REST‑эндпоинты (контракты, описанные выше).
5. **Постепенно удалить Supabase Auth** из конфигурации, оставив только PostgreSQL как БД.

### TDD для модуля авторизации

Для практики **Test-Driven Development** рекомендуется:

1. **Unit‑тесты сервиса (`features/auth/service_test.go`)**:
   - `TestRegister_Success/EmailAlreadyExists`.
   - `TestLogin_Success/WrongPassword/UserNotFound`.
   - `TestRefresh_Success/RevokedToken/ExpiredToken`.
   - `TestLogout_Success`.
2. **Интеграционные тесты HTTP‑эндпоинтов** (Fiber + тестовая БД, например, с использованием `testcontainers` или in‑memory PostgreSQL):
   - Проверка, что эндпоинты соответствуют контрактам фронтенда и правильно работают с JWT и refresh‑токенами.
3. При добавлении новой логики (ролей, прав доступа) — сначала писать тесты для бизнес‑правил, затем реализацию.

