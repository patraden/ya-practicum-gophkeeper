package main

import (
	"github.com/patraden/ya-practicum-gophkeeper/internal/app"
	"github.com/rs/zerolog"
)

func main() {
	app := app.App(zerolog.DebugLevel)
	app.Run()
}
