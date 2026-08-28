package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	App struct {
		Name string
		Port string
	}
	Database struct {
		Dsn string
	}
	Redis struct {
		Host     string
		Port     int
		Password string
		DB       int
	}
}

// 声明一个全局变量，用来储存文件信息    不是使用全局变量而是通过调用链传参来传递文件信息
//var Conf *Config

// 好了这是配置文件的初步代码了；
func SetUpConfig() Config {
	//viper.SetConfigName("ConfigLocal")
	//viper.SetConfigType("yaml")
	//增加viper的识别路径
	//viper.AddConfigPath("./config")
	//viper.AddConfigPath(".")
	config := os.Getenv("CONFIG_PATH")
	if config == "" {
		config = "./config/config.local.yaml"
	}
	viper.SetConfigFile(config)
	//file, err2 := os.ReadFile(config)
	//yaml.Unmarshal(file, &Conf)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
	var cfg Config
	err := viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatalf("Error unmarshalling config, %s", err)
	}
	return cfg
}
