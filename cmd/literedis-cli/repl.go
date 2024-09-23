package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type REPL struct {
	client *Client
}

func NewREPL(client *Client) *REPL {
	return &REPL{client: client}
}

func (r *REPL) Run() {
	fmt.Printf("Connected to LiteRedis server at %s\n", r.client.Addr())
	fmt.Println("Type 'help' for available commands or 'exit' to quit")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("literedis> ")
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		if input == "exit" {
			break
		}

		if input == "help" {
			printHelp()
			continue
		}

		r.executeCommand(input)
	}
}

func (r *REPL) executeCommand(input string) {
	args := strings.Fields(input)
	if len(args) == 0 {
		return
	}

	// 将 []string 转换为 []interface{}
	interfaceArgs := make([]interface{}, len(args)-1)
	for i, v := range args[1:] {
		interfaceArgs[i] = v
	}

	result, err := r.client.Do(args[0], interfaceArgs...)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("%v\n", result)
	}
}

func printHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  SET key value")
	fmt.Println("  GET key")
	fmt.Println("  DEL key [key ...]")
	fmt.Println("  EXISTS key [key ...]")
	fmt.Println("  INCR key")
	fmt.Println("  DECR key")
	fmt.Println("  EXPIRE key seconds")
	fmt.Println("  TTL key")
	fmt.Println("  INFO")
	fmt.Println("  MONITOR")
	fmt.Println("  SLOWLOG [GET|LEN|RESET]")
	fmt.Println("  CONFIG GET parameter")
	fmt.Println("  CONFIG SET parameter value")
	fmt.Println("  exit")
}
