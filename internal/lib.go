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
	logger := slog.With("id", id)

	// TODO: heartbeat + recovery
	const (
		subscribe = iota
		unsubscribe
		publish
	)
	logger.Debug("Start read command")

	cmd, err := conn.ReadByte()
	if err != nil {
		return err
	}
	logger.Debug("Command", "val", cmd)

	topicLen, err := conn.ReadByte()
	if err != nil {
		return err
	}
	logger.Debug("Topic len", "val", topicLen)

	topicBuf := make([]byte, topicLen)
	if _, err := io.ReadFull(conn, topicBuf); err != nil {
		return err
	}
	topic := string(topicBuf)
	logger.Debug("Topic", "val", topic)

	logger = logger.With("topic", topic)

	switch cmd {
	case subscribe:
		topicChan, err := pubsub.Subscribe(id, topic)
		if err != nil {
			return err
		}

		go func(id, topic string, pubsub *PubSub) {
			for msg := range topicChan {
				logger.Info("Client received message", "msg", msg)

				if _, err := conn.WriteString(msg + "\n"); err != nil {
					logger.Warn("Failed to write incoming message to client", "msg", msg, "error", err)
				}

				if err := conn.Flush(); err != nil {
					logger.Warn("Failed to flush buffer", "error", err)
				}
			}
			logger.Debug("Topic channel closed")
		}(id, topic, pubsub)
		logger.Info("Client subscribed to topic")

	case unsubscribe:
		if err := pubsub.Unsubscribe(id, topic); err != nil {
			return err
		}
		logger.Info("Client unsubscribed from topic")

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
		logger.Info("Client published message to topic", "msg", msg)
	default:
		return fmt.Errorf("Unknown command %v", cmd)
	}

	return nil
}
