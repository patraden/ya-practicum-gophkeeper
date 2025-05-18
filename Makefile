DOCKER := $(shell which docker)
DOCKER_COMPOSE_PATH := ./deployments/docker-compose.yaml
BUILD_DATE := $(shell date -u +"%d.%m.%Y")
BUILD_COMMIT := $(shell git rev-parse --short HEAD)
BUILD_VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "N/A")
TEST_COVERAGE_REPORT := coverage.txt

.PHONY: lint
lint:
	@goimports -e -w -local "github.com/patraden/ya-practicum-gophkeeper" .
	@gci write ./cmd ./internal ./pkg
	@gofumpt -w ./cmd ./internal ./pkg
	@golangci-lint run ./...

.PHONY: test
test:
	@go test -v -coverprofile=$(TEST_COVERAGE_REPORT) ./...
	@go tool cover -html=$(TEST_COVERAGE_REPORT)

.PHONY: clean
clean:
	@rm -f ./$(TEST_COVERAGE_REPORT)

.PHONY: docker\:certgen
docker\:certgen:
	@docker-compose -f $(DOCKER_COMPOSE_PATH) run --rm certgen

.PHONY: docker\:up 
docker\:up:
	@docker-compose -f $(DOCKER_COMPOSE_PATH) up -d

.PHONY: docker\:down
docker\:down:
	@docker-compose -f $(DOCKER_COMPOSE_PATH) down -v

.PHONY: gen\:avro
gen\:avro:
	@avrogen -pkg card -o ./internal/domain/card/avro_card.go -tags json:snake ./avro/card.avsc
	@avrogen -pkg creds -o ./internal/domain/creds/avro_creds.go -tags json:snake ./avro/creds.avsc

.PHONY: gen\:sql
gen\:sql:
	@sqlc generate

