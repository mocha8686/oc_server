package internal

import (
	"fmt"
	"io"
	"log/slog"
	"net"

	"github.com/mocha8686/oc_server/internal/oc"
)

type registeries struct {
	robots *Registry[Robot]
}

func Start(port uint16) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		return err
	}

	r := registeries{
		robots: NewRegistry[Robot](),
	}

	slog.Info("Listening", "port", port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			slog.Error("Failed to accept connection", "error", err)
			continue
		}

		slog.Info("Client connected")
		go handleConnection(conn, &r)
	}
}

func handleConnection(conn net.Conn, r *registeries) {
	defer conn.Close()

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		if err == io.EOF {
			slog.Info("Client disconnected")
			return
		}

		slog.Error("Failed to read client data", "error", err)
		return
	}

	data := buf[:n]
	slog.Info("Message received from client", "data", data)

	typeId := data[0]

	switch typeId {
		case OCRobot: // INFO: (echo '\x00\x03Bob'; cat) | nc localhost 8000
		nameLen := data[1]
		name := string(data[2 : 2+nameLen])
		slog.Info("OpenComputers Robot connected", "name", name)
		robot := oc.NewRobot(name, conn)
		r.robots.Register(name, robot)
		defer r.robots.Remove(name)

		robot.Run()

		case Tester: // HACK: nasty code !!! (only for testing) (echo '\x01'; cat | nc localhost 8000)
		slog.Info("Tester connected")
		for {
			n, err := conn.Read(buf)
			if err != nil {
				if err == io.EOF {
					slog.Info("Tester disconnected")
					return
				}

				slog.Warn("Failed to read tester command", "error", err)
				return
			}

			cmd := string(buf[:n])
			slog.Info("Command received from tester", "cmd", cmd)

			for _, r := range r.robots.Items() {
				slog.Info("Sending command to robot", "name", r.Name())
				r.Cmd() <- cmd
			}
		}
	default:
		slog.Warn("Unknown client", "typeid", typeId)
		return
	}
}
