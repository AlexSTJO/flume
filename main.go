package main

import (
	"github.com/AlexSTJO/flume/internal/server"
	"github.com/AlexSTJO/flume/internal/services"
)

func main() {
	var shell services.ShellService
	_ = shell.Name()
	if err := server.CreateServer(); err != nil {
		panic(err)
	}
}
