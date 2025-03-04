# Настройки
APP_NAME := spikeys
SRC_DIR := ./cmd/spikeys
BUILD_DIR := build
LINTER := golangci-lint  
LINT_FLAGS := run 

# Список целей
.PHONY: all lint test build clean run cover

# Цель по умолчанию
all: lint test cover build

# Линтинг кода
lint:
	@echo "===> Запуск линтеров..."
	@$(LINTER) $(LINT_FLAGS) 

# Тестирование
test:
	@echo "===> Запуск тестов..."
	@go test -v ./...

# Сборка проекта
build:
	@echo "===> Сборка проекта $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(APP_NAME) $(SRC_DIR)

# Очистка временных файлов
clean:
	@echo "===> Очистка..."
	@rm -rf $(BUILD_DIR)

# Запуск приложения
run: build
	@echo "===> Запуск приложения..."
	@./$(BUILD_DIR)/$(APP_NAME)
cover:
	@echo "===> Покрытие тестов..."
	@go tool cover -html=./tmp/coverage.out -o ./tmp/coverage.html	
