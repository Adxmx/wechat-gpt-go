package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	OpenAPI struct {
		Key         string  `yaml:"key"`
		Model       string  `yaml:"model"`
		Temperature float32 `yaml:"temperature"`
		MaxTokens   int     `yaml:"max_tokens"`
	} `yaml:"open_api"`

	ProxyAddr struct {
		Ip   string `yaml:"ip"`
		Port int    `yaml:"port"`
	} `yaml:"proxy_addr"`

	RobotInfo struct {
		Admin    []string `yaml:"admin"`
		Prologue string   `yaml:"prologue"`
		Nickname string
	} `yaml:"robot_info"`
}

// Conf 配置文件
var Conf Config

func RobotFillCallback(nickname string) {
	Conf.RobotInfo.Nickname = nickname
	Conf.RobotInfo.Prologue = fmt.Sprintf(Conf.RobotInfo.Prologue, nickname)
}

func init() {
	// 加载配置文件
	yamlFile, _ := ioutil.ReadFile("config.yaml")
	yaml.Unmarshal(yamlFile, &Conf)
	fmt.Println(Conf)
}
