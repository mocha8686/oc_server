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

	ctx := NewContext()

	slog.Info("Listening", "port", port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			slog.Error("Failed to accept connection", "error", err)
			continue
		}

		slog.Info("Connection received")

		go handleConnection(conn, ctx)
	}
}

func handleConnection(conn net.Conn, ctx Context) {
	defer conn.Close()

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	c := Client{rw}

	id, err := c.ReadString()
	if err != nil {
		slog.Error("Failed to read ID", "error", err)
		return
	}
	if err := ctx.GetRegistry().Register(id); err != nil {
		slog.Error("Failed to register ID", "err", err)
		if err == IdInUse {
			if err := c.WriteStringWithLen(err.Error(), slog.Default()); err != nil {
				slog.Error("Failed to send error message to client", "error", err)
			}
		}
		return
	}
	slog.Info("Client connected", "id", id)

	defer ctx.GetRegistry().Unregister(id)
	defer ctx.GetPubSub().UnsubscribeAll(id)

	for {
		if err := processCommand(c, id, ctx); err != nil {
			if err == io.EOF {
				slog.Info("Client disconnected", "id", id)
			} else {
				slog.Error("Error while processing command", "id", id, "error", err)
			}
			break
		}
	}
}

func processCommand(c Client, id string, ctx Context) error {
	logger := slog.With("id", id)

	// TODO: heartbeat + recovery
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
	case Subscribe:
		topicChan, err := ctx.GetPubSub().Subscribe(id, topic)
		if err != nil {
			return err
		}

		go func(id, topic string, ctx Context) {
			for msg := range topicChan {
				logger = logger.With("msg", msg)

				logger.Info("Client received message")

				if err := c.WriteStringWithLen(msg, logger); err != nil {
					logger.Warn("Failed to write message length to client", "error", err)
					continue
				}

				if err := c.Flush(); err != nil {
					logger.Warn("Failed to flush buffer", "error", err)
					continue
				}
			}
			logger.Debug("Topic channel closed")
		}(id, topic, ctx)
		logger.Info("Client subscribed to topic")

	case Unsubscribe:
		if err := ctx.GetPubSub().Unsubscribe(id, topic); err != nil {
			return err
		}
		logger.Info("Client unsubscribed from topic")

	case Publish:
		msg, err := c.ReadString()
		if err != nil {
			return err
		}

		ctx.GetPubSub().Publish(topic, msg)
		logger.Info("Client published message to topic", "msg", msg)
	default:
		return fmt.Errorf("Unknown command %v", cmd)
	}

	return nil
}
