package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
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
	"github.com/qw_pay/internal/logger"
	"github.com/qw_pay/internal/middleware"
	"github.com/qw_pay/internal/ratelimit"
	"github.com/qw_pay/internal/transaction"
)

func main() {
	_ = os.MkdirAll("logs", 0o750)
	logFile, err := os.OpenFile("logs/server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		slog.Error("failed to open log file", "error", err)
		os.Exit(1)
	}
	defer func() { _ = logFile.Close() }()

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	logger.Setup(multiWriter)

	logger.Info("server starting", "time", time.Now().Format(time.RFC3339))

	config.Load()
	logger.Info("config loaded",
		"port", config.C.ServerPort,
		"max_transfer", config.C.MaxTransferAmount,
		"daily_limit", config.C.DailyLimit,
	)

	database.Connect()
	defer database.Close()
	logger.Info("database connected")

	var fraudClient *antifraud.Client
	fraudClient = antifraud.NewClient(config.C.RedisAddr)
	if err := fraudClient.Ping(context.Background()); err != nil {
		logger.Warn("redis not available, anti-fraud disabled", "error", err)
		_ = fraudClient.Close()
		fraudClient = nil
	} else {
		logger.Info("redis connected, anti-fraud active")
		defer func() { _ = fraudClient.Close() }()
	}

	userRepo := auth.NewUserRepository(database.Pool)
	accountRepo := account.NewRepository(database.Pool)
	txRepo := transaction.NewRepository(database.Pool)
	refreshTokenRepo := auth.NewRefreshTokenRepository(database.Pool)

	authSvc := auth.NewService(userRepo)
	accountSvc := account.NewService(accountRepo)
	txSvc := transaction.NewService(database.Pool, txRepo, accountSvc)
	refreshTokenSvc := auth.NewRefreshTokenService(refreshTokenRepo)

	authH := auth.NewHandler(authSvc, refreshTokenSvc)
	accountH := account.NewHandler(accountSvc)
	txH := transaction.NewHandler(txSvc, accountSvc, fraudClient)
	refreshH := auth.NewRefreshHandler(refreshTokenSvc)

	var afH *antifraud.Handler
	if fraudClient != nil {
		afH = antifraud.NewHandler(fraudClient)
	}

	r := gin.Default()
	r.Use(middleware.RequestID())

	r.StaticFile("/", "./web/index.html")
	r.StaticFile("/demo", "./web/index.html")
	r.Static("/static", "./web")

	r.GET("/health", func(c *gin.Context) {
		status := gin.H{"status": "ok", "timestamp": time.Now().Format(time.RFC3339)}

		if err := database.Pool.Ping(c.Request.Context()); err != nil {
			status["database"] = "unreachable"
			status["status"] = "degraded"
			c.JSON(http.StatusServiceUnavailable, status)
			return
		}
		status["database"] = "ok"

		if fraudClient != nil {
			if err := fraudClient.Ping(c.Request.Context()); err != nil {
				status["redis"] = "unreachable"
				status["status"] = "degraded"
			} else {
				status["redis"] = "ok"
			}
		} else {
			status["redis"] = "disabled"
		}

		c.JSON(http.StatusOK, status)
	})

	v1 := r.Group("/api/v1")

	authRateLimit := ratelimit.New(10, 20)
	v1.POST("/register", authRateLimit.Middleware(), authH.Register)
	v1.POST("/verify", authRateLimit.Middleware(), authH.VerifyOTP)
	v1.POST("/login", authRateLimit.Middleware(), authH.Login)
	v1.POST("/refresh", authRateLimit.Middleware(), refreshH.Refresh)

	authGroup := v1.Group("")
	authGroup.Use(middleware.AuthRequired())
	authGroup.POST("/logout", refreshH.Logout)
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
		logger.Info("server listening",
			"addr", addr,
			"demo", fmt.Sprintf("http://localhost%s/demo", addr),
			"health", fmt.Sprintf("http://localhost%s/health", addr),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
		return
	}
	logger.Info("server exited gracefully")
}
