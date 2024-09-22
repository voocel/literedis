package config

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"log"
	"strings"
	"time"
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

	// 新增 RDB 相关配置
	RDB RDBConfig `mapstructure:"rdb"`
}

type RDBConfig struct {
	Filename         string        `mapstructure:"filename"`
	SaveInterval     time.Duration `mapstructure:"save_interval"`
	CompressionLevel int           `mapstructure:"compression_level"`
	AutoSaveChanges  int           `mapstructure:"auto_save_changes"`
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

	// 添加 RDB 相关的默认值
	viper.SetDefault("rdb.filename", "dump.rdb")
	viper.SetDefault("rdb.save_interval", "5m")
	viper.SetDefault("rdb.compression_level", 6) // gzip 默认压缩级别
	viper.SetDefault("rdb.auto_save_changes", 1000)

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

// 新增方法，用于获取 RDB 配置
func GetRDBConfig() RDBConfig {
	return Conf.RDB
}
