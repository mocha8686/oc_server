package internal

import (
	"bufio"
	"io"
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
