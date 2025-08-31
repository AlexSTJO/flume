package structures

import (
  "fmt" 
  "os" 
  "gopkg.in/yaml.v3"
)


type Pipeline struct { 
  Name string `yaml:"name"`
  Tasks map[string]Task `yaml:"tasks"`
  LogPath string `yaml:"log_path"`
  Trigger TriggerSpec `yaml:"trigger"`
  Infrastructure map[string]Deployment `yaml:"infrastructure"`
}

type Task struct {
  Version int `yaml:"version"`
  Service string `yaml:"service"`
  Dependencies []string `yaml:"dependencies"`
  Parameters map[string]string `yaml:"parameters"`
}

type TriggerSpec struct {
  Type string `yaml:"type"`
  CronExpression string `yaml:"cron_expression,omitempty"`
  Timezone string `yaml:"tz,omitempty"`
  Path string `yaml:"path,omitempty"`
}

type Deployment struct {
  Service string `yaml:"service"`
  Action string `yaml:"action"`
  Tags map[string][]string `yaml:"tags,omitempty"`
  States []string `yaml:"states,omitempty"`
}
  

func Initialize(filepath string) (*Pipeline, error) {
  var p Pipeline
  data, err := os.ReadFile(filepath)
  if err != nil {
    return nil, fmt.Errorf("Error reading filepath: %w", err)
  }

  err = yaml.Unmarshal(data, &p)
  if err != nil {
    return nil, fmt.Errorf("Error unmarshalling yaml file: %w", err)
  }

  err = validateTasks(p.Tasks)
  if err != nil {
    return nil, fmt.Errorf("Error validating tasks: %w", err)
  }

  return &p, nil
}

func validateTasks(t map[string]Task) error{
  for _, task := range(t) {
    s ,ok := Registry[task.Service] 
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



