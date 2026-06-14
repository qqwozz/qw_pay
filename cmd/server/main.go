package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/qw_pay/internal/account"
	"github.com/qw_pay/internal/antifraud"
	"github.com/qw_pay/internal/audit"
	"github.com/qw_pay/internal/auth"
	"github.com/qw_pay/internal/config"
	"github.com/qw_pay/internal/currency"
	"github.com/qw_pay/internal/database"
	"github.com/qw_pay/internal/middleware"
	"github.com/qw_pay/internal/transaction"
)

func setupLogger() {
	os.MkdirAll("logs", 0755)
	logFile, err := os.OpenFile("logs/server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
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

	fraudClient := initAntiFraud()

	userRepo := auth.NewUserRepository(database.Pool)
	accountRepo := account.NewRepository(database.Pool)
	txRepo := transaction.NewRepository(database.Pool)
	currencyRepo := currency.NewRepository(database.Pool)
	auditRepo := audit.NewRepository(database.Pool)

	authSvc := auth.NewService(userRepo)
	accountSvc := account.NewService(accountRepo)
	currencySvc := currency.NewService(currencyRepo)
	txSvc := transaction.NewService(database.Pool, txRepo, accountSvc, currencySvc)
	auditSvc := audit.NewService(auditRepo)

	authH := auth.NewHandler(authSvc, auditSvc)
	accountH := account.NewHandler(accountSvc, auditSvc)
	currencyH := currency.NewHandler(currencySvc)
	txH := transaction.NewHandler(txSvc, accountSvc, fraudClient, auditSvc)
	auditH := audit.NewHandler(auditSvc)

	r := gin.Default()
	registerRoutes(r, authH, accountH, currencyH, txH, auditH, fraudClient)

	addr := fmt.Sprintf(":%s", config.C.ServerPort)
	log.Printf("Server listening on http://localhost%s", addr)
	log.Printf("Demo page: http://localhost%s/demo", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}

func initAntiFraud() *antifraud.Client {
	client := antifraud.NewClient(config.C.RedisAddr)
	if err := client.Ping(context.Background()); err != nil {
		log.Printf("[ANTIFRAUD] Redis not available: %v — anti-fraud disabled", err)
		client.Close()
		return nil
	}
	log.Println("[ANTIFRAUD] Connected to Redis — anti-fraud active")
	return client
}

func registerRoutes(r *gin.Engine, authH *auth.Handler, accountH *account.Handler,
	currencyH *currency.Handler, txH *transaction.Handler, auditH *audit.Handler,
	fraudClient *antifraud.Client) {

	r.StaticFile("/", "./web/index.html")
	r.StaticFile("/demo", "./web/index.html")
	r.Static("/static", "./web")

	v1 := r.Group("/api/v1")
	{
		v1.POST("/register", authH.Register)
		v1.POST("/verify", authH.VerifyOTP)
		v1.POST("/login", authH.Login)

		v1.GET("/currencies", currencyH.ListCurrencies)
		v1.GET("/currencies/rates", currencyH.GetRates)
		v1.GET("/currencies/rate", currencyH.GetRate)
		v1.POST("/currencies/convert", currencyH.Convert)

		auth := v1.Group("")
		auth.Use(middleware.AuthRequired())
		{
			auth.POST("/accounts", accountH.Create)
			auth.GET("/accounts", accountH.List)
			auth.POST("/accounts/:id/block", accountH.Block)

			auth.POST("/transactions", txH.Create)
			auth.GET("/transactions", txH.List)

			auth.POST("/currencies/rate", currencyH.UpdateRate)

			if fraudClient != nil {
				afH := antifraud.NewHandler(fraudClient)
				auth.POST("/antifraud/block-user", afH.BlockUser)
				auth.POST("/antifraud/block-account", afH.BlockAccount)
			}

			admin := auth.Group("/admin")
			admin.Use(middleware.AdminRequired())
			{
				admin.GET("/audit-logs", auditH.ListLogs)
				admin.POST("/accounts/block", accountH.AdminBlock)
			}
		}
	}
}
