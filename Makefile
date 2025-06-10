DOCKER := $(shell which docker)
DOCKER_COMPOSE_PATH := ./deployments/docker-compose.yaml
BUILD_DATE := $(shell date -u +"%d.%m.%Y")
BUILD_COMMIT := $(shell git rev-parse --short HEAD)
BUILD_VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "N/A")
VERSION_PACKAGE := github.com/patraden/ya-practicum-gophkeeper/internal/version
DATABASE_DSN ?= postgres://postgres:postgres@localhost:5432/gophkeeper?sslmode=disable
TEST_COVERAGE_REPORT := coverage.txt

.PHONY: lint
lint:
	@echo "🔍 Running goimports..."
	@goimports -e -w -local "github.com/patraden/ya-practicum-gophkeeper" .
	@echo "🧼 Organizing imports with gci..."
	@gci write ./client ./server ./pkg ./certgen
	@echo "🧹 Formatting with gofumpt..." gofumpt -w ./client ./server ./pkg ./certgen
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
	@rm -f $(TEST_COVERAGE_REPORT)
	@rm -f *.key
	@rm -f *.crt
	@go mod tidy

.PHONY: docker-pg
docker-pg:
	@echo "Starting PostgreSQL container..."
	docker-compose -f $(DOCKER_COMPOSE_PATH) up -d postgres

.PHONY: docker-certgen
docker-certgen:
	@echo "Running certificate generator container..."
	docker-compose -f $(DOCKER_COMPOSE_PATH) run --rm certgen

.PHONY: docker-build
docker-build: docker-down docker-clean-volumes
	@echo "Building Docker containers..."
	BUILD_DATE=$(BUILD_DATE) \
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
	@echo "Starting all containers..."
	docker-compose -f $(DOCKER_COMPOSE_PATH) up -d

.PHONY: docker-stop
docker-stop:
	@echo "Stopping all containers..."
	docker-compose -f $(DOCKER_COMPOSE_PATH) stop

.PHONY: docker-down
docker-down:
	@echo "Bringing down all containers..."
	docker-compose -f $(DOCKER_COMPOSE_PATH) down

.PHONY: docker-down-all
docker-down-all:
	@echo "Bringing down all containers and cleaning all volumes/images..."
	docker-compose -f $(DOCKER_COMPOSE_PATH) down --volumes --remove-orphans --rmi all

.PHONY: docker-clean-volumes
docker-clean-volumes:
	@echo "Removing all Docker volumes except 'gophkeeper_app_certs'..."
	@docker volume ls -q \
		| grep -v gophkeeper_app_certs \
		| xargs -r docker volume rm

.PHONY: avro
avro:
	@avrogen -pkg card -o ./pkg/domain/card/avro_card.go -tags json:snake ./avro/card.avsc
	@avrogen -pkg creds -o ./pkg/domain/creds/avro_creds.go -tags json:snake ./avro/creds.avsc

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

.PHONY: mocks
mocks:
	@mockgen -source=server/internal/grpchandler/adapters.go -destination=server/internal/mock/grpc.go -package=mock UserServiceServer
	@mockgen -source=server/internal/grpchandler/adapters.go -destination=server/internal/mock/grpc.go -package=mock AdminServiceServer
	@mockgen -source=server/internal/infra/s3/client.go -destination=server/internal/mock/s3.go -package=mock Client

.PHONY: json
json:
	@easyjson -all pkg/dto/shares.go
	@easyjson -all pkg/dto/credentials.go