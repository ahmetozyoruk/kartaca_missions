package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Server struct {
	Username string `mapstructure:"server_username"`
	Password string    `mapstructure:"server_password"`
	Host string    `mapstructure:"server_host"`
	Port string    `mapstructure:"server_port"`
	FilePath string    `mapstructure:"server_file_path"`
}

type Database struct {
	DBHost            string `mapstructure:"db_host"`
	DBPort            string `mapstructure:"db_port"`
	DBName            string `mapstructure:"db_name"`
	DBUsername        string `mapstructure:"db_username"`
	DBPassword        string `mapstructure:"db_password"`
}

type Website struct {
	Url            string `mapstructure:"url"`
}

type Configuration struct {
	Server   `mapstructure:"server"`
	Database `mapstructure:"database"`
	Website `mapstructure:"website"`

}



func InitConfig() Configuration {
	viper.AddConfigPath("/config")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Error reading config file: ", err)
		os.Exit(1)
	}

	var config Configuration

	err := viper.Unmarshal(&config)
	if err != nil {
		fmt.Println("Unable to decode into struct: ", err)
	}

	return config
}