package internal

import (
	"bufio"
	"io"
	"log/slog"
	"math"
)

type Client struct {
	*bufio.ReadWriter
}

func (c *Client) ReadString() (string, error) {
	len, err := c.ReadByte()
	if err != nil {
		return "", err
	}

	buf := make([]byte, len)
	if _, err := io.ReadFull(c, buf); err != nil {
		return "", err
	}

	return string(buf), nil
}

func (c *Client) WriteStringWithLen(msg string, logger *slog.Logger) error {
	length := len(msg)
	if length > math.MaxUint8 {
		msg = msg[:math.MaxUint8]
		length = math.MaxUint8
	}

	if err := c.WriteByte(byte(length)); err != nil {
		logger.Warn("Failed to write message length to client", "error", err)
	}

	if _, err := c.WriteString(msg); err != nil {
		logger.Warn("Failed to write message to client", "error", err)
	}

	return nil
}
