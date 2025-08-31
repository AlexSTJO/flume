package structures

import (
  "github.com/AlexSTJO/flume/internal/logging"
)

type Service interface {
  Name() string
  Parameters() []string
  Run(t Task, n string, ctx *Context, l *logging.Config, d map[string][]string) error
}


var Registry = map[string]Service{}

