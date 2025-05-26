FROM golang:1.24.3 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ARG BUILD_VERSION
ARG BUILD_DATE
ARG BUILD_COMMIT
ARG VERSION_PACKAGE

RUN go build -ldflags="\
  -s -w \
  -X ${VERSION_PACKAGE}.buildVersion=${BUILD_VERSION} \
  -X ${VERSION_PACKAGE}.buildDate=${BUILD_DATE} \
  -X ${VERSION_PACKAGE}.buildCommit=${BUILD_COMMIT}" \
  -o /app/gophkeeper ./cmd/server

FROM golang:1.24.3

WORKDIR /app

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/gophkeeper /usr/local/bin/gophkeeper

CMD /usr/local/bin/gophkeeper