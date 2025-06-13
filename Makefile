DOCKER := $(shell which docker)
AVROGEN := $(shell which avrogen)
EASYJSON := $(shell which easyjson)
BUF := $(shell which buf)
DOCKER_COMPOSE_PATH := ./deployments/docker-compose.yaml
BUILD_DATE := $(shell date -u +"%d.%m.%Y")
BUILD_COMMIT := $(shell git rev-parse --short HEAD)
BUILD_VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "N/A")
VERSION_PACKAGE := github.com/patraden/ya-practicum-gophkeeper/internal/version
DATABASE_DSN ?= postgres://postgres:postgres@localhost:5432/gophkeeper?sslmode=disable
TEST_COVERAGE_REPORT := coverage.txt
CERT_DIR_LOCAL=./deployments/.certs
CRYPTO_DIR_LOCAL=./deployments/.crypto
CERT_DIR_DOCKER=/etc/ssl/certs/gophkeeper/

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

.PHONY: copy-certs
copy-certs:
	@echo "Copying certificates from MinIO container..."
	@$(DOCKER) cp "minio:$(CERT_DIR_DOCKER)/ca/ca-public.crt" "$(CERT_DIR_LOCAL)/ca.cert"
	@$(DOCKER) cp "minio:$(CERT_DIR_DOCKER)/backend/private.key" "$(CERT_DIR_LOCAL)/server-private.key"
	@$(DOCKER) cp "minio:$(CERT_DIR_DOCKER)/backend/public.crt" "$(CERT_DIR_LOCAL)/server-public.crt"
	@$(DOCKER) cp "minio:$(CERT_DIR_DOCKER)/minio/public.crt" "$(CERT_DIR_LOCAL)/minio-public.crt"

.PHONY: copy-shares
copy-shares:
	@echo "Copying REK shares locally..."
	@$(DOCKER) cp "gophkeeper-server:/app/shares.json" "${CRYPTO_DIR_LOCAL}/shares.json"

.PHONY: docker-infra
docker-infra: docker-down
	@echo "Starting PostgreSQL container..."
	docker-compose -f $(DOCKER_COMPOSE_PATH) up -d postgres
	@echo "Starting Redis container..."
	docker-compose -f $(DOCKER_COMPOSE_PATH) up -d redis
	@echo "Starting Minio container..."
	docker-compose -f $(DOCKER_COMPOSE_PATH) up -d minio

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

.PHONY: docker-up
docker-up:
	@echo "Starting all containers..."
	@docker-compose -f $(DOCKER_COMPOSE_PATH) up -d
	@echo "Sleeping to let server to initialize..."
	@sleep 3
	$(MAKE) copy-shares
	$(MAKE) copy-certs

.PHONY: docker-stop
docker-stop:
	@echo "Stopping all containers..."
	docker-compose -f $(DOCKER_COMPOSE_PATH) stop

.PHONY: docker-down
docker-down: 
	@echo "Bringing down all containers..."
	docker-compose -f $(DOCKER_COMPOSE_PATH) down
	$(MAKE) docker-clean-volumes

.PHONY: docker-down-all
docker-down-all:
	@echo "Bringing down all containers and cleaning all volumes/images..."
	docker-compose -f $(DOCKER_COMPOSE_PATH) down --volumes --remove-orphans --rmi all

.PHONY: docker-clean-volumes
docker-clean-volumes:
	@echo "Removing all Docker volumes except 'gophkeeper_app_certs'..."
	@docker volume ls -q | grep -v gophkeeper_app_certs | xargs -r docker volume rm || true

.PHONY: avro
avro:
	@$(AVROGEN) -pkg card -o ./pkg/domain/card/avro_card.go -tags json:snake ./avro/card.avsc
	@$(AVROGEN) -pkg creds -o ./pkg/domain/creds/avro_creds.go -tags json:snake ./avro/creds.avsc

.PHONY: sql
sql:
	@sqlc generate

.PHONY: proto
proto:
	@echo "🔍 Running buf lint..."
	@$(BUF) lint
	@echo "📦 Updating buf dependencies..."
	@$(BUF) dep update
	@echo "⚙️ Generating protobuf and validation code..."
	@$(BUF) generate
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
	@$(EASYJSON) -all pkg/dto/shares.go
	@$(EASYJSON) -all pkg/dto/credentials.go

.PHONY: run-server
run-server: copy-certs
	@echo "Running local server installation..."
	@DATABASE_DSN="$(DATABASE_DSN)" \
	REK_SHARES_PATH="${CRYPTO_DIR_LOCAL}/shares.json" \
	S3_TLS_CERT_PATH="$(CERT_DIR_LOCAL)/minio-public.crt" \
	go run ./server/cmd/main.go -d -install

	@echo "Starting local server..."
	@SERVER_TLS_KEY_PATH="$(CERT_DIR_LOCAL)/server-private.key" \
	SERVER_TLS_CERT_PATH="$(CERT_DIR_LOCAL)/server-public.crt" \
	S3_TLS_CERT_PATH="$(CERT_DIR_LOCAL)/minio-public.crt" \
	DATABASE_DSN="$(DATABASE_DSN)" \
	go run ./server/cmd/main.go -d
