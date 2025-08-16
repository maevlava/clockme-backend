package main

import (
	"context"
	"database/sql"
	"github.com/clockme/clockme-backend/internal/config"
	"github.com/clockme/clockme-backend/internal/db"
	"github.com/clockme/clockme-backend/internal/logger"
	"github.com/clockme/clockme-backend/internal/server"
	"github.com/rs/zerolog/log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"
)

func main() {
	logger.Init()
	cfg := config.Load()
	conn, queries := connectDB()
	defer conn.Close()

	srv := server.NewClockmeServer(cfg, queries)

	log.Info().Msgf("Starting server on %s", srv.Address)
	err := http.ListenAndServe(srv.Address, srv)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}
func connectDB() (*sql.DB, *db.Queries) {
	ctxBg := context.Background()

	dbSource := os.Getenv("DB_SOURCE")
	if dbSource == "" {
		log.Fatal().Msg("DB_SOURCE is not set")
	}

	conn, err := sql.Open("postgres", dbSource)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	ctx, cancel := context.WithTimeout(ctxBg, 10*time.Second)
	defer cancel()
	if err := conn.PingContext(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to ping database")
	}

	log.Info().Msg("Successfully connected to the database!")

	queries := db.New(conn)

	return conn, queries
}
