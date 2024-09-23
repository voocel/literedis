package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

type commandFactory func(*string, *int) *cobra.Command

var commandFactories = map[string]commandFactory{
	"info":    newInfoCommand,
	"monitor": newMonitorCommand,
	"slowlog": newSlowlogCommand,
	"config":  newConfigCommand,
	"keys":    newKeysCommand,
	"flushdb": newFlushDBCommand,
	"dbsize":  newDBSizeCommand,
	"ping":    newPingCommand,
	"time":    newTimeCommand,
}

func CreateCommands(host *string, port *int) []*cobra.Command {
	commands := make([]*cobra.Command, 0, len(commandFactories))
	for _, factory := range commandFactories {
		commands = append(commands, factory(host, port))
	}
	return commands
}

func newInfoCommand(host *string, port *int) *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Get information and statistics about the server",
		Run:   createCommandRunner("INFO", host, port),
	}
}

func newMonitorCommand(host *string, port *int) *cobra.Command {
	return &cobra.Command{
		Use:   "monitor",
		Short: "Listen for all requests received by the server in real-time",
		Run: func(cmd *cobra.Command, args []string) {
			client, err := pool.Get()
			if err != nil {
				fmt.Printf("Failed to get client: %v\n", err)
				return
			}
			defer pool.Put(client)

			fmt.Println("Entering monitoring mode. Press Ctrl-C to exit.")
			err = client.Monitor(func(command string) {
				fmt.Println(command)
			})
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		},
	}
}

func newSlowlogCommand(host *string, port *int) *cobra.Command {
	return &cobra.Command{
		Use:   "slowlog [get|len|reset]",
		Short: "Manage the slow logs",
		Run: func(cmd *cobra.Command, args []string) {
			runCommand(append([]string{"SLOWLOG"}, args...), host, port)
		},
	}
}

func newConfigCommand(host *string, port *int) *cobra.Command {
	return &cobra.Command{
		Use:   "config [get|set] parameter [value]",
		Short: "Get or set configuration parameters",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 2 {
				fmt.Println("Usage: config [get|set] parameter [value]")
				return
			}
			runCommand(append([]string{"CONFIG"}, args...), host, port)
		},
	}
}

func newKeysCommand(host *string, port *int) *cobra.Command {
	return &cobra.Command{
		Use:   "keys pattern",
		Short: "Find all keys matching the given pattern",
		Run:   createCommandRunner("KEYS", host, port),
	}
}

func newFlushDBCommand(host *string, port *int) *cobra.Command {
	return &cobra.Command{
		Use:   "flushdb",
		Short: "Remove all keys from the current database",
		Run:   createCommandRunner("FLUSHDB", host, port),
	}
}

func newDBSizeCommand(host *string, port *int) *cobra.Command {
	return &cobra.Command{
		Use:   "dbsize",
		Short: "Return the number of keys in the current database",
		Run:   createCommandRunner("DBSIZE", host, port),
	}
}

func newPingCommand(host *string, port *int) *cobra.Command {
	return &cobra.Command{
		Use:   "ping",
		Short: "Ping the server",
		Run:   createCommandRunner("PING", host, port),
	}
}

func newTimeCommand(host *string, port *int) *cobra.Command {
	return &cobra.Command{
		Use:   "time",
		Short: "Return the current server time",
		Run:   createCommandRunner("TIME", host, port),
	}
}

func createCommandRunner(command string, host *string, port *int) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		runCommand(append([]string{command}, args...), host, port)
	}
}

func runCommand(args []string, host *string, port *int) {
	if len(args) == 0 {
		fmt.Println("Error: No command provided")
		return
	}

	client, err := pool.Get()
	if err != nil {
		fmt.Printf("Failed to get client: %v\n", err)
		return
	}
	defer pool.Put(client)

	// 将 []string 转换为 []interface{}
	interfaceArgs := make([]interface{}, len(args)-1)
	for i, v := range args[1:] {
		interfaceArgs[i] = v
	}

	result, err := client.Do(args[0], interfaceArgs...)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("%v\n", result)
	}
}
