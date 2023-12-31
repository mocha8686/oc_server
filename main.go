package main

import (
	"log/slog"
	"strconv"

	"github.com/rs/zerolog"
	slogzerolog "github.com/samber/slog-zerolog/v2"

	"os"

	"github.com/mocha8686/oc_server/internal"
)

func main() {
	logger := zerolog.New(zerolog.NewConsoleWriter())
	slog.SetDefault(slog.New(slogzerolog.Option{
		Level:  slog.LevelDebug,
		Logger: &logger,
	}.NewZerologHandler()))

	portStr, exists := os.LookupEnv("PORT")
	var port uint16 = 8000

	if exists {
		parsedPort, err := strconv.ParseInt(portStr, 10, 16)
		if err == nil {
			port = uint16(parsedPort)
		} else {
			slog.Warn("Failed to parse port, using default port 8000", "error", err)
		}
	}

	if err := internal.Start(port); err != nil {
		slog.Error("Error while starting server", "error", err)
	}
}
