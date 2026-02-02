-- Приведение event_drafts к схеме (если таблица создана старым 003 без published_at, created_by).
-- Удаляем колонку status (константа "draft" убрана из модели).

ALTER TABLE public.event_drafts DROP COLUMN IF EXISTS status;
ALTER TABLE public.event_drafts ADD COLUMN IF NOT EXISTS published_at TIMESTAMPTZ NULL;
ALTER TABLE public.event_drafts ADD COLUMN IF NOT EXISTS created_by UUID NULL REFERENCES auth.users (id) ON DELETE CASCADE;

-- Для существующих строк без автора: не делаем SET NOT NULL. В новом коде created_by заполняется из JWT.
