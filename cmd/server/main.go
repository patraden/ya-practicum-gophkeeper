package main

import (
	"github.com/patraden/ya-practicum-gophkeeper/internal/app"
	"github.com/patraden/ya-practicum-gophkeeper/internal/config"
)

func main() {
	config := config.LoadConfig()
	app := app.App(config)

	app.Run()
}
