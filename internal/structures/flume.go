package structures

import (
  "fmt" 
  "os" 
  "gopkg.in/yaml.v3"
)


type Pipeline struct { 
  Name string `yaml:"name"`
  Tasks map[string]Task `yaml:"tasks"`
  DisableLogging bool `yaml:"disable_logging,omitempty"`
  Trigger TriggerSpec `yaml:"trigger"`
  Infrastructure map[string]Deployment `yaml:"infrastructure"`
}

type Task struct {
  Version int `yaml:"version"`
  Service string `yaml:"service"`
  Dependencies []string `yaml:"dependencies"`
  RunIf string `yaml:"run_if,omitempty"`
  SkipIf string `yaml:"skip_if,omitempty"`
  Parameters map[string]any `yaml:"parameters"`
  Resources []string `yaml:"resources,omitempty"`
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
  Repo string `yaml:"repo"`
  VarFile string `yaml:"var-file,omitempty"`
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
      return fmt.Errorf("Extra Parameters in Task: %s", task.Service)
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

func (t Task) StringParam(key string) (string, error) {
    v, ok := t.Parameters[key]
    if !ok {
        return "", fmt.Errorf("missing parameter %q", key)
    }

    s, ok := v.(string)
    if !ok {
        return "", fmt.Errorf("parameter %q must be string, got %T", key, v)
    }

    return s, nil
}



