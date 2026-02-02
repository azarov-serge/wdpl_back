# WDPL Backend на Supabase — MVP

## Оглавление

- [WDPL Backend на Supabase — MVP](#wdpl-backend-на-supabase--mvp)
  - [Оглавление](#оглавление)
  - [Архитектура решения](#архитектура-решения)
  - [Описание MVP](#описание-mvp)
    - [Цель проекта](#цель-проекта)
    - [Потребители](#потребители)
  - [Домены и модели данных](#домены-и-модели-данных)
    - [1. Аутентификация и пользователи](#1-аутентификация-и-пользователи)
      - [User (Пользователь)](#user-пользователь)
      - [RefreshToken (Токен обновления)](#refreshtoken-токен-обновления)
    - [2. События](#2-события)
      - [Event (Мероприятие)](#event-мероприятие)
      - [EventDraft (Черновик мероприятия)](#eventdraft-черновик-мероприятия)
    - [3. Расписание (оптимизированная структура)](#3-расписание-оптимизированная-структура)
      - [EventDay (Календарный день мероприятия) - JSONB структура](#eventday-календарный-день-мероприятия---jsonb-структура)
      - [EventDayDraft (Черновик дня события)](#eventdaydraft-черновик-дня-события)
      - [SessionType (Тип сессии)](#sessiontype-тип-сессии)
    - [4. Команды и участники](#4-команды-и-участники)
      - [Team (Команда)](#team-команда)
      - [Person (Персона/участник)](#person-персонаучастник)
      - [TeamMember (Состав команды)](#teammember-состав-команды)
    - [5. Материалы](#5-материалы)
      - [Asset (Вложения/материалы)](#asset-вложенияматериалы)
  - [Пошаговая реализация](#пошаговая-реализация)
    - [Этап 1: Настройка Supabase проекта](#этап-1-настройка-supabase-проекта)
      - [1.1. Создание проекта](#11-создание-проекта)
      - [1.2. Настройка клиентов](#12-настройка-клиентов)
    - [Этап 2: Создание структуры базы данных](#этап-2-создание-структуры-базы-данных)
      - [2.1. Создание таблиц](#21-создание-таблиц)
      - [2.2. Функции для вычисления метаданных event\_days](#22-функции-для-вычисления-метаданных-event_days)
      - [2.3. Функции валидации](#23-функции-валидации)
      - [2.4. Триггеры для автоматического обновления updated\_at](#24-триггеры-для-автоматического-обновления-updated_at)
    - [Этап 3: Row Level Security (RLS)](#этап-3-row-level-security-rls)
      - [3.1. Политики для users](#31-политики-для-users)
      - [3.2. Политики для events](#32-политики-для-events)
      - [3.3. Политики для event\_days](#33-политики-для-event_days)
    - [Этап 4: Реализация на клиентах](#этап-4-реализация-на-клиентах)
      - [4.1. React TypeScript - Типы](#41-react-typescript---типы)
      - [4.2. React TypeScript - Хуки](#42-react-typescript---хуки)
      - [4.3. Flutter - Модели](#43-flutter---модели)
    - [Этап 5: Supabase Storage для файлов](#этап-5-supabase-storage-для-файлов)
      - [5.1. Настройка Storage](#51-настройка-storage)
    - [Этап 6: Тестирование](#этап-6-тестирование)
      - [6.1. Проверка RLS политик](#61-проверка-rls-политик)
      - [6.2. Проверка JSONB структуры](#62-проверка-jsonb-структуры)
      - [6.3. Проверка real-time](#63-проверка-real-time)
  - [Ключевые преимущества архитектуры](#ключевые-преимущества-архитектуры)

## Архитектура решения

- **Backend**: Supabase (PostgreSQL + Auth + Storage + Real-time)
- **Frontend**: React TypeScript с SupabaseJS
- **Mobile**: Flutter с Supabase Flutter SDK
- **Нет отдельного Go backend** - вся логика на стороне Supabase и клиентов

## Описание MVP

### Цель проекта
Минимальный, но полноценный backend для управления событиями и расписанием WD Planning. Поддержка CRUD по мероприятиям, моделирование расписания (сессии, дни), команды/людей, вложения. Аутентификация и авторизация на основе JWT (access/refresh), возможность администратору включать/выключать пользователей.

### Потребители
- **Организаторы/администраторы** - создают события, управляют расписанием и доступами
- **Модераторы/редакторы** - добавляют/корректируют сессии, материалы
- **Участники/команды** - получают доступ к программе, материалам
- **Клиентские приложения** (веб/мобайл) - основной потребитель REST API через Supabase

## Домены и модели данных

### 1. Аутентификация и пользователи

#### User (Пользователь)
```typescript
interface User {
  id: string;                    // UUID
  email: string;                  // уникальная почта
  username: string;               // уникальный логин
  avatar_url?: string;            // ссылка на аватар (Supabase Storage)
  role: 'admin' | 'organizer' | 'viewer';
  is_active: boolean;             // администратор может включать/выключать
  created_at: string;
  updated_at: string;
}
```

**Таблица в Supabase:**
```sql
CREATE TABLE users (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  email       text UNIQUE NOT NULL,
  username    text UNIQUE NOT NULL,
  avatar_url  text,
  role        text NOT NULL CHECK (role IN ('admin','organizer','viewer')),
  is_active   boolean NOT NULL DEFAULT true,
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now()
);

-- Используем встроенную auth.users от Supabase
-- Дополнительные поля храним в public.users
```

#### RefreshToken (Токен обновления)
Управляется через Supabase Auth - не требует отдельной таблицы для MVP.

### 2. События

#### Event (Мероприятие)
```typescript
interface Event {
  id: string;
  title: string;
  description?: string;
  start_date: string;             // ISO date string (всегда по Москве)
  end_date: string;               // ISO date string (всегда по Москве)
  status: 'draft' | 'published' | 'archived';
  created_at: string;
  updated_at: string;
}
```

**Таблица:**
```sql
CREATE TABLE events (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  title           text NOT NULL,
  description     text,
  start_date      date NOT NULL,
  end_date        date NOT NULL,
  status          text NOT NULL CHECK (status IN ('draft','published','archived')),
  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now()
);
```

#### EventDraft (Черновик мероприятия)
```typescript
interface EventDraft {
  id: string;
  event_id?: string;              // UUID опубликованного события, если уже публиковали
  title: string;
  description?: string;
  start_date: string;
  end_date: string;
  status: 'draft';                // статус черновика, для явности
  published_at?: string | null;   // время последней публикации
  created_by: string;             // UUID автора
  created_at: string;
  updated_at: string;
}
```

**Таблица (черновики событий):**
```sql
CREATE TABLE event_drafts (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id     uuid REFERENCES events(id) ON DELETE SET NULL,
  title        text NOT NULL,
  description  text,
  start_date   date NOT NULL,
  end_date     date NOT NULL,
  status       text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft')),
  published_at timestamptz,
  created_by   uuid NOT NULL REFERENCES auth.users(id),
  created_at   timestamptz NOT NULL DEFAULT now(),
  updated_at   timestamptz NOT NULL DEFAULT now()
);
```

### 3. Расписание (оптимизированная структура)

#### EventDay (Календарный день мероприятия) - JSONB структура
```typescript
interface EventDay {
  id: string;
  event_id: string;
  date: string;                   // ISO date string
  schedule: DaySchedule;          // Вся структура дня в JSONB
  session_count: number;          // generated column
  first_session_start: string | null;
  last_session_end: string | null;
  created_at: string;
  updated_at: string;
}

interface DaySchedule {
  sessions: SessionData[];
}

interface SessionData {
  id: string;
  session_type_id: string;
  title: string;
  description?: string;
  starts_at: string;              // ISO timestamp
  ends_at: string;                // ISO timestamp
  is_public: boolean;
  status: 'planned' | 'published' | 'cancelled';
  location?: Location;
  participants: ParticipantReference[];
  teams: TeamReference[];
}

interface Location {
  type: 'physical' | 'online' | 'hybrid';
  address?: string;
  url?: string;
}

interface ParticipantReference {
  person_id: string;
  role: string;                   // speaker, moderator, jury, organizer
  person_name?: string;           // денормализация для быстрого отображения
}

interface TeamReference {
  team_id: string;
  purpose: string;                 // as_is, to_be, defense, workshop
  team_name?: string;             // денормализация
}
```

**Таблица:**
```sql
CREATE TABLE event_days (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id    uuid NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  date        date NOT NULL,
  
  -- Вся структура дня в JSONB (сессии, участники, команды)
  schedule    jsonb NOT NULL DEFAULT '{"sessions": []}'::jsonb,
  
  -- Метаданные для быстрого поиска (generated columns)
  session_count int GENERATED ALWAYS AS (
    jsonb_array_length(schedule->'sessions')
  ) STORED,
  
  first_session_start timestamptz GENERATED ALWAYS AS (
    CASE 
      WHEN jsonb_array_length(schedule->'sessions') > 0 
      THEN (schedule->'sessions'->0->>'starts_at')::timestamptz
      ELSE NULL
    END
  ) STORED,
  
  last_session_end timestamptz GENERATED ALWAYS AS (
    CASE 
      WHEN jsonb_array_length(schedule->'sessions') > 0 
      THEN (schedule->'sessions'->-1->>'ends_at')::timestamptz
      ELSE NULL
    END
  ) STORED,
  
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now(),
  
  UNIQUE (event_id, date)
);

-- Индексы
CREATE INDEX idx_event_days_event_date ON event_days (event_id, date);
CREATE INDEX idx_event_days_schedule_gin ON event_days USING GIN (schedule);
CREATE INDEX idx_event_days_time_range ON event_days (first_session_start, last_session_end);
```

#### EventDayDraft (Черновик дня события)
```typescript
interface EventDayDraft {
  id: string;
  event_id: string;               // UUID опубликованного события
  date: string;                   // ISO date string
  schedule: DaySchedule;
  published_at?: string | null;   // время последней публикации
  created_by: string;             // UUID автора
  created_at: string;
  updated_at: string;
}
```

**Таблица (черновики дней):**
```sql
CREATE TABLE event_days_drafts (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id     uuid NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  date         date NOT NULL,
  schedule     jsonb NOT NULL DEFAULT '{"sessions": []}'::jsonb,
  published_at timestamptz,
  created_by   uuid NOT NULL REFERENCES auth.users(id),
  created_at   timestamptz NOT NULL DEFAULT now(),
  updated_at   timestamptz NOT NULL DEFAULT now(),
  UNIQUE (event_id, date)
);

CREATE INDEX idx_event_days_drafts_event_date ON event_days_drafts (event_id, date);
CREATE INDEX idx_event_days_drafts_schedule_gin ON event_days_drafts USING GIN (schedule);
```

#### SessionType (Тип сессии)
```typescript
interface SessionType {
  id: string;
  code: string;                   // unique: break, coffee_break, lunch, intro, pitch, team_work, contest
  title: string;                  // человекочитаемое название
  created_at: string;
  updated_at: string;
}
```

**Таблица:**
```sql
CREATE TABLE session_types (
  id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  code       text NOT NULL UNIQUE,
  title      text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
```

### 4. Команды и участники

#### Team (Команда)
```typescript
interface Team {
  id: string;
  event_id: string;
  title: string;                  // "WATCHDOG", "Day2", "WDGX"
  created_at: string;
  updated_at: string;
}
```

**Таблица:**
```sql
CREATE TABLE teams (
  id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id   uuid NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  title      text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (event_id, title)
);
```

#### Person (Персона/участник)
```typescript
interface Person {
  id: string;
  full_name: string;
  email?: string;
  phone?: string;
  org_role?: string;              // должность/роль в орг.
  created_at: string;
  updated_at: string;
}
```

**Таблица:**
```sql
CREATE TABLE people (
  id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  full_name  text NOT NULL,
  email      text,
  phone      text,
  org_role   text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
```

#### TeamMember (Состав команды)
```typescript
interface TeamMember {
  id: string;
  team_id: string;
  person_id: string;
  role_in_team?: string;
  created_at: string;
  updated_at: string;
}
```

**Таблица:**
```sql
CREATE TABLE team_members (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  team_id     uuid NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
  person_id   uuid NOT NULL REFERENCES people(id) ON DELETE RESTRICT,
  role_in_team text,
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now(),
  UNIQUE (team_id, person_id)
);
```

### 5. Материалы

#### Asset (Вложения/материалы)
```typescript
interface Asset {
  id: string;
  event_id?: string;              // привязка к событию
  session_id?: string;             // привязка к сессии (из JSONB schedule)
  title: string;
  url: string;                     // ссылка на файл в Supabase Storage
  mime_type: string;
  created_at: string;
  updated_at: string;
}
```

**Таблица:**
```sql
CREATE TABLE assets (
  id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id   uuid REFERENCES events(id) ON DELETE CASCADE,
  session_id text,                -- UUID из JSONB schedule (не FK)
  title      text NOT NULL,
  url        text NOT NULL,       -- ссылка на Supabase Storage
  mime_type  text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CHECK (event_id IS NOT NULL OR session_id IS NOT NULL)
);

CREATE INDEX idx_assets_event ON assets(event_id);
CREATE INDEX idx_assets_session ON assets(session_id);
```

## Пошаговая реализация

### Этап 1: Настройка Supabase проекта

#### 1.1. Создание проекта
1. Создать проект на [supabase.com](https://supabase.com)
2. Сохранить credentials:
   - Project URL
   - Anon key
   - Service role key

#### 1.2. Настройка клиентов

**React:**
```bash
npm install @supabase/supabase-js @tanstack/react-query
```

```typescript
// lib/supabase.ts
import { createClient } from '@supabase/supabase-js';

const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL!;
const supabaseAnonKey = process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!;

export const supabase = createClient(supabaseUrl, supabaseAnonKey);
```

**Flutter:**
```yaml
# pubspec.yaml
dependencies:
  supabase_flutter: ^2.0.0
```

```dart
// lib/main.dart
import 'package:supabase_flutter/supabase_flutter.dart';

void main() async {
  await Supabase.initialize(
    url: 'YOUR_SUPABASE_URL',
    anonKey: 'YOUR_SUPABASE_ANON_KEY',
  );
  runApp(MyApp());
}
```

### Этап 2: Создание структуры базы данных

> **Примечание**: Если у вас уже есть данные, пропустите этот этап. Данный раздел предназначен для создания структуры БД с нуля.

#### 2.1. Создание таблиц

Выполнить SQL скрипты в Supabase Dashboard → SQL Editor в указанном порядке:

**Скрипт 1: Пользователи**
```sql
-- Создание таблицы users (дополняет auth.users)
CREATE TABLE public.users (
  id          uuid PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
  email       text UNIQUE NOT NULL,
  username    text UNIQUE NOT NULL,
  avatar_url  text,
  role        text NOT NULL CHECK (role IN ('admin','organizer','viewer')) DEFAULT 'viewer',
  is_active   boolean NOT NULL DEFAULT true,
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now()
);

-- Триггер для автоматического создания записи в public.users при регистрации
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS TRIGGER AS $$
BEGIN
  INSERT INTO public.users (id, email, username)
  VALUES (NEW.id, NEW.email, COALESCE(NEW.raw_user_meta_data->>'username', NEW.email));
  RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

CREATE TRIGGER on_auth_user_created
  AFTER INSERT ON auth.users
  FOR EACH ROW EXECUTE FUNCTION public.handle_new_user();
```

**Скрипт 2: События и расписание**
```sql
-- События (время всегда по Москве)
CREATE TABLE public.events (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  title           text NOT NULL,
  description     text,
  start_date      date NOT NULL,
  end_date        date NOT NULL,
  status          text NOT NULL CHECK (status IN ('draft','published','archived')) DEFAULT 'draft',
  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now()
);

-- Черновики событий (редактирование без публикации)
CREATE TABLE public.event_drafts (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id     uuid REFERENCES public.events(id) ON DELETE SET NULL,
  title        text NOT NULL,
  description  text,
  start_date   date NOT NULL,
  end_date     date NOT NULL,
  status       text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft')),
  published_at timestamptz,
  created_by   uuid NOT NULL REFERENCES auth.users(id),
  created_at   timestamptz NOT NULL DEFAULT now(),
  updated_at   timestamptz NOT NULL DEFAULT now()
);

-- Типы сессий
-- Справочник типов сессий для категоризации событий в расписании.
-- Каждая сессия в event_days.schedule ссылается на session_type_id.
-- 
-- Примеры возможных типов сессий:
--   - break: Перерыв
--   - coffee_break: Кофе-брейк
--   - lunch: Обед
--   - intro: Вводная сессия / Открытие
--   - keynote: Ключевое выступление
--   - pitch: Презентация / Питч
--   - team_work: Командная работа
--   - workshop: Воркшоп
--   - contest: Соревнование / Конкурс
--   - defense: Защита проекта
--   - networking: Нетворкинг
--   - closing: Закрытие мероприятия
--
-- Поля:
--   code: Уникальный код типа (используется в коде, например 'coffee_break')
--   title: Человекочитаемое название (например 'Кофе-брейк')
CREATE TABLE public.session_types (
  id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  code       text NOT NULL UNIQUE,
  title      text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- Дни событий (с JSONB структурой)
CREATE TABLE public.event_days (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id    uuid NOT NULL REFERENCES public.events(id) ON DELETE CASCADE,
  date        date NOT NULL,
  schedule    jsonb NOT NULL DEFAULT '{"sessions": []}'::jsonb,
  session_count int,
  first_session_start timestamptz,
  last_session_end timestamptz,
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now(),
  UNIQUE (event_id, date)
);

CREATE INDEX idx_event_days_event_date ON public.event_days (event_id, date);
CREATE INDEX idx_event_days_schedule_gin ON public.event_days USING GIN (schedule);
CREATE INDEX idx_event_days_time_range ON public.event_days (first_session_start, last_session_end);

-- Черновики дней событий (редактирование без публикации)
CREATE TABLE public.event_days_drafts (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id     uuid NOT NULL REFERENCES public.events(id) ON DELETE CASCADE,
  date         date NOT NULL,
  schedule     jsonb NOT NULL DEFAULT '{"sessions": []}'::jsonb,
  published_at timestamptz,
  created_by   uuid NOT NULL REFERENCES auth.users(id),
  created_at   timestamptz NOT NULL DEFAULT now(),
  updated_at   timestamptz NOT NULL DEFAULT now(),
  UNIQUE (event_id, date)
);

CREATE INDEX idx_event_days_drafts_event_date ON public.event_days_drafts (event_id, date);
CREATE INDEX idx_event_days_drafts_schedule_gin ON public.event_days_drafts USING GIN (schedule);
```

**Скрипт 3: Команды и участники**
```sql
-- Команды
CREATE TABLE public.teams (
  id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id   uuid NOT NULL REFERENCES public.events(id) ON DELETE CASCADE,
  title      text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (event_id, title)
);

-- Персоны
CREATE TABLE public.people (
  id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  full_name  text NOT NULL,
  email      text,
  phone      text,
  org_role   text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- Состав команды
CREATE TABLE public.team_members (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  team_id     uuid NOT NULL REFERENCES public.teams(id) ON DELETE CASCADE,
  person_id   uuid NOT NULL REFERENCES public.people(id) ON DELETE RESTRICT,
  role_in_team text,
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now(),
  UNIQUE (team_id, person_id)
);
```

**Скрипт 4: Материалы**
```sql
CREATE TABLE public.assets (
  id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id   uuid REFERENCES public.events(id) ON DELETE CASCADE,
  session_id text,                -- UUID из JSONB schedule
  title      text NOT NULL,
  url        text NOT NULL,       -- ссылка на Supabase Storage
  mime_type  text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CHECK (event_id IS NOT NULL OR session_id IS NOT NULL)
);

CREATE INDEX idx_assets_event ON public.assets(event_id);
CREATE INDEX idx_assets_session ON public.assets(session_id);
```

#### 2.2. Функции для вычисления метаданных event_days

```sql
-- Функция для вычисления количества сессий
CREATE OR REPLACE FUNCTION public.get_session_count(schedule_data jsonb)
RETURNS int AS $$
BEGIN
  IF NOT (schedule_data ? 'sessions') THEN
    RETURN 0;
  END IF;
  RETURN COALESCE(jsonb_array_length(schedule_data->'sessions'), 0);
END;
$$ LANGUAGE plpgsql;

-- Функция для получения времени начала первой сессии
CREATE OR REPLACE FUNCTION public.get_first_session_start(schedule_data jsonb)
RETURNS timestamptz AS $$
DECLARE
  sessions jsonb;
  first_session jsonb;
BEGIN
  IF NOT (schedule_data ? 'sessions') THEN
    RETURN NULL;
  END IF;
  sessions := schedule_data->'sessions';
  IF jsonb_array_length(sessions) = 0 THEN
    RETURN NULL;
  END IF;
  first_session := sessions->0;
  IF first_session IS NULL OR NOT (first_session ? 'starts_at') THEN
    RETURN NULL;
  END IF;
  RETURN (first_session->>'starts_at')::timestamptz;
END;
$$ LANGUAGE plpgsql;

-- Функция для получения времени окончания последней сессии
CREATE OR REPLACE FUNCTION public.get_last_session_end(schedule_data jsonb)
RETURNS timestamptz AS $$
DECLARE
  sessions jsonb;
  arr_len int;
  last_session jsonb;
BEGIN
  IF NOT (schedule_data ? 'sessions') THEN
    RETURN NULL;
  END IF;
  sessions := schedule_data->'sessions';
  arr_len := jsonb_array_length(sessions);
  IF arr_len = 0 THEN
    RETURN NULL;
  END IF;
  -- Получаем последний элемент через подзапрос (надежный способ)
  SELECT elem.value INTO last_session
  FROM jsonb_array_elements(sessions) WITH ORDINALITY AS elem(value, idx)
  ORDER BY elem.idx DESC
  LIMIT 1;
  
  IF last_session IS NULL OR NOT (last_session ? 'ends_at') THEN
    RETURN NULL;
  END IF;
  RETURN (last_session->>'ends_at')::timestamptz;
END;
$$ LANGUAGE plpgsql;

-- Триггер для обновления метаданных event_days
CREATE OR REPLACE FUNCTION public.update_event_day_metadata()
RETURNS TRIGGER AS $$
BEGIN
  NEW.session_count := public.get_session_count(NEW.schedule);
  NEW.first_session_start := public.get_first_session_start(NEW.schedule);
  NEW.last_session_end := public.get_last_session_end(NEW.schedule);
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_event_day_metadata_trigger
  BEFORE INSERT OR UPDATE ON public.event_days
  FOR EACH ROW
  EXECUTE FUNCTION public.update_event_day_metadata();
```

#### 2.3. Функции валидации

```sql
-- Валидация структуры schedule
CREATE OR REPLACE FUNCTION public.validate_schedule(schedule_data jsonb)
RETURNS boolean AS $$
DECLARE
  session jsonb;
BEGIN
  IF NOT (schedule_data ? 'sessions') THEN
    RETURN false;
  END IF;

  IF jsonb_typeof(schedule_data->'sessions') != 'array' THEN
    RETURN false;
  END IF;

  -- Итерация по элементам массива sessions
  FOR session IN SELECT value FROM jsonb_array_elements(schedule_data->'sessions')
  LOOP
    IF NOT (session ? 'id' AND
            session ? 'session_type_id' AND
            session ? 'title' AND
            session ? 'starts_at' AND
            session ? 'ends_at') THEN
      RETURN false;
    END IF;

    IF (session->>'starts_at')::timestamptz >= 
       (session->>'ends_at')::timestamptz THEN
      RETURN false;
    END IF;
  END LOOP;

  RETURN true;
END;
$$ LANGUAGE plpgsql;

-- Триггер валидации
CREATE OR REPLACE FUNCTION public.check_schedule_validity()
RETURNS TRIGGER AS $$
BEGIN
  IF NOT public.validate_schedule(NEW.schedule) THEN
    RAISE EXCEPTION 'Invalid schedule structure';
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER validate_event_day_schedule
  BEFORE INSERT OR UPDATE ON public.event_days
  FOR EACH ROW
  EXECUTE FUNCTION public.check_schedule_validity();
```

#### 2.4. Триггеры для автоматического обновления updated_at

```sql
-- Универсальная функция для обновления updated_at
CREATE OR REPLACE FUNCTION public.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Триггеры для всех таблиц с updated_at
CREATE TRIGGER update_users_updated_at
  BEFORE UPDATE ON public.users
  FOR EACH ROW
  EXECUTE FUNCTION public.update_updated_at_column();

CREATE TRIGGER update_events_updated_at
  BEFORE UPDATE ON public.events
  FOR EACH ROW
  EXECUTE FUNCTION public.update_updated_at_column();

CREATE TRIGGER update_event_drafts_updated_at
  BEFORE UPDATE ON public.event_drafts
  FOR EACH ROW
  EXECUTE FUNCTION public.update_updated_at_column();

CREATE TRIGGER update_session_types_updated_at
  BEFORE UPDATE ON public.session_types
  FOR EACH ROW
  EXECUTE FUNCTION public.update_updated_at_column();

CREATE TRIGGER update_event_days_updated_at
  BEFORE UPDATE ON public.event_days
  FOR EACH ROW
  EXECUTE FUNCTION public.update_updated_at_column();

CREATE TRIGGER update_event_days_drafts_updated_at
  BEFORE UPDATE ON public.event_days_drafts
  FOR EACH ROW
  EXECUTE FUNCTION public.update_updated_at_column();

CREATE TRIGGER update_teams_updated_at
  BEFORE UPDATE ON public.teams
  FOR EACH ROW
  EXECUTE FUNCTION public.update_updated_at_column();

CREATE TRIGGER update_people_updated_at
  BEFORE UPDATE ON public.people
  FOR EACH ROW
  EXECUTE FUNCTION public.update_updated_at_column();

CREATE TRIGGER update_team_members_updated_at
  BEFORE UPDATE ON public.team_members
  FOR EACH ROW
  EXECUTE FUNCTION public.update_updated_at_column();

CREATE TRIGGER update_assets_updated_at
  BEFORE UPDATE ON public.assets
  FOR EACH ROW
  EXECUTE FUNCTION public.update_updated_at_column();
```

### Этап 3: Row Level Security (RLS)

#### 3.1. Политики для users

```sql
ALTER TABLE public.users ENABLE ROW LEVEL SECURITY;

-- Пользователи могут видеть свой профиль
CREATE POLICY "Users can view own profile"
  ON public.users FOR SELECT
  USING (auth.uid() = id);

-- Админы могут видеть всех пользователей
CREATE POLICY "Admins can view all users"
  ON public.users FOR SELECT
  USING (
    EXISTS (
      SELECT 1 FROM public.users
      WHERE id = auth.uid() AND role = 'admin' AND is_active = true
    )
  );

-- Пользователи могут обновлять свой профиль
CREATE POLICY "Users can update own profile"
  ON public.users FOR UPDATE
  USING (auth.uid() = id)
  WITH CHECK (auth.uid() = id);

-- Только админы могут обновлять роль и активность
CREATE POLICY "Admins can update user roles"
  ON public.users FOR UPDATE
  USING (
    EXISTS (
      SELECT 1 FROM public.users
      WHERE id = auth.uid() AND role = 'admin' AND is_active = true
    )
  );
```

#### 3.2. Политики для events

```sql
ALTER TABLE public.events ENABLE ROW LEVEL SECURITY;

-- Аутентифицированные пользователи видят только опубликованные события
CREATE POLICY "Authenticated users can view published events"
  ON public.events FOR SELECT
  USING (
    auth.role() = 'authenticated'
    AND status = 'published'
  );

-- Организаторы и админы видят все события
CREATE POLICY "Organizers and admins can view all events"
  ON public.events FOR SELECT
  USING (
    EXISTS (
      SELECT 1 FROM public.users
      WHERE id = auth.uid()
        AND role IN ('admin', 'organizer')
        AND is_active = true
    )
  );

-- Организаторы и админы могут создавать события
CREATE POLICY "Organizers and admins can create events"
  ON public.events FOR INSERT
  WITH CHECK (
    EXISTS (
      SELECT 1 FROM public.users
      WHERE id = auth.uid()
        AND role IN ('admin', 'organizer')
        AND is_active = true
    )
  );

-- Организаторы и админы могут обновлять события
CREATE POLICY "Organizers and admins can update events"
  ON public.events FOR UPDATE
  USING (
    EXISTS (
      SELECT 1 FROM public.users
      WHERE id = auth.uid()
        AND role IN ('admin', 'organizer')
        AND is_active = true
    )
  );

-- Только админы могут удалять события
CREATE POLICY "Admins can delete events"
  ON public.events FOR DELETE
  USING (
    EXISTS (
      SELECT 1 FROM public.users
      WHERE id = auth.uid()
        AND role = 'admin'
        AND is_active = true
    )
  );

ALTER TABLE public.event_drafts ENABLE ROW LEVEL SECURITY;

-- Организаторы и админы могут читать черновики событий
CREATE POLICY "Organizers and admins can view event drafts"
  ON public.event_drafts FOR SELECT
  USING (
    EXISTS (
      SELECT 1 FROM public.users
      WHERE id = auth.uid()
        AND role IN ('admin', 'organizer')
        AND is_active = true
    )
  );

-- Организаторы и админы могут создавать черновики событий
CREATE POLICY "Organizers and admins can create event drafts"
  ON public.event_drafts FOR INSERT
  WITH CHECK (
    EXISTS (
      SELECT 1 FROM public.users
      WHERE id = auth.uid()
        AND role IN ('admin', 'organizer')
        AND is_active = true
    )
  );

-- Организаторы и админы могут обновлять черновики событий
CREATE POLICY "Organizers and admins can update event drafts"
  ON public.event_drafts FOR UPDATE
  USING (
    EXISTS (
      SELECT 1 FROM public.users
      WHERE id = auth.uid()
        AND role IN ('admin', 'organizer')
        AND is_active = true
    )
  );

-- Организаторы и админы могут удалять черновики событий
CREATE POLICY "Organizers and admins can delete event drafts"
  ON public.event_drafts FOR DELETE
  USING (
    EXISTS (
      SELECT 1 FROM public.users
      WHERE id = auth.uid()
        AND role IN ('admin', 'organizer')
        AND is_active = true
    )
  );
```

#### 3.3. Политики для event_days

```sql
ALTER TABLE public.event_days ENABLE ROW LEVEL SECURITY;

-- Аутентифицированные пользователи видят дни только опубликованных событий
CREATE POLICY "Authenticated users can view event days for published events"
  ON public.event_days FOR SELECT
  USING (
    auth.role() = 'authenticated'
    AND EXISTS (
      SELECT 1 FROM public.events
      WHERE id = event_days.event_id
        AND status = 'published'
    )
  );

-- Организаторы и админы видят все дни событий
CREATE POLICY "Organizers and admins can view event days"
  ON public.event_days FOR SELECT
  USING (
    EXISTS (
      SELECT 1 FROM public.users
      WHERE id = auth.uid()
        AND role IN ('admin', 'organizer')
        AND is_active = true
    )
  );

-- Организаторы и админы могут создавать дни
CREATE POLICY "Organizers and admins can create event days"
  ON public.event_days FOR INSERT
  WITH CHECK (
    EXISTS (
      SELECT 1 FROM public.users
      WHERE id = auth.uid()
        AND role IN ('admin', 'organizer')
        AND is_active = true
    )
  );

-- Организаторы и админы могут обновлять дни
CREATE POLICY "Organizers and admins can update event days"
  ON public.event_days FOR UPDATE
  USING (
    EXISTS (
      SELECT 1 FROM public.users
      WHERE id = auth.uid()
        AND role IN ('admin', 'organizer')
        AND is_active = true
    )
  );

-- Только админы могут удалять дни
CREATE POLICY "Admins can delete event days"
  ON public.event_days FOR DELETE
  USING (
    EXISTS (
      SELECT 1 FROM public.users
      WHERE id = auth.uid()
        AND role = 'admin'
        AND is_active = true
    )
  );

ALTER TABLE public.event_days_drafts ENABLE ROW LEVEL SECURITY;

-- Организаторы и админы могут читать черновики дней
CREATE POLICY "Organizers and admins can view event day drafts"
  ON public.event_days_drafts FOR SELECT
  USING (
    EXISTS (
      SELECT 1 FROM public.users
      WHERE id = auth.uid()
        AND role IN ('admin', 'organizer')
        AND is_active = true
    )
  );

-- Организаторы и админы могут создавать черновики дней
CREATE POLICY "Organizers and admins can create event day drafts"
  ON public.event_days_drafts FOR INSERT
  WITH CHECK (
    EXISTS (
      SELECT 1 FROM public.users
      WHERE id = auth.uid()
        AND role IN ('admin', 'organizer')
        AND is_active = true
    )
  );

-- Организаторы и админы могут обновлять черновики дней
CREATE POLICY "Organizers and admins can update event day drafts"
  ON public.event_days_drafts FOR UPDATE
  USING (
    EXISTS (
      SELECT 1 FROM public.users
      WHERE id = auth.uid()
        AND role IN ('admin', 'organizer')
        AND is_active = true
    )
  );

-- Организаторы и админы могут удалять черновики дней
CREATE POLICY "Organizers and admins can delete event day drafts"
  ON public.event_days_drafts FOR DELETE
  USING (
    EXISTS (
      SELECT 1 FROM public.users
      WHERE id = auth.uid()
        AND role IN ('admin', 'organizer')
        AND is_active = true
    )
  );
```

### Этап 4: Реализация на клиентах

#### 4.1. React TypeScript - Типы

```typescript
// types/index.ts
export interface User {
  id: string;
  email: string;
  username: string;
  avatar_url?: string;
  role: 'admin' | 'organizer' | 'viewer';
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Event {
  id: string;
  title: string;
  description?: string;
  start_date: string;  // всегда по Москве
  end_date: string;    // всегда по Москве
  status: 'draft' | 'published' | 'archived';
  created_at: string;
  updated_at: string;
}

export interface EventDraft {
  id: string;
  event_id?: string;
  title: string;
  description?: string;
  start_date: string;
  end_date: string;
  status: 'draft';
  published_at?: string | null;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface EventDay {
  id: string;
  event_id: string;
  date: string;
  schedule: DaySchedule;
  session_count: number;
  first_session_start: string | null;
  last_session_end: string | null;
  created_at: string;
  updated_at: string;
}

export interface EventDayDraft {
  id: string;
  event_id: string;
  date: string;
  schedule: DaySchedule;
  published_at?: string | null;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface DaySchedule {
  sessions: SessionData[];
}

export interface SessionData {
  id: string;
  session_type_id: string;
  title: string;
  description?: string;
  starts_at: string;
  ends_at: string;
  is_public: boolean;
  status: 'planned' | 'published' | 'cancelled';
  location?: Location;
  participants: ParticipantReference[];
  teams: TeamReference[];
}

export interface Location {
  type: 'physical' | 'online' | 'hybrid';
  address?: string;
  url?: string;
}

export interface ParticipantReference {
  person_id: string;
  role: string;
  person_name?: string;
}

export interface TeamReference {
  team_id: string;
  purpose: string;
  team_name?: string;
}
```

#### 4.2. React TypeScript - Хуки

```typescript
// hooks/useEventDay.ts
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { supabase } from '@/lib/supabase';
import type { EventDay, DaySchedule } from '@/types';

export function useEventDay(dayId: string) {
  return useQuery({
    queryKey: ['event-day', dayId],
    queryFn: async () => {
      const { data, error } = await supabase
        .from('event_days')
        .select('*')
        .eq('id', dayId)
        .single();

      if (error) throw error;
      return data as EventDay;
    },
  });
}

export function useCreateEventDay(eventId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: { date: string; schedule: DaySchedule }) => {
      const { data: day, error } = await supabase
        .from('event_days')
        .insert({
          event_id: eventId,
          date: data.date,
          schedule: data.schedule,
        })
        .select()
        .single();

      if (error) throw error;
      return day as EventDay;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['event-days', eventId] });
    },
  });
}
```

#### 4.3. Flutter - Модели

```dart
// models/event_day.dart
class EventDay {
  final String id;
  final String eventId;
  final DateTime date;
  final DaySchedule schedule;
  final int sessionCount;
  final DateTime? firstSessionStart;
  final DateTime? lastSessionEnd;
  final DateTime createdAt;
  final DateTime updatedAt;

  EventDay({
    required this.id,
    required this.eventId,
    required this.date,
    required this.schedule,
    required this.sessionCount,
    this.firstSessionStart,
    this.lastSessionEnd,
    required this.createdAt,
    required this.updatedAt,
  });

  factory EventDay.fromJson(Map<String, dynamic> json) {
    return EventDay(
      id: json['id'],
      eventId: json['event_id'],
      date: DateTime.parse(json['date']),
      schedule: DaySchedule.fromJson(json['schedule']),
      sessionCount: json['session_count'] ?? 0,
      firstSessionStart: json['first_session_start'] != null
          ? DateTime.parse(json['first_session_start'])
          : null,
      lastSessionEnd: json['last_session_end'] != null
          ? DateTime.parse(json['last_session_end'])
          : null,
      createdAt: DateTime.parse(json['created_at']),
      updatedAt: DateTime.parse(json['updated_at']),
    );
  }
}
```

### Этап 5: Supabase Storage для файлов

#### 5.1. Настройка Storage

```sql
-- Создать bucket для файлов
INSERT INTO storage.buckets (id, name, public)
VALUES ('event-assets', 'event-assets', true);

-- Политика доступа
CREATE POLICY "Users can upload assets"
  ON storage.objects FOR INSERT
  WITH CHECK (
    bucket_id = 'event-assets' AND
    auth.role() = 'authenticated'
  );

CREATE POLICY "Public can view assets"
  ON storage.objects FOR SELECT
  USING (bucket_id = 'event-assets');
```

### Этап 6: Тестирование

#### 6.1. Проверка RLS политик
- Тестировать доступ для разных ролей
- Проверить создание/чтение/обновление/удаление

#### 6.2. Проверка JSONB структуры
- Создание дня с несколькими сессиями
- Обновление расписания
- Валидация структуры

#### 6.3. Проверка real-time
- Подписки на изменения в React
- Подписки на изменения в Flutter

## Ключевые преимущества архитектуры

1. **Минимум запросов** - один запрос для создания/чтения всего дня
2. **Атомарность** - весь день создается/обновляется целиком
3. **Безопасность** - RLS на уровне БД
4. **Real-time** - автоматическая синхронизация между клиентами
5. **Производительность** - JSONB + GIN индексы для быстрого поиска
6. **Масштабируемость** - Supabase управляет инфраструктурой
