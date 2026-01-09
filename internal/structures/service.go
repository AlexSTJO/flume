package structures

import (
  "github.com/AlexSTJO/flume/internal/logging"
)

type Service interface {
  Name() string
  Parameters() []string
  Run(t Task, n string, ctx *Context, infra_outputs *map[string]map[string]string, l *logging.Config, r *RunInfo) error 
}


var Registry = map[string]Service{}

