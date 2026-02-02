# Архитектура проекта: Feature-First

## Обзор

Проект использует **Feature-First (Vertical Slice) архитектуру**, где код организован по бизнес-фичам, а не по техническим слоям. Каждая фича содержит весь необходимый код от domain моделей до HTTP handlers.

## Структура проекта

```
internal/
  shared/                    # Общие компоненты, используемые всеми фичами
    config/                  # Конфигурация приложения
    logger/                  # Логгер (slog)
    util/                    # Утилиты (UUID, errors, time)
    domain/                  # Общие domain ошибки
      errors.go
    http/                    # Общие HTTP компоненты
      middleware/            # Middleware (request_id, logger, cors, auth, authorize)
      response/              # Обёртка для ответов API
      validator/             # Валидация DTO
      router.go              # Настройка роутинга
    authutils/               # Утилиты безопасности (НЕ фича!)
      password.go            # Хэширование паролей (bcrypt)
      jwt.go                 # JWT токены (access/refresh)
      claims.go              # Структуры claims
    postgres/                # Подключение к БД
      db.go

  features/                  # Бизнес-фичи (Feature-First)
    auth/                    # Фича: Аутентификация и авторизация
      domain.go              # User, RefreshToken, константы ролей
      repository.go          # Интерфейсы репозиториев
      repository_postgres.go # Реализация репозиториев для PostgreSQL
      service.go             # Бизнес-логика (Register, Login, Refresh, Logout)
      handler.go             # HTTP handlers
      dto.go                 # DTO для запросов/ответов
      migrations/            # SQL миграции для этой фичи (опционально)
    users/                   # Фича: Управление пользователями
      domain.go
      repository.go
      service.go
      handler.go
      dto.go
    organizations/           # Фича: Организации (будущая)
      ...
    events/                  # Фича: События (будущая)
      ...
    schedule/                # Фича: Расписание (будущая)
      ...
    teams/                   # Фича: Команды и участники (будущая)
      ...
    assets/                  # Фича: Материалы (будущая)
      ...
```

## Принципы организации

### Shared компоненты

**Shared** (`internal/shared/`) содержит компоненты, которые используются несколькими фичами:

- **config**, **logger**, **util** — инфраструктурные утилиты
- **domain/errors.go** — общие ошибки (ErrNotFound, ErrUnauthorized, и т.д.)
- **http/** — общие HTTP компоненты (middleware, response обёртка, валидатор)
- **authutils/** — утилиты безопасности (password hashing, JWT), используются фичами
- **postgres/db.go** — подключение к БД

### Features компоненты

Каждая фича (`internal/features/{feature_name}/`) содержит:

1. **domain.go** — domain модели и бизнес-логика этой фичи
2. **repository.go** — интерфейсы репозиториев
3. **repository_postgres.go** — реализация репозиториев для PostgreSQL
4. **service.go** — бизнес-логика (use cases)
5. **handler.go** — HTTP handlers
6. **dto.go** — DTO для HTTP запросов/ответов

**Важно**: Каждая фича изолирована и может использоваться независимо от других.

## Преимущества Feature-First

### 1. Локальность
✅ Всё для одной фичи в одном месте
- Легко найти код
- Легко понять контекст
- Удобно для изучения Go (видна полная картина фичи)

### 2. Независимость фич
✅ Фичи не зависят друг от друга
- Легко добавить/удалить фичу
- Меньше риска сломать другие фичи
- Хорошо для MVP (гибкость)

### 3. Соответствие процессу разработки
✅ Структура соответствует процессу
- Реализуем фичу полностью: domain → repository → service → handler
- Всё для фичи находится в одном месте

### 4. Масштабируемость
✅ Легко выделить фичу в микросервис
- Вся логика фичи изолирована
- Хорошо для растущей команды

### 5. Тестируемость
✅ Каждую фичу можно тестировать отдельно
- Удобно писать unit-тесты
- Удобно писать интеграционные тесты

## Разделение ответственности

### Shared компоненты

- **config/** — чтение конфигурации из .env
- **logger/** — логирование (slog)
- **util/** — переиспользуемые утилиты (UUID, errors, time)
- **domain/errors.go** — общие ошибки домена
- **http/middleware/** — HTTP middleware (request_id, logger, cors, auth, authorize)
- **http/response/** — стандартизация ответов API
- **http/validator/** — валидация DTO
- **authutils/** — утилиты безопасности (password, JWT) — используются фичами
- **postgres/db.go** — подключение к PostgreSQL

### Features компоненты

Каждая фича реализует полный цикл:

1. **Domain** — бизнес-модели и правила
2. **Repository** — интерфейсы и реализация для работы с БД
3. **Service** — бизнес-логика (use cases)
4. **Handler** — HTTP обработчики
5. **DTO** — структуры для HTTP запросов/ответов

## Правила использования Shared

### Когда использовать Shared?

✅ Используйте `shared/` для:
- Компонентов, используемых **несколькими фичами**
- Инфраструктурных утилит (config, logger, util)
- Общих HTTP компонентов (middleware, response)
- Утилит безопасности (authutils) — используются фичами

❌ Не используйте `shared/` для:
- Кода, специфичного для одной фичи
- Domain моделей конкретной фичи (они в features/{feature}/)

### Примеры

```go
// ✅ Правильно: используем shared для общего компонента
import "wdpl_back/internal/shared/domain"
if err == domain.ErrNotFound { ... }

// ✅ Правильно: используем shared для утилиты
import "wdpl_back/internal/shared/util"
id := util.NewUUID()

// ✅ Правильно: используем shared для authutils
import "wdpl_back/internal/shared/authutils"
hash := authutils.HashPassword(password)

// ❌ Неправильно: не используем shared для фичи
// import "wdpl_back/internal/shared/auth"  // НЕТ! Используем features/auth
import "wdpl_back/internal/features/auth"
user, err := authService.Register(...)
```

## Зависимости между фичами

### Правило: фичи не зависят друг от друга напрямую

Если фича A нуждается в данных фичи B:

1. **Через общие модели в shared/domain/** — если модели действительно общие
2. **Через интерфейсы** — фича A определяет интерфейс, фича B его реализует
3. **Через shared/repository/** — если нужен общий доступ к данным

**Пример**: Если фича `events` нужна информация о пользователе:
- Можно использовать `features/auth` через интерфейс (предпочтительно)
- Или вынести User в `shared/domain/` если он используется многими фичами

## Миграции БД

Миграции могут быть:
- В `migrations/` (общая директория) — как сейчас
- Или в `features/{feature}/migrations/` — для изоляции

Для MVP используем общую директорию `migrations/` для простоты.

## Импорты

### Структура импортов

```go
package handler

import (
    // Стандартная библиотека
    "fmt"
    "time"
    
    // Внешние зависимости
    "github.com/gofiber/fiber/v2"
    
    // Shared компоненты
    "wdpl_back/internal/shared/domain"
    "wdpl_back/internal/shared/util"
    "wdpl_back/internal/shared/http/response"
    
    // Текущая фича
    "wdpl_back/internal/features/auth"
)
```

### Правила импортов

1. ✅ Фичи импортируют shared компоненты
2. ✅ Shared компоненты НЕ импортируют фичи
3. ❌ Фичи НЕ импортируют другие фичи напрямую (через интерфейсы)
4. ✅ Можно использовать общие модели через shared/domain/

## Пример: Фича Auth

```go
// internal/features/auth/domain.go
package auth

type User struct { ... }
type RefreshToken struct { ... }

// internal/features/auth/repository.go
package auth

type UserRepository interface { ... }

// internal/features/auth/repository_postgres.go
package auth

type userRepository struct { ... }

// internal/features/auth/service.go
package auth

type Service struct {
    userRepo UserRepository
    ...
}

// internal/features/auth/handler.go
package auth

type Handler struct {
    service *Service
}

// internal/features/auth/dto.go
package auth

type RegisterRequest struct { ... }
```

## Шаг 1: Модуль авторизации

Подробные рекомендации по реализации модуля авторизации (Go Fiber + миграция с Supabase, БД‑схема, API, TDD) вынесены в отдельный файл:

- `./steps/STEP1.md`

## Шаг 2: Фича Users (управление пользователями)

Рекомендации по фиче `users` (структура, БД‑схема профилей, API, TDD‑подход) описаны в:

- `./steps/STEP2.md`

## Шаг 3: Черновики событий

Домен черновиков событий и дней событий (`event_drafts`, `event_day_drafts`) описан в:

- `./steps/STEP3.md`

## Шаг 4: Публикованные события и публикация

Домен публичных событий (`events`, `event_days`) и операция публикации черновиков в публичный домен описаны в:

- `./steps/STEP4.md`
