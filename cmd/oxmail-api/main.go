package main

import (
	"log/slog"
	"os"

	"github.com/MYusufEka/oxmail/internal/api"
	"github.com/MYusufEka/oxmail/internal/database"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	port := os.Getenv("OXMAIL_API_PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("OXMAIL_DB_PATH")
	if dbPath == "" {
		dbPath = "oxmail.db"
	}

	db, err := database.Open(dbPath, logger)
	if err != nil {
		logger.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	logger.Info("database connected", "path", dbPath)

	srv := api.NewServer(db.Conn)

	if err := srv.ListenAndServe(port); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}
