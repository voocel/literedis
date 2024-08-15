package main

import (
	"literedis/config"
	"literedis/internal/app"
	"literedis/pkg/log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	config.LoadConfig()
	log.Init("redis", "debug")

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
