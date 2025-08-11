package reader

import (
  "fmt"
  
  "os"
  "gopkg.in/yaml.v3"
)


type Pipeline struct { 
  tasks map[string]Task `yaml:"tasks`
}

type Task struct {
  version int `yaml:"version"`
  service string `yaml:"service"`
  dependents []string `yaml:"dependents`
  parameters []string `yaml:"parameters`
}



func Reader(filepath) (struct Pipeline, err){
  data, err := os.ReadFile(filepath)
  if err != nil {
    return nil ,fmt.Errorf("Error Reading")
  }
  p := Pipeline{}
  err := yaml.Unmarshal(data, &p)
  if err != nil{
    return nil ,fmt.Errorf("Error Unmarshalling")
  }


  return p, nil

}
