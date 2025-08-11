package main

import (
  "fmt"

  "github.com/AlexSTJO/internal"
)

func main(){
  p, err := parser.Reader("sample.yaml")
  if err != nil {
    fmt.Printf("Error")
  } else {
    fmt.Println(p)
  }
}
