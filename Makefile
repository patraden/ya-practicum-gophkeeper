DOCKER := $(shell which docker)
DOCKER_COMPOSE_PATH := ./deployments/docker-compose.yaml
BUILD_DATE := $(shell date -u +"%d.%m.%Y")
BUILD_COMMIT := $(shell git rev-parse --short HEAD)
BUILD_VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "N/A")
VERSION_PACKAGE := github.com/patraden/ya-practicum-gophkeeper/internal/version
TEST_COVERAGE_REPORT := coverage.txt

.PHONY: lint
lint:
	@echo "🔍 Running goimports..."
	@goimports -e -w -local "github.com/patraden/ya-practicum-gophkeeper" .
	@echo "🧼 Organizing imports with gci..."
	@gci write ./cmd ./internal ./pkg
	@echo "🧹 Formatting with gofumpt..."
	@gofumpt -w ./cmd ./internal ./pkg
	@echo "🔎 Running golangci-lint..."
	@golangci-lint run ./...
	@echo "✅ Linting complete."

.PHONY: test
test:
	@echo "🧪 Running tests with coverage..."
	@go test -v -coverprofile=$(TEST_COVERAGE_REPORT) ./...
	@echo "📊 Generating coverage report..."
	@go tool cover -html=$(TEST_COVERAGE_REPORT)
	@echo "✅ Tests complete."

.PHONY: clean
clean:
	@rm -f ./$(TEST_COVERAGE_REPORT)

.PHONY: docker-certgen
docker-certgen:
	@docker-compose -f $(DOCKER_COMPOSE_PATH) run --rm certgen

.PHONY: docker-build
docker-build: docker-down
	@BUILD_DATE=$(BUILD_DATE) \
	BUILD_COMMIT=$(BUILD_COMMIT) \
	BUILD_VERSION=$(BUILD_VERSION) \
	VERSION_PACKAGE=$(VERSION_PACKAGE) \
	docker-compose -f $(DOCKER_COMPOSE_PATH) build \
		--build-arg VERSION_PACKAGE=$(VERSION_PACKAGE) \
		--build-arg BUILD_VERSION=$(BUILD_VERSION) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		--build-arg BUILD_COMMIT=$(BUILD_COMMIT) \
		--no-cache
	$(MAKE) docker-up

.PHONY: docker-up 
docker-up:
	@docker-compose -f $(DOCKER_COMPOSE_PATH) up -d

.PHONY: docker-stop
docker-stop:
	@docker-compose -f $(DOCKER_COMPOSE_PATH) stop

.PHONY: docker-down
docker-down:
	@docker-compose -f $(DOCKER_COMPOSE_PATH) down

.PHONY: docker-down-all
docker-down-all:
	@docker-compose -f $(DOCKER_COMPOSE_PATH) down --volumes --remove-orphans --rmi all

.PHONY: avro
avro:
	@avrogen -pkg card -o ./internal/domain/card/avro_card.go -tags json:snake ./avro/card.avsc
	@avrogen -pkg creds -o ./internal/domain/creds/avro_creds.go -tags json:snake ./avro/creds.avsc

.PHONY: sql
sql:
	@sqlc generate

.PHONY: proto
proto:
	@echo "🔍 Running buf lint..."
	@buf lint
	@echo "📦 Updating buf dependencies..."
	@buf dep update
	@echo "⚙️ Generating protobuf and validation code..."
	@buf generate
	@echo "📥 Ensuring protovalidate runtime is installed..."
	@go get buf.build/go/protovalidate
	@echo "🧹 Tidying go.mod..."
	@go mod tidy
	@echo "✅ Proto generation complete."

