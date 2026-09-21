package main

import (
	"context"
	"log"

	"github.com/thriftin/api/internal/config"
	"github.com/thriftin/api/internal/seed"
	"github.com/thriftin/api/pkg/database"
)

func main() {
	cfg := config.Load()
	pool, err := database.NewPool(context.Background(), cfg.DBURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	if err := seed.Run(context.Background(), pool); err != nil {
		log.Fatal(err)
	}
}
