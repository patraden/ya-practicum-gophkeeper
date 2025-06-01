FROM golang:1.24.3 AS builder

WORKDIR /certgen

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /out/certgen ./cmd/certgen

FROM golang:1.24.3

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /out/certgen /usr/local/bin/certgen

WORKDIR /certs

ENTRYPOINT ["/usr/local/bin/certgen"]