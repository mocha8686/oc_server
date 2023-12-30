package oc

import (
	"io"
	"log/slog"
	"net"
)

type Robot struct {
	name string
	conn net.Conn
	cmd chan string
}

func NewRobot(name string, conn net.Conn) *Robot {
	return &Robot{
		name: name,
		conn: conn,
		cmd: make(chan string),
	}
}

func (r *Robot) Cmd() chan<- string {
	return r.cmd
}

func (r *Robot) Name() string {
	return r.name
}

func (r *Robot) Run() {
	out := make(chan string)
	close := make(chan struct{})
	go r.readLoop(out, close)

	for {
		select {
		case cmd := <-r.cmd:
			slog.Info("Command received", "name", r.name, "cmd", cmd)
			_, err := r.conn.Write([]byte(cmd))
			if err != nil {
				slog.Warn("Error writing command to robot", "name", r.name, "error", err)
				continue
			}
		case out := <-out:
			slog.Info("Message from client", "name", r.name, "msg", string(out))
		case <-close:
			return
		}
	}
}

func (r *Robot) readLoop(out chan<- string, close chan<- struct{}) {
	defer func() {
		r.conn.Close()
		close <- struct{}{}
	}()
	buf := make([]byte, 1024)
	for {
		n, err := r.conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				slog.Info("Connection closed", "name", r.name)
			} else {
				slog.Error("Error while reading client data", "name", r.name, "error", err)
			}
			return
		}

		out <- string(buf[:n])
	}
}
