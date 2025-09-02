package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Host string
		Port string
		Env  string
	}
	Mysql struct {
		Dsn string
	}
	Redis struct {
		Addr string
		Pwd  string
		DB   int
	}
	Auth struct {
		AccessSecret  string
		AccessExpire  int
		RefreshExpire int
	}
	Logger struct {
		Level        string `yaml:"level"`
		Prefix       string `yaml:"prefix"`
		Director     string `yaml:"director"`
		ShowLine     bool   `yaml:"show_line"`      //是否显示行号
		LogInConsole bool   `yaml:"log_in_console"` //是否显示打印的路径
	}
}

func InitConf() *Config {
	var cfg Config
	viper.SetConfigName("settings")       // 文件名，不带扩展名
	viper.SetConfigType("yaml")           // 配置文件类型
	viper.AddConfigPath("./internal/etc") // 配置文件路径
	// 读取配置
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("读取配置文件失败: %s", err))
	}
	// 解析到结构体
	if err := viper.Unmarshal(&cfg); err != nil {
		panic(fmt.Errorf("解析配置文件失败: %s", err))
	}
	fmt.Println("✅ 配置加载成功:", viper.ConfigFileUsed())
	fmt.Printf("%+v\n", cfg)
	return &cfg
}
