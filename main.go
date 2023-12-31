package main

import (
	"log/slog"

	"github.com/rs/zerolog"
	slogzerolog "github.com/samber/slog-zerolog/v2"

	"github.com/mocha8686/oc_server/internal"
)

func main() {
	logger := zerolog.New(zerolog.NewConsoleWriter())
	slog.SetDefault(slog.New(slogzerolog.Option{
		Level:  slog.LevelDebug,
		Logger: &logger,
	}.NewZerologHandler()))

	if err := internal.Start(8000); err != nil {
		slog.Error("Error while starting server", "error", err)
	}
}
