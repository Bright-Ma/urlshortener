package main

import "github.com/aeilang/urlshortener/application"

func main() {
	a := application.Application{}
	if err := a.Init("./config/config.yaml"); err != nil {
		panic(err)
	}
	a.Run()
}
