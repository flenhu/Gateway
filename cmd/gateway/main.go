package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/flenhu/Gateway/internal/config"
	"github.com/flenhu/Gateway/internal/provider"
	"github.com/flenhu/Gateway/internal/router"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	providers := configuredProviders(cfg)
	registry, err := provider.NewRegistry(providers...)
	if err != nil {
		log.Fatalf("build provider registry: %v", err)
	}

	server := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           router.New(registry),
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
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

func configuredProviders(cfg config.Config) []provider.Provider {
	providers := make([]provider.Provider, 0, 1)

	if strings.TrimSpace(cfg.Providers.GroqAPIKey) != "" {
		timeout := cfg.Server.WriteTimeout
		if timeout <= 0 {
			timeout = 30 * time.Second
		}

		providers = append(providers, provider.NewGroqProvider(cfg.Providers.GroqAPIKey, timeout))
	}

	if len(providers) == 0 {
		log.Printf("no upstream providers configured; set %s to enable routing", strings.Join([]string{
			"GROQ_API_KEY",
			"GEMINI_API_KEY",
			"OPENAI_API_KEY",
			"ANTHROPIC_API_KEY",
		}, ", "))
	}

	return providers
}
