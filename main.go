package main

import (
	"fmt"

	"github.com/AlexSTJO/flume/internal/engine"
	"github.com/AlexSTJO/flume/internal/logging"
	"github.com/AlexSTJO/flume/internal/services"
	"github.com/AlexSTJO/flume/internal/structures"
)

func main(){
  var shell services.ShellService
  _ = shell.Name()

  p, err := structures.Initialize("sample.yaml")
  if err != nil {
    fmt.Printf("An Error Occurred: %v\n", err)
  }

  c := logging.New(p.LogPath)
  e, err := engine.Build(p)
  if err != nil {
    c.ErrorLogger(err)
  }

  if err = e.Start(); err != nil {
    c.ErrorLogger(err)
  }

}
