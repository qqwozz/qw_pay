package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/qw_pay/internal/antifraud"
	"github.com/qw_pay/internal/config"
	"github.com/qw_pay/internal/database"
	"github.com/qw_pay/internal/middleware"

	"github.com/qw_pay/internal/auth"
	"github.com/qw_pay/internal/account"
	"github.com/qw_pay/internal/transaction"
)

func setupLogger() {
	os.MkdirAll("logs", 0755)
	logFile, err := os.OpenFile("logs/server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
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

	// Anti-fraud client (Redis)
	fraudClient := antifraud.NewClient("127.0.0.1:6379")
	if err := fraudClient.Ping(context.Background()); err != nil {
		log.Printf("[ANTIFRAUD] Redis not available: %v — anti-fraud disabled", err)
		fraudClient.Close()
		fraudClient = nil
	} else {
		log.Println("[ANTIFRAUD] Connected to Redis — anti-fraud active")
		defer fraudClient.Close()
	}

	// Repositories
	userRepo := auth.NewUserRepository(database.Pool)
	accountRepo := account.NewRepository(database.Pool)
	txRepo := transaction.NewRepository(database.Pool)

	// Services
	authSvc := auth.NewService(userRepo)
	accountSvc := account.NewService(accountRepo)
	txSvc := transaction.NewService(database.Pool, txRepo, accountSvc)

	// Handlers
	authH := auth.NewHandler(authSvc)
	accountH := account.NewHandler(accountSvc)
	txH := transaction.NewHandler(txSvc, accountSvc, fraudClient)

	// Anti-fraud handler
	var afH *antifraud.Handler
	if fraudClient != nil {
		afH = antifraud.NewHandler(fraudClient)
	}

	// Router
	r := gin.Default()

	r.StaticFile("/", "./web/index.html")
	r.StaticFile("/demo", "./web/index.html")
	r.Static("/static", "./web")

	v1 := r.Group("/api/v1")
	{
		v1.POST("/register", authH.Register)
		v1.POST("/verify", authH.VerifyOTP)
		v1.POST("/login", authH.Login)

		auth := v1.Group("")
		auth.Use(middleware.AuthRequired())
		{
			auth.POST("/accounts", accountH.Create)
			auth.GET("/accounts", accountH.List)
			auth.POST("/accounts/:id/block", accountH.Block)

			auth.POST("/transactions", txH.Create)
			auth.GET("/transactions", txH.List)

			if afH != nil {
				auth.POST("/antifraud/block-user", afH.BlockUser)
				auth.POST("/antifraud/block-account", afH.BlockAccount)
			}
		}
	}

	addr := fmt.Sprintf(":%s", config.C.ServerPort)
	log.Printf("Server listening on http://localhost%s", addr)
	log.Printf("Demo page: http://localhost%s/demo", addr)
	log.Printf("Logs: logs/server.log")
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
