package conf

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/joho/godotenv"
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
	Server string `yaml:"server"`
	MySQL  MySQL  `yaml:"mysql"`
	Redis  Redis  `yaml:"redis"`
	MinIO  MinIO  `yaml:"minio"`
}

type MySQL struct {
	DSN string `yaml:"dsn"`
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
	envFile := filepath.Join(prefix, ".env")

	err := godotenv.Load(envFile)
	if err != nil {
		panic(err)
	}

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
		pretty.Printf("%v", config)
	}
}

func getEnv() string {
	env := os.Getenv("GO_ENV")
	if env != "" {
		return env
	}

	return Test
}
