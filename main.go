package main

import (
  "fmt"

  "github.com/AlexSTJO/flume/internal/structures"
  "github.com/AlexSTJO/flume/internal/services"
)

func main(){
  shell := services.ShellService{}
  p, err := structures.Initialize("sample.yaml")
  if err != nil {
    fmt.Println(err)
    return
  } 
  shell.Name() 

  g, err := structures.Build(p)
  if err != nil {
    fmt.Println(err)
    return
  }

  levels, err := g.Levels()
  if err != nil {
    fmt.Println(err)
    return
  }

  fmt.Println(levels)
}
