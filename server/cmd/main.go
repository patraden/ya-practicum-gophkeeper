package main

import (
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/bootstrap"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/config"
)

func main() {
	config := config.LoadConfig()
	app := bootstrap.Server(config)

	app.Run()
}
