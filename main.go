package main

import (
	"fmt"

  "github.com/AlexSTJO/flume/internal/logging"
	"github.com/AlexSTJO/flume/internal/services"
	"github.com/AlexSTJO/flume/internal/structures"
	"github.com/AlexSTJO/flume/internal/server"
)

func main(){
  var shell services.ShellService
  _ = shell.Name()

  p, err := structures.Initialize("sample.yaml")
  if err != nil {
    fmt.Printf("An Error Occurred: %v\n", err)
  }

  c := logging.New(p.LogPath)
  if err = server.CreateServer(); err != nil {
    c.ErrorLogger(err)
  }
}
