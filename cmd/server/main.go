// Package main — точка входа HTTP‑сервера. Только инициализация и порядок запуска; бизнес-логика в internal.
package main

import (
	"os"

	"github.com/joho/godotenv"

	"wdpl_back/internal/shared/config"
	"wdpl_back/internal/shared/http/router"
	"wdpl_back/internal/shared/logger"
	"wdpl_back/internal/shared/postgres"
)

func main() {
	// Загружаем .env только в dev. В Docker задаём APP_ENV=production — тогда используются переменные контейнера.
	// Локально удобно хранить секреты в .env и не задавать их вручную.
	// В Docker переменные задаются через environment / env в конфиге контейнера,
	// файла .env в образе обычно нет (и не должен быть из соображений безопасности).
	// Условие по APP_ENV как раз разделяет: в dev подгружаем .env,
	// в production полагаемся только на переменные окружения контейнера.
	if env := os.Getenv("APP_ENV"); env != "production" && env != "prod" {
		_ = godotenv.Load()
	}

	// Отсутствие обязательных переменных (DATABASE_URL, JWT_SECRET, REFRESH_SECRET)
	cfg := config.MustLoad()

	logg := logger.New(cfg)

	db, err := postgres.Connect(cfg, logg)
	if err != nil {
		logg.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Миграции до поднятия роутера — иначе хендлеры могут обратиться к ещё не созданным таблицам.
	if err := postgres.RunMigrations(db, logg); err != nil {
		logg.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	app := router.NewFiberApp(cfg, logg, db)

	if err := app.Listen(cfg.ServerAddress()); err != nil {
		logg.Error("server stopped with error", "error", err)
		os.Exit(1)
	}
}
