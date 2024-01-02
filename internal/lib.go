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
	c := Client{rw}

	id, err := c.ReadString()
	if err != nil {
		slog.Error("Failed to read ID", "error", err)
		return
	}
	slog.Info("Client connected", "id", id)

	defer pubsub.UnsubscribeAll(id)

	for {
		if err := processCommand(c, id, pubsub); err != nil {
			if err == io.EOF {
				slog.Info("Client disconnected", "id", id)
			} else {
				slog.Error("Error while processing command", "id", id, "error", err)
			}
			break
		}
	}
}

func processCommand(c Client, id string, pubsub *PubSub) error {
	logger := slog.With("id", id)

	// TODO: heartbeat + recovery
	const (
		subscribe = iota
		unsubscribe
		publish
	)
	cmd, err := c.ReadByte()
	if err != nil {
		return err
	}

	topic, err := c.ReadString()
	if err != nil {
		return err
	}

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

				if _, err := c.WriteString(msg + "\n"); err != nil {
					logger.Warn("Failed to write incoming message to client", "msg", msg, "error", err)
				}

				if err := c.Flush(); err != nil {
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
		msg, err := c.ReadString()
		if err != nil {
			return err
		}

		pubsub.Publish(topic, msg)
		logger.Info("Client published message to topic", "msg", msg)
	default:
		return fmt.Errorf("Unknown command %v", cmd)
	}

	return nil
}
