package literedis

import (
	"bufio"
	"errors"
	"strconv"
)

func readLine(reader *bufio.Reader) ([]byte, error) {
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	if len(line) < 2 || line[len(line)-2] != '\r' {
		return nil, errors.New("invalid line format")
	}
	return line[:len(line)-2], nil
}

func parseReply(reader *bufio.Reader) (interface{}, error) {
	line, err := readLine(reader)
	if err != nil {
		return nil, err
	}

	switch line[0] {
	case '+':
		return string(line[1:]), nil
	case '-':
		return nil, errors.New(string(line[1:]))
	case ':':
		return strconv.ParseInt(string(line[1:]), 10, 64)
	case '$':
		length, err := strconv.Atoi(string(line[1:]))
		if err != nil {
			return nil, err
		}
		if length == -1 {
			return nil, nil
		}
		data, err := readLine(reader)
		if err != nil {
			return nil, err
		}
		return string(data), nil
	case '*':
		length, err := strconv.Atoi(string(line[1:]))
		if err != nil {
			return nil, err
		}
		if length == -1 {
			return nil, nil
		}
		array := make([]interface{}, length)
		for i := 0; i < length; i++ {
			array[i], err = parseReply(reader)
			if err != nil {
				return nil, err
			}
		}
		return array, nil
	default:
		return nil, errors.New("unknown reply type")
	}
}
