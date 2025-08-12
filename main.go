package main

import (
  "fmt"

  "github.com/AlexSTJO/flume/internal/structures"
  "github.com/AlexSTJO/flume/internal/services"
)

func main(){
  shell := services.ShellService{}
  var p structures.Pipeline
  err := p.Initialize("sample.yaml")
  if err != nil {
    fmt.Println(err)
  } 
  
  shell.Run(p.Tasks["build"]) 
}
