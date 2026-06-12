package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/qw_pay/internal/config"
	"github.com/qw_pay/internal/database"
	"github.com/qw_pay/internal/handlers"
	"github.com/qw_pay/internal/middleware"
	"github.com/qw_pay/internal/services"
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

	// Инициализация сервисов
	authSvc := services.NewAuthService(database.Pool)
	accountSvc := services.NewAccountService(database.Pool)
	txSvc := services.NewTransactionService(database.Pool, accountSvc)

	// Инициализация хендлеров
	authH := handlers.NewAuthHandler(authSvc)
	accountH := handlers.NewAccountHandler(accountSvc)
	txH := handlers.NewTransactionHandler(txSvc, accountSvc)

	// Настройка HTTP-сервера
	r := gin.Default()

	// Демо-страница
	r.StaticFile("/", "./web/index.html")
	r.StaticFile("/demo", "./web/index.html")
	r.Static("/static", "./web")

	v1 := r.Group("/api/v1")
	{
		// Публичные эндпоинты (без авторизации)
		v1.POST("/register", authH.Register)
		v1.POST("/verify", authH.VerifyOTP)
		v1.POST("/login", authH.Login)

		// Защищённые эндпоинты (требуют JWT)
		auth := v1.Group("")
		auth.Use(middleware.AuthRequired())
		{
			// Управление счетами
			auth.POST("/accounts", accountH.Create)
			auth.GET("/accounts", accountH.List)
			auth.POST("/accounts/:id/block", accountH.Block)

			// Переводы
			auth.POST("/transactions", txH.Create)
			auth.GET("/transactions", txH.List)
		}
	}

	// Запуск сервера
	addr := fmt.Sprintf(":%s", config.C.ServerPort)
	log.Printf("Server listening on http://localhost%s", addr)
	log.Printf("Demo page: http://localhost%s/demo", addr)
	log.Printf("Logs: logs/server.log")
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
