package internal

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"net"
)

func Start(port uint16) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		return err
	}

	pubsub := NewPubSub()

	slog.Info("Listening", "port", port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			slog.Error("Failed to accept connection", "error", err)
			continue
		}

		slog.Info("Connection received")

		go handleConnection(conn, pubsub)
	}
}

func handleConnection(conn net.Conn, pubsub *PubSub) {
	defer conn.Close()

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	idLen, err := rw.ReadByte()
	if err != nil {
		slog.Error("Failed to read client ID length")
		return
	}

	idBuf := make([]byte, idLen)
	if _, err := io.ReadFull(rw, idBuf); err != nil {
		slog.Error("Failed to read client ID")
		return
	}

	id := string(idBuf)
	slog.Info("Client connected", "id", id)

	defer pubsub.UnsubscribeAll(id)

	for {
		if err := processCommand(rw, id, pubsub); err != nil {
			if err == io.EOF {
				slog.Info("Client disconnected", "id", id)
			} else {
				slog.Error("Error while processing command", "id", id, "error", err)
			}
			break
		}
	}
}

func processCommand(conn *bufio.ReadWriter, id string, pubsub *PubSub) error {
	// TODO: heartbeat + recovery
	const (
		subscribe = iota
		unsubscribe
		publish
	)
	slog.Debug("Start read command")

	cmd, err := conn.ReadByte()
	if err != nil {
		return err
	}
	slog.Debug("Command", "val", cmd)

	topicLen, err := conn.ReadByte()
	if err != nil {
		return err
	}
	slog.Debug("Topic len", "val", topicLen)

	topicBuf := make([]byte, topicLen)
	if _, err := io.ReadFull(conn, topicBuf); err != nil {
		return err
	}
	topic := string(topicBuf)
	slog.Debug("Topic", "val", topic)

	switch cmd {
	case subscribe:
		topicChan, err := pubsub.Subscribe(id, topic)
		if err != nil {
			return err
		}

		go func(id, topic string, pubsub *PubSub) {
			for msg := range topicChan {
				slog.Info("Client received message", "id", id, "topic", topic, "msg", msg)

				if _, err := conn.WriteString(msg + "\n"); err != nil {
					slog.Warn("Failed to write incoming message to client", "id", id, "topic", topic, "msg", msg, "error", err)
				}

				if err := conn.Flush(); err != nil {
					slog.Warn("Failed to flush buffer", "id", id, "error", err)
				}
			}
			slog.Debug("Topic channel closed", "id", id, "topic", topic)
		}(id, topic, pubsub)
		slog.Info("Client subscribed to topic", "id", id, "topic", topic)

	case unsubscribe:
		if err := pubsub.Unsubscribe(id, topic); err != nil {
			return err
		}
		slog.Info("Client unsubscribed from topic", "id", id, "topic", topic)

	case publish:
		msgLen, err := conn.ReadByte()
		if err != nil {
			return err
		}

		msgBuf := make([]byte, msgLen)
		if _, err := io.ReadFull(conn, msgBuf); err != nil {
			return err
		}
		msg := string(msgBuf)

		pubsub.Publish(topic, msg)
		slog.Info("Client published message to topic", "id", id, "topic", topic, "msg", msg)
	default:
		return fmt.Errorf("Unknown command %v", cmd)
	}

	return nil
}
