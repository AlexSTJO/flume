package main

import (
	"fmt"

	"github.com/AlexSTJO/flume/internal/infra"
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
  di, err := infra.Build()
  if err != nil {
    c.ErrorLogger(err)
  }



  r, err := di.Services["ec2"].Call(p.Infrastructure["ec2Check"])
  if err != nil {
    c.ErrorLogger(err)
  }
  
  if err != nil {
    c.ErrorLogger(err)
  }

  for _, n := range(r) {
    fmt.Println(n)  
  }
}
