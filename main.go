package main

import (
	"log/slog"

	"github.com/mocha8686/oc_server/internal"
)

func main() {
	if err := internal.Start(8000); err != nil {
		slog.Error("Error while starting server", "error", err)
	}
}
