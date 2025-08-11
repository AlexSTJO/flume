package main

import (
  "fmt"

  "github.com/AlexSTJO/flume/internal/reader"
)

func main(){
  p, err := reader.Reader("sample.yaml")
  if err != nil {
    fmt.Println(err)
  } else {
    fmt.Println(p)
  }
  
  err = reader.ValidateTasks(p.Tasks)
  if err != nil {
    fmt.Printf("Error validating task: \n - %v \n", err)
  }
}
