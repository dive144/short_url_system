package config

import (
	"fmt"
	"log"

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
}

// 声明一个全局变量，用来储存文件信息
var Conf *Config

// 好了这是配置文件的初步代码了；
func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	//增加viper的识别路径
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
	Conf = &Config{}
	err := viper.Unmarshal(Conf)
	if err != nil {
		log.Fatalf("Error unmarshalling config, %s", err)
	}
	initDb()
	fmt.Println(Conf.App.Name, Conf.App.Port)
}
