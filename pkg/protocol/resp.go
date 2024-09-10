package protocol

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// RESPProtocol implements the Protocol interface using RESP
type RESPProtocol struct {
}

func NewRESPProtocol() *RESPProtocol {
	return &RESPProtocol{}
}

const (
	SimpleStringPrefix = '+'
	ErrorPrefix        = '-'
	IntegerPrefix      = ':'
	BulkStringPrefix   = '$'
	ArrayPrefix        = '*'
	CRLF               = "\r\n"
)

// Pack packs a Message into a RESP packet
func (p *RESPProtocol) Pack(msg *Message) ([]byte, error) {
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	switch msg.Type {
	case "SimpleString":
		_, err := writer.WriteString(fmt.Sprintf("%c%s%s", SimpleStringPrefix, msg.Content, CRLF))
		if err != nil {
			return nil, err
		}
	case "Error":
		_, err := writer.WriteString(fmt.Sprintf("%c%s%s", ErrorPrefix, msg.Content, CRLF))
		if err != nil {
			return nil, err
		}
	case "Integer":
		_, err := writer.WriteString(fmt.Sprintf("%c%d%s", IntegerPrefix, msg.Content, CRLF))
		if err != nil {
			return nil, err
		}
	case "BulkString":
		content := msg.Content.([]byte)
		if content == nil {
			_, err := writer.WriteString(fmt.Sprintf("%c-1%s", BulkStringPrefix, CRLF))
			if err != nil {
				return nil, err
			}
		} else {
			_, err := writer.WriteString(fmt.Sprintf("%c%d%s%s%s", BulkStringPrefix, len(content), CRLF, content, CRLF))
			if err != nil {
				return nil, err
			}
		}
	case "Array":
		content := msg.Content.([]*Message)
		_, err := writer.WriteString(fmt.Sprintf("%c%d%s", ArrayPrefix, len(content), CRLF))
		if err != nil {
			return nil, err
		}
		for _, elem := range content {
			elemBytes, err := p.Pack(elem)
			if err != nil {
				return nil, err
			}
			_, err = writer.Write(elemBytes)
			if err != nil {
				return nil, err
			}
		}
	default:
		return nil, fmt.Errorf("unknown message type: %s", msg.Type)
	}

	writer.Flush()
	return buf.Bytes(), nil
}

// Unpack unpacks a RESP packet into a Message
func (p *RESPProtocol) Unpack(reader io.Reader) (*Message, error) {
	bufReader := bufio.NewReader(reader)
	prefix, err := bufReader.ReadByte()
	if err != nil {
		return nil, err
	}

	switch prefix {
	case SimpleStringPrefix:
		line, err := bufReader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		return &Message{Type: "SimpleString", Content: strings.TrimSuffix(line, CRLF)}, nil
	case ErrorPrefix:
		line, err := bufReader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		return &Message{Type: "Error", Content: strings.TrimSuffix(line, CRLF)}, nil
	case IntegerPrefix:
		line, err := bufReader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		value, err := strconv.ParseInt(strings.TrimSuffix(line, CRLF), 10, 64)
		if err != nil {
			return nil, err
		}
		return &Message{Type: "Integer", Content: value}, nil
	case BulkStringPrefix:
		line, err := bufReader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		length, err := strconv.Atoi(strings.TrimSuffix(line, CRLF))
		if err != nil {
			return nil, err
		}
		if length == -1 {
			return &Message{Type: "BulkString", Content: nil}, nil
		}
		data := make([]byte, length+2)
		_, err = bufReader.Read(data)
		if err != nil {
			return nil, err
		}
		return &Message{Type: "BulkString", Content: data[:length]}, nil
	case ArrayPrefix:
		line, err := bufReader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		length, err := strconv.Atoi(strings.TrimSuffix(line, CRLF))
		if err != nil {
			return nil, err
		}
		array := make([]*Message, length)
		for i := 0; i < length; i++ {
			element, err := p.Unpack(bufReader)
			if err != nil {
				return nil, err
			}
			array[i] = element
		}
		return &Message{Type: "Array", Content: array}, nil
	default:
		return nil, fmt.Errorf("unknown prefix: %c", prefix)
	}
}
