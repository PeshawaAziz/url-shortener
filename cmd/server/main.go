package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/PeshawaAziz/url-shortener/internal/domain/url"
	"github.com/PeshawaAziz/url-shortener/internal/infra/analytics"
	"github.com/PeshawaAziz/url-shortener/internal/infra/cache"
	"github.com/PeshawaAziz/url-shortener/internal/infra/db"
	"github.com/PeshawaAziz/url-shortener/internal/infra/security"
	"github.com/PeshawaAziz/url-shortener/internal/infra/slug"
	"github.com/PeshawaAziz/url-shortener/internal/infra/validation"
	"github.com/PeshawaAziz/url-shortener/internal/infra/workers"
	httpInterfaces "github.com/PeshawaAziz/url-shortener/internal/interfaces/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 1. Initialize Infrastructure (Adapters)
	dsn := "host=localhost user=postgres password=postgres dbname=shortener port=5432 sslmode=disable"
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	urlRepo := db.NewPostgresURLRepository(gormDB)
	slugGen := slug.NewCryptoGenerator(7)
	bloomFilter := cache.NewLocalBloomFilter(14400000, 10) // ~1.8MB memory footprint
	reservedChecker := validation.NewConfigReservedChecker([]string{"billing", "sales", "support"})
	idempotencyStore := cache.NewRedisIdempotencyStore(rdb)

	urlCache := cache.NewRedisURLCache(rdb)
	clickReporter := analytics.NewAsyncClickReporter(100000) // 100k buffer
	rateLimiter := cache.NewRedisRateLimiter(rdb)
	hasher := security.NewArgon2Hasher()

	// 2. Initialize Domain Services & Engines
	abHasher := &url.SHA1ABTestHasher{}
	rulesEngine := url.NewRulesEngine(abHasher)

	shortenerService := url.NewShortenerService(urlRepo, slugGen, bloomFilter, reservedChecker, idempotencyStore)
	redirectService := url.NewRedirectService(urlRepo, urlCache, clickReporter, rulesEngine, rateLimiter)
	lifecycleService := url.NewLifecycleService(urlRepo)
	passwordService := url.NewPasswordService(urlRepo, hasher)

	// 3. Initialize Background Workers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	expirationWorker := workers.NewExpirationWorker(lifecycleService)
	go expirationWorker.Start(ctx) // Run expiration sweep in background goroutine

	// 4. Initialize HTTP Layer
	router := gin.Default()

	// Wire up the handlers
	linkHandler := httpInterfaces.NewLinkHandler(shortenerService)
	redirectHandler := httpInterfaces.NewRedirectHandler(redirectService, passwordService)

	// Register Routes
	router.POST("/v1/links", linkHandler.HandleCreate)
	router.GET("/:slug", redirectHandler.HandleRedirect)
	router.POST("/v1/links/:slug/unlock", redirectHandler.HandleUnlock)

	// 5. Start Server with Graceful Shutdown
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		log.Println("Server starting on port :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully. Goodbye!")
}
