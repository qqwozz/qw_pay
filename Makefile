.PHONY: build run db stop clean demo logs help

APP_NAME := server
MAIN := ./cmd/server
PORT := 8080

help:
	@echo "QW Pay — команды:"
	@echo ""
	@echo "  make db        — запустить PostgreSQL"
	@echo "  make stop      — остановить контейнеры"
	@echo "  make build     — собрать бинарник"
	@echo "  make run       — собрать и запустить сервер"
	@echo "  make demo      — открыть демо-страницу в браузере"
	@echo "  make logs      — показать логи сервера"
	@echo "  make clean     — удалить бинарник"

db:
	docker-compose up -d db
	@echo "Waiting for PostgreSQL..."
	@sleep 2
	@echo "PostgreSQL is ready on localhost:5432"

stop:
	docker-compose down

build:
	CGO_ENABLED=0 go build -o $(APP_NAME) $(MAIN)

run: build
	./$(APP_NAME)

demo:
	@echo "Opening demo page at http://localhost:$(PORT)/demo"
	@xdg-open "http://localhost:$(PORT)/demo" 2>/dev/null || open "http://localhost:$(PORT)/demo" 2>/dev/null || echo "Open http://localhost:$(PORT)/demo in your browser"

logs:
	@tail -f logs/server.log 2>/dev/null || echo "No logs yet. Start server with 'make run'"

clean:
	rm -f $(APP_NAME)
