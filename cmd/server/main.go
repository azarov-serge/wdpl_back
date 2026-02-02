package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"wdpl_back/internal/shared/config"
	"wdpl_back/internal/shared/http/router"
	"wdpl_back/internal/shared/logger"
	"wdpl_back/internal/shared/postgres"
)

func main() {
	// Загружаем .env только в dev; в Docker можно использовать переменные окружения контейнера.
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logg := logger.New(cfg)

	db, err := postgres.Connect(cfg, logg)
	if err != nil {
		logg.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

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
