package conf

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/kr/pretty"
	"github.com/spf13/viper"
)

const (
	Test = "test"
	Pro  = "pro"
)

var (
	config *Config
	once   sync.Once
)

type Config struct {
	Server  string  `yaml:"server"`
	Postgre Postgre `yaml:"postgre"`
	Redis   Redis   `yaml:"redis"`
	MinIO   MinIO   `yaml:"minio"`
}

type Postgre struct {
	DSN         string `yaml:"dsn"`
	PoolSize    int    `yaml:"poolSize"`
	PoolMaxTime int64  `yaml:"poolMaxTime"`
}

type Redis struct {
	Addr string `yaml:"addr"`
}

type MinIO struct {
	Endpoint  []string `yaml:"endpoint"`
	SecretKey string   `yaml:"secretKey"`
	AccessKey string   `yaml:"accessKey"`
}

func GetConf() *Config {
	once.Do(func() {
		initConf()
	})

	return config
}

func initConf() {
	prefix := "conf"
	env := getEnv()
	cfgFile := filepath.Join(prefix, filepath.Join(env, "conf.yml"))
	viper.SetConfigFile(cfgFile)
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	config = new(Config)
	if err := viper.Unmarshal(&config); err != nil {
		panic(err)
	}

	if env == Test {
		pretty.Printf("%v\n", config)
	}
}

func getEnv() string {
	env := os.Getenv("GO_ENV")
	if env != "" {
		return env
	}

	return Test
}
