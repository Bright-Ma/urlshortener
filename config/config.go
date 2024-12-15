package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Server    ServerConfig    `mapstructure:"server"`
	App       AppConfig       `mapstructure:"app"`
	ShortCode ShortCodeConfig `mapstructure:"shortcode"`
	Logger    LogConfig       `mapstructure:"logger"`
	Email     EmailConfig     `mapstructure:"email"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	RandNum   RandNumConfig   `mapstructure:"rand_num"`
}

var Cfg *Config

func NewConfig(filePath string) (*Config, error) {
	viper.SetConfigFile(filePath)

	viper.SetEnvPrefix("URL_SHORTENER")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

type DatabaseConfig struct {
	Driver       string `mapstructure:"driver"`
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"dbname"`
	SSLMode      string `mapstructure:"ssl_mode"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=%s", d.Driver, d.User, d.Password, d.Host, d.Port, d.DBName, d.SSLMode)
}

type LogConfig struct {
	Level string `mapstructure:"level"`
}

type RandNumConfig struct {
	Length int `mapstructure:"length"`
}

type JWTConfig struct {
	Secret   string        `mapstructure:"secret"`
	Duration time.Duration `mapstructure:"duration"`
}

type RedisConfig struct {
	Address           string        `mapstructure:"address"`
	Password          string        `mapstructure:"password"`
	DB                int           `mapstructure:"db"`
	UrlDuration       time.Duration `mapstructure:"url_duration"`
	EmailCodeDuration time.Duration `mapstructure:"email_code_duration"`
}

type EmailConfig struct {
	Password    string `mapstructure:"password"`
	Username    string `mapstructure:"username"`
	HostAddress string `mapstructure:"host_address"`
	HostPort    string `mapstructure:"host_port"`
	Subject     string `mapstructure:"subject"`
	TestMail    string `mapstructure:"test_mail"`
}

type ServerConfig struct {
	Addr         string        `mapstructure:"addr"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
}

type AppConfig struct {
	BaseURL          string        `mapstructure:"base_url"`
	DefaultDuration  time.Duration `mapstructure:"default_duration"`
	SyncViewDuration time.Duration `mapstructure:"sync_view_duration"`
}

type ShortCodeConfig struct {
	Length int `mapstructure:"length"`
}
