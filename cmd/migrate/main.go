package main

import (
	"os"

	"github.com/lyimoexiao/akari/internal/config"
	"github.com/lyimoexiao/akari/internal/database"
	"github.com/lyimoexiao/akari/internal/logger"
)

func main() {
	cfg, err := config.Load(os.Getenv("CONFIG_PATH"))
	if err != nil {
		panic(err)
	}

	logger.Init(&cfg.Logger)
	l := logger.L

	db, err := database.New(&cfg.Database)
	if err != nil {
		l.Fatalw("failed to setup database", "error", err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	l.Info("migration complete")
}
