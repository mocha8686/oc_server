package internal

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
)

func Start(port uint16) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		slog.Error("Failed to create listener", "error", err)
		os.Exit(1)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			slog.Error("Failed to accept connection", "error", err)
			continue
		}

		slog.Info("Client connected")
		go handleConnection(conn)
	}
}

func handleConnection(c net.Conn) {
	defer c.Close()

	buf := make([]byte, 1024)
	for {
		n, err := c.Read(buf)
		if err != nil {
			if err == io.EOF {
				slog.Info("Client disconnected")
				return
			}

			slog.Warn("Failed to read client data", "error", err)
			return
		}

		msg := buf[:n]
		slog.Info("Message received from client", "msg", string(msg))

		_, err = c.Write(msg)
		if err != nil {
			slog.Warn("Failed to write to client", "error", err)
			return
		}
	}
}
