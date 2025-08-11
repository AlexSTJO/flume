package reader

import (
  "fmt" 
  "os"

  "github.com/AlexSTJO/flume/internal/services"
  "gopkg.in/yaml.v3"
)


type Pipeline struct { 
  Tasks map[string]Task `yaml:"tasks"`
}

type Task struct {
  Version int `yaml:"version"`
  Service string `yaml:"service"`
  Dependents []string `yaml:"dependents"`
  Parameters map[string]string `yaml:"parameters"`
}



func Reader(filepath string) (Pipeline, error){
  data, err := os.ReadFile(filepath)
  if err != nil {
    return Pipeline{} ,fmt.Errorf("Error Reading: %w", err)
  }

  var p Pipeline
  err = yaml.Unmarshal(data, &p)
  if err != nil{
    return Pipeline{} ,fmt.Errorf("Error Unmarshalling: %w", err)
  }

  return p, nil

}

func ValidateTasks(t map[string]Task) error{
  for _, task := range(t) {
    s ,ok := services.Registry[task.Service] 
    if !ok {
      return fmt.Errorf("Invalid Service Name: %s", task.Service)
    } 

    if len(task.Parameters) != len(s.Parameters()) {
      return fmt.Errorf("Extra Parameters in Task")
    }

    for _, p := range(s.Parameters()) {
      _, ok := task.Parameters[p]
      if !ok {
        return fmt.Errorf("Parameter Name Not Found: %s", p)
      }
    }
      }

  return nil
}
