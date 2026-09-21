package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/thriftin/api/internal/config"
	"github.com/thriftin/api/internal/handler"
	"github.com/thriftin/api/internal/repository"
	"github.com/thriftin/api/internal/router"
	"github.com/thriftin/api/internal/seed"
	"github.com/thriftin/api/internal/service"
	"github.com/thriftin/api/pkg/database"
	redisPkg "github.com/thriftin/api/pkg/redis"
	_ "github.com/thriftin/api/docs"
)

// @title Thriftin API
// @version 1.0
// @description Thriftin second-hand marketplace API
// @BasePath /api/v1
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
func main() {
	cfg := config.Load()
	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.DBURL)
	if err != nil { log.Fatal(err) }
	defer pool.Close()
	if err := runMigrations(ctx, pool); err != nil { log.Printf("migrate: %v", err) }
	if err := seed.Run(ctx, pool); err != nil { log.Printf("seed: %v", err) }
	rdb, _ := redisPkg.New(cfg.RedisURL)

	repos := repository.New(pool, rdb)
	svcs := service.New(repos, cfg)
	h := &router.Handlers{
		Auth: handler.NewAuth(svcs.Auth), Product: handler.NewProduct(svcs.Product),
		Category: handler.NewCategory(svcs.Category), Cart: handler.NewCart(svcs.Cart),
		Order: handler.NewOrder(svcs.Order), Wishlist: handler.NewWishlist(svcs.Wishlist),
		Review: handler.NewReview(svcs.Review), Follow: handler.NewFollow(svcs.Follow),
		Chat: handler.NewChat(svcs.Chat), Notification: handler.NewNotification(svcs.Notification),
		Admin: handler.NewAdmin(),
	}
	gin.SetMode(gin.ReleaseMode)
	if cfg.Env == "development" { gin.SetMode(gin.DebugMode) }
	r := gin.New()
	r.Use(gin.Recovery())
	router.Setup(r, h, cfg.JWTSecret)
	r.GET("/health", func(c *gin.Context){ c.JSON(200, gin.H{"status":"ok"}) })
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	port := cfg.Port
	if port == "" { port = "8080" }
	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil { log.Fatal(err) }
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	matches, _ := filepath.Glob("migrations/*.up.sql")
	for _, m := range matches {
		b, err := os.ReadFile(m)
		if err != nil { continue }
		if _, err := pool.Exec(ctx, string(b)); err != nil { return err }
	}
	return nil
}
