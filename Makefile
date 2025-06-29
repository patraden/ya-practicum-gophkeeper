GO := $(shell which go)
PODMAN := $(shell which podman)
AVROGEN := $(shell which avrogen)
EASYJSON := $(shell which easyjson)
BUF := $(shell which buf)

POD_NAME=gophkeeper-pod

BUILD_DATE := $(shell date -u +"%d.%m.%Y")
BUILD_COMMIT := $(shell git rev-parse --short HEAD)
BUILD_VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "N/A")
VERSION_PACKAGE := github.com/patraden/ya-practicum-gophkeeper/internal/version

TEST_COVERAGE_REPORT := coverage.txt

SERVER_ADDRESS=localhost:3200
SERVER_ADDRESS_LOCAL=localhost:3300

POSTGRES_IMAGE_VER=15.1
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=gophkeeper
DATABASE_DSN ?= postgres://postgres:postgres@localhost:5432/gophkeeper?sslmode=disable

CRYPTO_DIR_LOCAL=./deployments/.crypto
CERT_DIR_LOCAL=./deployments/.certs
CERT_DIR_CONTAINER=/etc/ssl/certs/gophkeeper/

S3_ENDPOINT=locahost:9000

KEYCLOAK_ADMIN=admin
KEYCLOAK_ADMIN_PASSWORD=admin
KEYCLOAK_IMAGE_VER=21.0.0
KEYCLOAK_DIR_LOCAL=./deployments/podman/keycloak
KEYCLOAK_REALM=gophkeeper

MINIO_IMAGE_VER=RELEASE.2025-05-24T17-08-30Z
MINIO_ROOT_USER=gophkeeper
MINIO_ROOT_PASSWORD=gophkeeper
MINIO_REGION=eu-central-1
MINIO_REGION_NAME=eu-central-1
MINIO_COMPRESSION_ENABLE=off
MINIO_SSO_CLIENT_ID=minio
MINIO_SSO_CLIENT_SECRET=FRwPgFU1i6reAgN6lH2iM4qgQZeAMjSv
MINIO_NOTIFY_REDIS_ENABLE=on
MINIO_NOTIFY_REDIS_ADDRESS=redis:6379
MINIO_NOTIFY_REDIS_KEY=minioevents
MINIO_NOTIFY_REDIS_FORMAT=namespace


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
	@mockgen -source=server/internal/grpchandler/adapters.go -destination=server/internal/mock/grpc.go -package=mock SecretServiceServer
	@mockgen -source=server/internal/crypto/keystore/keystore.go -destination=server/internal/mock/keystore.go -package=mock Keystore
	@mockgen -source=server/internal/identity/identity.go -destination=server/internal/mock/identity.go -package=mock -mock_names "Manager=MockIdentityManager" Manager
	@mockgen -source=pkg/s3/operations.go -destination=server/internal/mock/s3.go -package=mock ServerOperator

.PHONY: json
json:
	@$(EASYJSON) -all pkg/dto/shares.go
	@$(EASYJSON) -all pkg/dto/credentials.go
	@$(EASYJSON) -all pkg/domain/secret/meta.go
	@$(EASYJSON) -all client/internal/config/config.go

.PHONY: run-server-local
run-server-local:
	@echo "Running local server installation..."
	@DATABASE_DSN="$(DATABASE_DSN)" \
	REK_SHARES_PATH="${CRYPTO_DIR_LOCAL}/shares.json" \
	S3_TLS_CERT_PATH="$(CERT_DIR_LOCAL)/minio-public.crt" \
	$(GO) run ./server/cmd/main.go -d -install

	@echo "Starting local server..."
	@SERVER_TLS_KEY_PATH="$(CERT_DIR_LOCAL)/server-private.key" \
	SERVER_ADDRESS="$(SERVER_ADDRESS_LOCAL)" \
	SERVER_TLS_CERT_PATH="$(CERT_DIR_LOCAL)/server-public.crt" \
	S3_TLS_CERT_PATH="$(CERT_DIR_LOCAL)/minio-public.crt" \
	DATABASE_DSN="$(DATABASE_DSN)" \
	$(GO) run ./server/cmd/main.go -d

.PHONY: pod-stop
pod-stop:
	@echo "Stopping all containers in pod $(POD_NAME)..."
	@$(PODMAN) ps --filter pod=$(POD_NAME) --format '{{.ID}}' | xargs -r $(PODMAN) stop
	@echo "Stopping pod $(POD_NAME)..."
	@$(PODMAN) pod stop $(POD_NAME)

.PHONY: pod-start
pod-start:
	@echo "Starting pod $(POD_NAME)..."
	@$(PODMAN) pod start $(POD_NAME)
	@echo "Starting all containers in pod $(POD_NAME)..."
	@$(PODMAN) ps -a --filter pod=$(POD_NAME) --format '{{.ID}}' | xargs -r $(PODMAN) start

.PHONY: pod-rm
pod-rm: pod-stop
	@echo "Removing all containers in pod $(POD_NAME)..."
	@$(PODMAN) ps -a --filter pod=$(POD_NAME) --format '{{.ID}}' | xargs -r $(PODMAN) rm
	@echo "Removing pod $(POD_NAME)..."
	@$(PODMAN) pod rm -f $(POD_NAME)

.PHONY: pod-status
pod-status:
	@$(PODMAN) pod ps

.PHONY: podman-vm-sync-time
podman-vm-sync-time:
	@$(PODMAN) machine ssh -- sudo systemctl restart systemd-timesyncd
	@$(PODMAN) machine ssh -- sleep 1
	@$(PODMAN) machine ssh -- timedatectl status

.PHONY: pod
pod:
	@echo "Creating gophkeeper dev pod..."
	@$(PODMAN) pod create \
  --name $(POD_NAME) \
  -p 5432:5432 \
  -p 6379:6379 \
  -p 9000:9000 \
  -p 9001:9001 \
  -p 3200:3200 \
  -p 8080:8080

.PHONY: podman-build-all
podman-build-all: podman-build-certgen

.PHONY: podman-prune
podman-prune:
	@echo "Pruning unused Podman data..."
	@$(PODMAN) container prune -f
	@$(PODMAN) volume prune -f

.PHONY: podman-build-certgen
podman-build-certgen:
	@echo "Building gophkeeper certgen image..."
	@$(PODMAN) build --rm --layers=false -f deployments/podman/certgen/Dockerfile -t gophkeeper/certgen:latest .

podman-build-server:
	@echo "Building gophkeeper server image..."
	@$(PODMAN) build --rm --layers=false \
		-f deployments/podman/server/Dockerfile \
		-t gophkeeper/server:latest \
		--build-arg VERSION_PACKAGE=$(VERSION_PACKAGE) \
		--build-arg BUILD_VERSION=$(BUILD_VERSION) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		--build-arg BUILD_COMMIT=$(BUILD_COMMIT) \
		.

.PHONY: podman-server-run
podman-server-run:
	@echo "Running gophkeeper server..."
	@$(PODMAN) run -dt \
		--name server \
		--pod $(POD_NAME) \
		--restart=on-failure \
		-e SERVER_ADDRESS=$(SERVER_ADDRESS) \
		-e S3_ENDPOINT=$(S3_ENDPOINT) \
		-e DATABASE_DSN="postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:5432/$(POSTGRES_DB)?sslmode=disable" \
		-v gophkeeper_app_certs:/etc/ssl/certs \
		gophkeeper/server:latest

.PHONY: podman-certgen-run
podman-certgen-run:
	@echo "Creating .certs dir..."
	@mkdir -p $(CERT_DIR_LOCAL)
	@echo "Running certgen to generate certs..."
	@$(PODMAN) run --name certgen --userns=keep-id \
		-v gophkeeper_app_certs:/certs/gophkeeper \
		gophkeeper/certgen:latest \
		sh -c ' \
			mkdir -p /certs/gophkeeper/ca /certs/gophkeeper/minio /certs/gophkeeper/backend && \
			echo "Generating CA certificate..." && \
			cd /certs/gophkeeper/ca/ && \
			/usr/local/bin/certgen -org-name GophKeeper -ca && \
			echo "Generating GophKeeper certificate signed by CA..." && \
			cd /certs/gophkeeper/backend/ && \
			/usr/local/bin/certgen -host 127.0.0.1,localhost,gophkeeper-server -org-name GophKeeper \
				-ca-cert /certs/gophkeeper/ca/ca-public.crt -ca-key /certs/gophkeeper/ca/ca-private.key && \
			echo "Generating GophKeeper Minio certificate signed by CA..." && \
			cd /certs/gophkeeper/minio/ && \
			/usr/local/bin/certgen -host 127.0.0.1,localhost,minio -org-name GophKeeper \
				-ca-cert /certs/gophkeeper/ca/ca-public.crt -ca-key /certs/gophkeeper/ca/ca-private.key && \
			echo "======CA CERTIFICATE======" && cat /certs/gophkeeper/ca/ca-public.crt \
		'
	@echo "Copying certificates from certgen container..."
	@$(PODMAN) cp certgen:/certs/gophkeeper/ca/ca-public.crt $(CERT_DIR_LOCAL)/ca.cert
	@$(PODMAN) cp certgen:/certs/gophkeeper/backend/private.key $(CERT_DIR_LOCAL)/server-private.key
	@$(PODMAN) cp certgen:/certs/gophkeeper/backend/public.crt $(CERT_DIR_LOCAL)/server-public.crt
	@$(PODMAN) cp certgen:/certs/gophkeeper/minio/public.crt $(CERT_DIR_LOCAL)/minio-public.crt
	@echo "Cleaning up certgen container..."
	@$(PODMAN) rm certgen

# ------------------------------------------------------------------------------
# Postgres container
# ------------------------------------------------------------------------------

.PHONY: podman-postgres-run
podman-postgres-run:
	@echo "Running gophkeeper PostgreSQL container..."
	@$(PODMAN) run -dt \
		--name postgres \
		--pod $(POD_NAME) \
		-e POSTGRES_USER=$(POSTGRES_USER) \
		-e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
		-e POSTGRES_DB=$(POSTGRES_DB) \
		-v postgres_data:/var/lib/postgresql/data \
		-v $(PWD)/deployments/podman/postgres/init-keycloak.sql:/docker-entrypoint-initdb.d/init-keycloak.sql:ro \
		--health-cmd="pg_isready -U $(POSTGRES_USER) -d $(POSTGRES_DB) -h 127.0.0.1" \
		--health-interval=1s \
		--health-timeout=3s \
		--health-retries=5 \
		--health-start-period=15s \
		postgres:$(POSTGRES_IMAGE_VER)

.PHONY: podman-postgres-down
podman-postgres-down:
	@echo "Stopping gophkeeper postgres container..."
	@$(PODMAN) stop postgres || true
	@echo "Removing gophkeeper postgres container..."
	@$(PODMAN) rm postgres || true
	@echo "Removing gophkeeper postgres volume..."
	@$(PODMAN) volume rm postgres_data || true


# ------------------------------------------------------------------------------
# Keycloak container
#
# https://www.keycloak.org/
# https://www.keycloak.org/server/containers
# https://www.keycloak.org/getting-started/getting-started-docker
# https://www.keycloak.org/server/configuration
# https://www.keycloak.org/server/all-config
# ------------------------------------------------------------------------------

.PHONY: podman-keycloak-run
podman-keycloak-run:
	@echo "Running gophkeeper Keycloak container with PostgreSQL..."
	@$(PODMAN) run -dt \
		--name keycloak \
		--pod $(POD_NAME) \
		-e KEYCLOAK_ADMIN=$(KEYCLOAK_ADMIN) \
		-e KEYCLOAK_ADMIN_PASSWORD=$(KEYCLOAK_ADMIN_PASSWORD) \
		-e KC_DB=postgres \
		-e KC_DB_URL=jdbc:postgresql://localhost:5432/$(POSTGRES_DB) \
		-e KC_DB_USERNAME=$(POSTGRES_USER) \
		-e KC_DB_SCHEMA=keycloak \
		-e KC_DB_PASSWORD=$(POSTGRES_PASSWORD) \
		-e KC_TRUSTSTORE_PATHS=$(CERT_DIR_CONTAINER)/ca/ca-public.crt \
		-v keycloak_data:/opt/keycloak/data \
		-v gophkeeper_app_certs:$(CERT_DIR_CONTAINER) \
		-v $(shell pwd)/deployments/podman/keycloak/import:/opt/keycloak/data/import:ro \
		quay.io/keycloak/keycloak:$(KEYCLOAK_IMAGE_VER) start-dev --import-realm

.PHONY: podman-keycloak-down
podman-keycloak-down:
	@echo "Stopping gophkeeper Keycloak container..."
	@$(PODMAN) stop keycloak || true
	@echo "Removing gophkeeper keycloak container..."
	@$(PODMAN) rm keycloak || true
	@echo "Removing gophkeeper keycloak volume..."
	@$(PODMAN) volume rm keycloak_data || true

.PHONY: podman-keycloak-start
podman-keycloak-start:
	@echo "Starting existing gophkeeper Keycloak container..."
	@$(PODMAN) start keycloak

.PHONY: podman-keycloak-stop
podman-keycloak-stop:
	@echo "Stopping gophkeeper Keycloak container..."
	@$(PODMAN) stop keycloak

.PHONY: podman-keycloak-export
podman-keycloak-export:
	@echo "Exporting Keycloak 'gophkeeper' realm..."
	@$(PODMAN) run --name keycloak-export \
		-v keycloak_data:/opt/keycloak/data \
		quay.io/keycloak/keycloak:$(KEYCLOAK_IMAGE_VER) export \
			--dir /opt/keycloak/data/export \
			--realm gophkeeper \
			--users realm_file
	@echo "Copying exported file to host..."
	@mkdir -p $(KEYCLOAK_DIR_LOCAL)/import
	@$(PODMAN) cp keycloak-export:/opt/keycloak/data/export/gophkeeper-realm.json $(KEYCLOAK_DIR_LOCAL)/import/gophkeeper-realm-new.json
	@echo "Cleaning up export container..."
	@$(PODMAN) rm keycloak-export

# ------------------------------------------------------------------------------
# Minio container
#
# https://min.io/
# https://min.io/docs/minio/linux/reference/minio-server/settings/iam/openid.html
# https://github.com/minio/minio/blob/master/docs/sts/keycloak.md
# https://min.io/docs/minio/container/operations/external-iam/configure-keycloak-identity-management.html
# https://min.io/docs/minio/linux/developers/security-token-service/AssumeRoleWithWebIdentity.html#minio-sts-assumerolewithwebidentity
# ------------------------------------------------------------------------------

.PHONY: podman-minio-run
podman-minio-run:
	@echo "Running gophkeeper Minio container..."
	@$(PODMAN) run -dt \
	--name minio \
	--pod $(POD_NAME) \
	-e "MINIO_ROOT_USER=$(MINIO_ROOT_USER)" \
	-e "MINIO_ROOT_PASSWORD=$(MINIO_ROOT_PASSWORD)" \
	-e "MINIO_REGION=$(MINIO_REGION)" \
	-e "MINIO_REGION_NAME=$(MINIO_REGION_NAME)" \
	-e "MINIO_COMPRESSION_ENABLE=$(MINIO_COMPRESSION_ENABLE)" \
	-e "MINIO_IDENTITY_OPENID_REDIRECT_URI_DYNAMIC_PRIMARY=on" \
	-e "MINIO_IDENTITY_OPENID_CONFIG_URL_PRIMARY=http://localhost:8080/realms/$(KEYCLOAK_REALM)/.well-known/openid-configuration" \
	-e "MINIO_IDENTITY_OPENID_CLIENT_ID_PRIMARY=$(MINIO_SSO_CLIENT_ID)" \
	-e "MINIO_IDENTITY_OPENID_CLIENT_SECRET_PRIMARY=$(MINIO_SSO_CLIENT_SECRET)" \
	-e "MINIO_IDENTITY_OPENID_DISPLAY_NAME_PRIMARY=sso_keycloak" \
	-e "MINIO_IDENTITY_OPENID_SCOPES_PRIMARY=email,profile,openid,minio-authorization,minio-admin-api-access" \
	-e "MINIO_IDENTITY_OPENID_VENDOR_PRIMARY=keycloak" \
	-e "MINIO_IDENTITY_OPENID_KEYCLOAK_ADMIN_URL_PRIMARY=http://localhost:8080/admin" \
	-e "MINIO_IDENTITY_OPENID_KEYCLOAK_REALM_PRIMARY=$(KEYCLOAK_REALM)" \
	-v gophkeeper_app_certs:$(CERT_DIR_CONTAINER) \
	-v minio_data:/data \
	quay.io/minio/minio:$(MINIO_IMAGE_VER) server /data \
		--console-address ":9001" \
		--certs-dir="$(CERT_DIR_CONTAINER)/minio/"

.PHONY: podman-minio-down
podman-minio-down:
	@echo "Stopping gophkeeper MinIO container..."
	@$(PODMAN) stop minio || true
	@echo "Removing gophkeeper MinIO container..."
	@$(PODMAN) rm minio || true
	@echo "Removing gophkeeper MinIO volume..."
	@$(PODMAN) volume rm minio_data || true

.PHONY: podman-minio-stop
podman-minio-stop:
	@echo "Stopping gophkeeper Minio container..."
	@$(PODMAN) stop minio

.PHONY: podman-minio-start
podman-minio-start:
	@echo "Starting existing gophkeeper MinIO container..."
	@$(PODMAN) start minio

.PHONY: podman-minio-rm
podman-minio-rm:
	@echo "Removing gophkeeper Minio container..."
	@$(PODMAN) rm minio


.PHONY: podman-all-down
podman-all-down: podman-minio-down podman-keycloak-down podman-postgres-down
	@echo "✅ All GophKeeper containers and volumes have been stopped and removed."
