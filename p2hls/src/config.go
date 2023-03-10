package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Config struct {
	Port           int //`json:"port"`
	Host           string
	Parse2hlsPath  string
	Parse2HlsParam string
	//MainParamFlag  int

	DownLoadDir string
	CallBack    string
}

func InitConfig() {

	if len(gConfig.Host) <= 0 {
		gConfig.Host = "127.0.0.1"
	}

	if gConfig.Port == 0 {
		gConfig.Port = 9001
	}

	if len(gConfig.DownLoadDir) <= 0 {
		gConfig.DownLoadDir = "/opt/fonsview/NE/parse2hls/download/"
	}

	if len(gConfig.Parse2hlsPath) <= 0 {
		gConfig.Parse2hlsPath = "/opt/fonsview/NE/parse2hls/bin/parse2hls"
	}

	if len(gConfig.Parse2HlsParam) <= 0 {
		gConfig.Parse2HlsParam = "-f 2 "
	}

	logs.Info("config:%v", gConfig)
}

func LoadConfig(path string) int {
	// 读取配置文件
	data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("读取配置文件失败")
		return -1
	}

	err = json.Unmarshal(data, &gConfig)
	if err != nil {
		fmt.Println("解析配置文件失败")
		return -1
	}

	InitConfig()
	return 0

}
