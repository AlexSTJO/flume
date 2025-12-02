package main

import (
  "github.com/AlexSTJO/flume/internal/services"
  "github.com/AlexSTJO/flume/internal/logging"
	"github.com/AlexSTJO/flume/internal/server"
)

func main(){
  var shell services.ShellService
  _ = shell.Name()
  c := logging.New("")
  if err := server.CreateServer(); err != nil {
    c.ErrorLogger(err)
  }
}
