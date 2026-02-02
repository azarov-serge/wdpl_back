-- Основные таблицы для событий и черновиков событий.

CREATE TABLE IF NOT EXISTS public.events (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status TEXT NOT NULL DEFAULT 'published', -- published | archived
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Черновики событий: published_at, created_by, event_id ON DELETE SET NULL.
CREATE TABLE IF NOT EXISTS public.event_drafts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NULL REFERENCES public.events (id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    description TEXT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    published_at TIMESTAMPTZ NULL,
    created_by UUID NOT NULL REFERENCES auth.users (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Дни событий (публичное расписание, JSONB).
CREATE TABLE IF NOT EXISTS public.event_days (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES public.events (id) ON DELETE CASCADE,
    date DATE NOT NULL,
    schedule JSONB NOT NULL DEFAULT '{"sessions": []}'::jsonb,
    session_count INT NULL,
    first_session_start TIMESTAMPTZ NULL,
    last_session_end TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (event_id, date)
);

CREATE INDEX IF NOT EXISTS idx_event_days_event_date ON public.event_days (event_id, date);

-- Черновики дней событий (редактирование без публикации).
CREATE TABLE IF NOT EXISTS public.event_days_drafts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES public.events (id) ON DELETE CASCADE,
    date DATE NOT NULL,
    schedule JSONB NOT NULL DEFAULT '{"sessions": []}'::jsonb,
    published_at TIMESTAMPTZ NULL,
    created_by UUID NOT NULL REFERENCES auth.users (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (event_id, date)
);

CREATE INDEX IF NOT EXISTS idx_event_days_drafts_event_date ON public.event_days_drafts (event_id, date);

