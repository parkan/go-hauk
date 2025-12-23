package main

import (
	"log"
	"net/http"

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

	log.Printf("starting hauk on %s", cfg.ListenAddr)
	if err := http.ListenAndServe(cfg.ListenAddr, srv); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
