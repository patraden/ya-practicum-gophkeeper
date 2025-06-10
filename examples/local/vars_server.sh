export SERVER_TLS_KEY_PATH=./.certs/server-private.key
export SERVER_TLS_CERT_PATH=./.certs/server-public.crt
export DATABASE_DSN="postgres://postgres:postgres@localhost:5432/gophkeeper?sslmode=disable"