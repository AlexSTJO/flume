package main

import (
  "github.com/AlexSTJO/flume/internal/services"
	"github.com/AlexSTJO/flume/internal/server"
)

func main(){
  var shell services.ShellService
  _ = shell.Name()
  if err := server.CreateServer(); err != nil {
    panic(err)
  }
}
