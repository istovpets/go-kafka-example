# Конфигурация
COMPOSE_FILE = docker-compose.yml
KAFKA_NODES  = kafka-1 kafka-2 kafka-3

# Команды Docker Compose (для краткости)
DOCKER_COMPOSE = docker-compose -f $(COMPOSE_FILE)

.PHONY: kafka-up kafka-down kafka-logs kafka-status clean help

## --- KAFKA OPS ---

# Запуск только инфраструктуры Kafka
kafka-up:
	@echo "Starting Kafka cluster (KRaft mode)..."
	$(DOCKER_COMPOSE) up -d $(KAFKA_NODES)

# Остановка только инфраструктуры Kafka
kafka-down:
	$(DOCKER_COMPOSE) stop $(KAFKA_NODES)
	$(DOCKER_COMPOSE) rm -f $(KAFKA_NODES)

# Просмотр логов только брокеров
kafka-logs:
	$(DOCKER_COMPOSE) logs -f $(KAFKA_NODES)

# Статус брокеров
kafka-status:
	$(DOCKER_COMPOSE) ps $(KAFKA_NODES)

## --- SYSTEM OPS ---

.PHONY: dep-up
dep-up: kafka-up
	@echo "Waiting for Kafka to be ready..."
	@sleep 7
	@echo "Creating topic 'foo'..."
	@docker exec kafka-1 /opt/kafka/bin/kafka-topics.sh --create --topic foo --bootstrap-server kafka-1:19092 --partitions 3 --replication-factor 1 --if-not-exists 2>/dev/null || true

.PHONY: dep-down
dep-down: kafka-down

# Полная очистка: остановка всех контейнеров и удаление volumes с данными
clean:
	@echo "Cleaning up all containers and volumes..."
	$(DOCKER_COMPOSE) down -v

# Помощь: выводит список всех команд
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'