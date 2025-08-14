package main

import (
	"fmt"

  "github.com/AlexSTJO/flume/internal/engine"
  "github.com/AlexSTJO/flume/internal/logging"
  "github.com/AlexSTJO/flume/internal/services"
)

func main(){
  var shell services.ShellService
  _ = shell.Name()
  e, err := engine.Build("sample")
  if err != nil {
    fmt.Printf("An error occurred: %v", err)
  }


  logger := logging.Config{
    NoColor: false,
    LogPath: e.LogPath,
  }

  err = e.Start()
  if err != nil {
    logger.ErrorLogger(err)
  }
  

  
}
