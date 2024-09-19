package literedis

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"
)

type Client struct {
	conn   net.Conn
	reader *bufio.Reader
}

func NewClient(address string) (*Client, error) {
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	return &Client{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Do(cmd string, args ...interface{}) (interface{}, error) {
	err := c.writeCommand(cmd, args...)
	if err != nil {
		return nil, err
	}
	return c.readReply()
}

func (c *Client) writeCommand(cmd string, args ...interface{}) error {
	_, err := fmt.Fprintf(c.conn, "*%d\r\n$%d\r\n%s\r\n", len(args)+1, len(cmd), cmd)
	if err != nil {
		return fmt.Errorf("failed to write command: %w", err)
	}
	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			_, err = fmt.Fprintf(c.conn, "$%d\r\n%s\r\n", len(v), v)
		case int:
			_, err = fmt.Fprintf(c.conn, "$%d\r\n%d\r\n", len(strconv.Itoa(v)), v)
		case int64:
			_, err = fmt.Fprintf(c.conn, "$%d\r\n%d\r\n", len(strconv.FormatInt(v, 10)), v)
		default:
			return fmt.Errorf("unsupported argument type: %T", arg)
		}
		if err != nil {
			return fmt.Errorf("failed to write argument: %w", err)
		}
	}
	return nil
}

func (c *Client) readReply() (interface{}, error) {
	line, err := c.reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read reply: %w", err)
	}
	if len(line) < 2 || line[len(line)-2] != '\r' {
		return nil, errors.New("invalid reply format")
	}
	line = line[:len(line)-2]

	switch line[0] {
	case '+':
		return string(line[1:]), nil
	case '-':
		return nil, errors.New(string(line[1:]))
	case ':':
		return parseInt(line[1:])
	case '$':
		return c.readBulkString(line)
	case '*':
		return c.readArray(line)
	default:
		return nil, fmt.Errorf("unknown reply type: %c", line[0])
	}
}

func (c *Client) readBulkString(line []byte) (interface{}, error) {
	length, err := parseInt(line[1:])
	if err != nil {
		return nil, err
	}
	if length == -1 {
		return nil, nil
	}
	data := make([]byte, length+2) // +2 for \r\n
	_, err = c.reader.Read(data)
	if err != nil {
		return nil, fmt.Errorf("failed to read bulk string: %w", err)
	}
	return string(data[:length]), nil
}

func (c *Client) readArray(line []byte) (interface{}, error) {
	count, err := parseInt(line[1:])
	if err != nil {
		return nil, err
	}
	if count == -1 {
		return nil, nil
	}
	array := make([]interface{}, count)
	for i := int64(0); i < count; i++ {
		array[i], err = c.readReply()
		if err != nil {
			return nil, err
		}
	}
	return array, nil
}

func (c *Client) Ping() error {
	_, err := c.Do("PING")
	return err
}

func parseInt(b []byte) (int64, error) {
	return strconv.ParseInt(string(b), 10, 64)
}
