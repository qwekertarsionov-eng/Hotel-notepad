package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"guest-accounting/internal/config"
	"guest-accounting/internal/repository/postgres"
	transport "guest-accounting/internal/transport/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := postgres.Connect(cfg.DATABASE_URL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer db.Close()

	if err := postgres.Up(db); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	handler, err := transport.NewHandler(db)
	if err != nil {
		log.Fatalf("handler: %v", err)
	}

	srv := &http.Server{
		Addr:    cfg.HTTP_ADDR,
		Handler: handler.Routes(),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
}
