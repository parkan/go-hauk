package main

import (
	"log"
	"net/http"
	"time"

	"github.com/parkan/go-hauk/api"
	"github.com/parkan/go-hauk/config"
	"github.com/parkan/go-hauk/store"
)

func main() {
	cfg := config.Load()

	redis, err := store.NewRedis(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisPrefix)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}

	srv := api.NewServer(cfg, redis)

	server := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           srv,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	log.Printf("starting hauk on %s", cfg.ListenAddr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
