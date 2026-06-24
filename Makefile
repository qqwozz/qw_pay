.PHONY: build run db stop clean demo logs antifraud antifraud-build help lint test test-race coverage vet tidy migrate

APP_NAME := server
MAIN := ./cmd/server
PORT := 8080

help:
	@echo ""
	@echo "  QW Pay — Команды разработчика"
	@echo "  ============================"
	@echo ""
	@echo "  Инфраструктура:"
	@echo "    make db              — запустить PostgreSQL + Redis"
	@echo "    make stop            — остановить все контейнеры"
	@echo ""
	@echo "  Сервер:"
	@echo "    make build           — собрать Go сервер"
	@echo "    make run             — собрать и запустить сервер"
	@echo "    make lint            — запустить линтер"
	@echo "    make test            — запустить тесты"
	@echo "    make test-race       — тесты с race detector"
	@echo "    make coverage        — покрытие тестов"
	@echo "    make vet             — static analysis"
	@echo "    make tidy            — go mod tidy"
	@echo ""
	@echo "  Anti-Fraud:"
	@echo "    make antifraud-build — собрать C++ движок"
	@echo "    make antifraud       — запустить C++ + Python движки"
	@echo ""
	@echo "  Прочее:"
	@echo "    make demo            — открыть демо-страницу"
	@echo "    make logs            — показать логи сервера"
	@echo "    make clean           — удалить бинарники"
	@echo ""

db:
	docker compose up -d db redis
	@echo "Waiting for services..."
	@sleep 3
	@echo "PostgreSQL: localhost:5432"
	@echo "Redis:      localhost:6379"

stop:
	docker compose down

build:
	CGO_ENABLED=0 go build -o $(APP_NAME) $(MAIN)

run: build
	./$(APP_NAME)

lint:
	@command -v golangci-lint >/dev/null 2>&1 || { echo "Install: https://golangci-lint.run/usage/install/"; exit 1; }
	golangci-lint run ./...

test:
	go test -v ./...

test-race:
	go test -race ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	@echo "Coverage report: coverage.out"

vet:
	go vet ./...

tidy:
	go mod tidy

antifraud-build:
	$(MAKE) -C antifraud/cpp

antifraud: antifraud-build
	@echo "Starting anti-fraud engines (C++ + Python)..."
	python3 antifraud/orchestrator.py

demo:
	@echo "Opening demo page at http://localhost:$(PORT)/demo"
	@xdg-open "http://localhost:$(PORT)/demo" 2>/dev/null || open "http://localhost:$(PORT)/demo" 2>/dev/null || echo "Open http://localhost:$(PORT)/demo in your browser"

logs:
	@tail -f logs/server.log 2>/dev/null || echo "No logs yet. Start server with 'make run'"

clean:
	rm -f $(APP_NAME)
	rm -f coverage.out
	$(MAKE) -C antifraud/cpp clean
