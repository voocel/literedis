package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/jessevdk/go-flags"
	"literedis/config"
	"literedis/internal/app"
	"literedis/pkg/log"
)

func main() {
	config.LoadConfig()

	currentFlags, err := initFlags()
	if err != nil {
		panic(err)
	}

	log.Init("redis", currentFlags.LogLevel)

	a := app.NewApp()
	a.Start()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	for {
		sig := <-ch
		switch sig {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			a.Stop()
			log.Sync()
			return
		case syscall.SIGHUP:
			config.LoadConfig()
		default:
			return
		}
	}
}

type Flags struct {
	Host     string `short:"h" long:"host" description:"Specific server host" default:""`
	Port     string `short:"p" long:"port" description:"Specific server port" default:""`
	Setup    bool   `short:"S" long:"setup" description:"Run setup"`
	Stream   bool   `short:"s" long:"stream" description:"Stream"`
	Model    string `short:"m" long:"model" description:"Choose model"`
	Message  string `hidden:"true" description:"Message to send to chat"`
	Output   string `short:"o" long:"output" description:"Output to file" default:""`
	LogLevel string `short:"l" long:"log-level" description:"Log level" default:"debug"`
}

func initFlags() (ret *Flags, err error) {
	var message string

	ret = &Flags{}
	parser := flags.NewParser(ret, flags.Default)
	var args []string
	if args, err = parser.Parse(); err != nil {
		return
	}

	info, _ := os.Stdin.Stat()
	hasStdin := (info.Mode() & os.ModeCharDevice) == 0

	// takes input from stdin if it exists, otherwise takes input from args (the last argument)
	if hasStdin {
		if message, err = readStdin(); err != nil {
			err = errors.New("error: could not read from stdin")
			return
		}
	} else if len(args) > 0 {
		message = args[len(args)-1]
	} else {
		message = ""
	}
	ret.Message = message

	return
}

func readStdin() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	var input string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", fmt.Errorf("error reading from stdin: %w", err)
		}
		input += line
	}
	return input, nil
}
