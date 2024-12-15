package main

import (
	"github.com/aeilang/urlshortener/application"
)

func main() {
	// 初始化应用程序，从配置文件加载配置
	app, err := application.InitApp("./config/config.yaml")
	if err != nil {
		panic(err)
	}

	// 启动应用程序
	app.Start()
}
