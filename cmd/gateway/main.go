package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/flenhu/Gateway/internal/config"
	"github.com/flenhu/Gateway/internal/router"
)

func main() {
	cfg := config.Load()

	server := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           router.New(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErr := make(chan error, 1)

	go func() {
		log.Printf("LLM Gateway listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErr:
		log.Fatalf("gateway server failed: %v", err)
	case <-ctx.Done():
		log.Println("shutting down gateway server")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("gateway server shutdown failed: %v", err)
	}
}
