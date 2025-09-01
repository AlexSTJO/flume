package services

import (
  "github.com/AlexSTJO/flume/internal/structures"
  "github.com/AlexSTJO/flume/internal/logging"
)


func Name() string {
  return "send_email"
}

func Parameters() []string {
  return []string{"username", "password", "port", "host", "subject", "body"}
}

func Run(t structures.Task, n string, ctx *structures.Context, l *logging.Config, d map[string][]string) error {
  
  
  return nil
}
