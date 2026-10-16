package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/GhandyP/opportunity-follow-up/internal/agent"
	"github.com/GhandyP/opportunity-follow-up/internal/config"
	"github.com/GhandyP/opportunity-follow-up/internal/history"
	"github.com/GhandyP/opportunity-follow-up/internal/web"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	store, err := history.Open(ctx, cfg.HistoryDBPath)
	if err != nil {
		log.Fatalf("open local opportunity history: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			log.Printf("close local opportunity history: %v", err)
		}
	}()

	var generator agent.Generator
	if cfg.GoogleAPIKey != "" {
		generator, err = agent.NewADKGenerator(ctx, cfg.GoogleAPIKey, cfg.GeminiModel)
		if err != nil {
			log.Fatal("initialize Gemini generator failed")
		}
	}

	handler, err := web.NewApplicationHandler(agent.NewService(generator), store)
	if err != nil {
		log.Fatalf("initialize web application: %v", err)
	}

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("listening on http://%s", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
