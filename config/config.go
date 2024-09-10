package config

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"log"
	"strings"
)

var Conf = new(config)

type config struct {
	Mode            string
	Name            string
	AOFFile         string
	IOBufferLength  int
	LogLevel        string `mapstructure:"log_level"`
	LogPath         string `mapstructure:"log_path"`
	LogLevelAddr    string `mapstructure:"log_level_addr"`
	LogLevelPattern string `mapstructure:"log_level_pattern"`

	Bind           string
	Port           int
	AppendOnly     bool   `mapstructure:"append_only"`
	AppendFilename string `mapstructure:"append_filename"`
	MaxClients     int    `mapstructure:"max_clients"`
	RequirePass    string `mapstructure:"require_pass"`
	Databases      int

	Peers []string `cfg:"peers"`
	Self  string   `cfg:"self"`
}

func LoadConfig(paths ...string) {
	if len(paths) == 0 {
		viper.AddConfigPath(".")
		viper.AddConfigPath("config")
		viper.AddConfigPath("../config")
		viper.AddConfigPath("../../config")
	} else {
		for _, path := range paths {
			viper.AddConfigPath(path)
		}
	}

	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("social")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("mode", "debug")
	viper.SetDefault("log_level", "info")
	viper.SetDefault("log_path", "log")
	viper.SetDefault("log_level_pattern", "/log/level")
	viper.SetDefault("atomic_level_addr", "4240")

	viper.SetDefault("http.addr", ":8090")

	if err := viper.ReadInConfig(); err != nil {
		log.Panicf("read config error: %v", err)
	}
	if err := viper.Unmarshal(Conf); err != nil {
		log.Panicf("unmarshal config err: %v", err)
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Printf("config change: %s, %s, %s\n", e.Op.String(), e.Name, e.String())
		if err := viper.Unmarshal(Conf); err != nil {
			log.Printf("config change unmarshal err: %v", err)
		}
	})
	log.Println("load config successfully")
}
