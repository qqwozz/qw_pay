package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/qw_pay/internal/account"
	"github.com/qw_pay/internal/antifraud"
	"github.com/qw_pay/internal/auth"
	"github.com/qw_pay/internal/config"
	"github.com/qw_pay/internal/database"
	"github.com/qw_pay/internal/middleware"
	"github.com/qw_pay/internal/ratelimit"
	"github.com/qw_pay/internal/transaction"
)

func setupLogger() {
	_ = os.MkdirAll("logs", 0o750)
	logFile, err := os.OpenFile("logs/server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	setupLogger()
	log.Printf("=== QW Pay server starting at %s ===", time.Now().Format(time.RFC3339))

	config.Load()
	log.Printf("Config loaded — port=%s, max_transfer=%.0f, daily_limit=%.0f",
		config.C.ServerPort, config.C.MaxTransferAmount, config.C.DailyLimit)

	database.Connect()
	defer database.Close()
	log.Println("Database connected")

	var fraudClient *antifraud.Client
	fraudClient = antifraud.NewClient(config.C.RedisAddr)
	if err := fraudClient.Ping(context.Background()); err != nil {
		log.Printf("[ANTIFRAUD] Redis not available: %v — anti-fraud disabled", err)
		_ = fraudClient.Close()
		fraudClient = nil
	} else {
		log.Println("[ANTIFRAUD] Connected to Redis — anti-fraud active")
		defer func() { _ = fraudClient.Close() }()
	}

	userRepo := auth.NewUserRepository(database.Pool)
	accountRepo := account.NewRepository(database.Pool)
	txRepo := transaction.NewRepository(database.Pool)

	authSvc := auth.NewService(userRepo)
	accountSvc := account.NewService(accountRepo)
	txSvc := transaction.NewService(database.Pool, txRepo, accountSvc)

	authH := auth.NewHandler(authSvc)
	accountH := account.NewHandler(accountSvc)
	txH := transaction.NewHandler(txSvc, accountSvc, fraudClient)

	var afH *antifraud.Handler
	if fraudClient != nil {
		afH = antifraud.NewHandler(fraudClient)
	}

	r := gin.Default()

	r.StaticFile("/", "./web/index.html")
	r.StaticFile("/demo", "./web/index.html")
	r.Static("/static", "./web")

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "timestamp": time.Now().Format(time.RFC3339)})
	})

	v1 := r.Group("/api/v1")

	authRateLimit := ratelimit.New(10, 20)
	v1.POST("/register", authRateLimit.Middleware(), authH.Register)
	v1.POST("/verify", authRateLimit.Middleware(), authH.VerifyOTP)
	v1.POST("/login", authRateLimit.Middleware(), authH.Login)

	authGroup := v1.Group("")
	authGroup.Use(middleware.AuthRequired())
	authGroup.POST("/accounts", accountH.Create)
	authGroup.GET("/accounts", accountH.List)
	authGroup.POST("/accounts/:id/block", accountH.Block)
	authGroup.POST("/transactions", txH.Create)
	authGroup.GET("/transactions", txH.List)

	if afH != nil {
		authGroup.POST("/antifraud/block-user", afH.BlockUser)
		authGroup.POST("/antifraud/block-account", afH.BlockAccount)
	}

	addr := fmt.Sprintf(":%s", config.C.ServerPort)

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server listening on http://localhost%s", addr)
		log.Printf("Demo page: http://localhost%s/demo", addr)
		log.Printf("Health: http://localhost%s/health", addr)
		log.Printf("Logs: logs/server.log")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		cancel()
		return
	}
	log.Println("Server exited gracefully")
}
