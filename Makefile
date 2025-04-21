.PHONY: build run test cover clean docker-build docker-run docker-stop help

# Переменные
APP_NAME = avito_pvz_service
MAIN_PATH = ./cmd/main/main.go
COVERAGE_FILE = coverage.out
COVERAGE_HTML = coverage.html

# Цвета для вывода в консоль
GREEN = \033[0;32m
NC = \033[0m # No Color

help:
	@echo "${GREEN}Доступные команды:${NC}"
	@echo "  make build        - Собрать приложение"
	@echo "  make run          - Запустить приложение"
	@echo "  make test         - Запустить тесты"
	@echo "  make cover        - Запустить тесты с покрытием"
	@echo "  make cover-html   - Сгенерировать HTML-отчет о покрытии"
	@echo "  make clean        - Очистить артефакты сборки"
	@echo "  make docker-build - Собрать Docker-образ"
	@echo "  make docker-run   - Запустить в Docker Compose"
	@echo "  make docker-stop  - Остановить Docker Compose"

# Сборка приложения
build:
	@echo "${GREEN}Сборка приложения...${NC}"
	go build -o $(APP_NAME) $(MAIN_PATH)

# Запуск приложения
run:
	@echo "${GREEN}Запуск приложения...${NC}"
	go run $(MAIN_PATH)

# Запуск тестов
test:
	@echo "${GREEN}Запуск тестов...${NC}"
	go test ./internal/...

# Запуск тестов с покрытием
cover:
	@echo "${GREEN}Запуск тестов с покрытием...${NC}"
	go test -coverprofile=$(COVERAGE_FILE) ./internal/...
	go tool cover -func=$(COVERAGE_FILE)

# Генерация HTML-отчета о покрытии
cover-html: cover
	@echo "${GREEN}Генерация HTML-отчета о покрытии...${NC}"
	go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Отчет сохранен в $(COVERAGE_HTML)"

# Очистка артефактов сборки
clean:
	@echo "${GREEN}Очистка артефактов сборки...${NC}"
	rm -f $(APP_NAME)
	rm -f $(COVERAGE_FILE)
	rm -f $(COVERAGE_HTML)

# Docker
docker-build:
	@echo "${GREEN}Сборка Docker-образа...${NC}"
	docker-compose build

docker-run:
	@echo "${GREEN}Запуск в Docker Compose...${NC}"
	docker-compose up -d

docker-stop:
	@echo "${GREEN}Остановка Docker Compose...${NC}"
	docker-compose down 