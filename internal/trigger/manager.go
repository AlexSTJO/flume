package trigger

import (
	"fmt"
	"net/http"
  "log"

	"github.com/AlexSTJO/flume/internal/structures"
  "github.com/AlexSTJO/flume/internal/engine"
)

type TriggerManager struct {
  srv *http.Server
}

func New() TriggerManager {
  return TriggerManager{}
}


func (t *TriggerManager) Handle(p *structures.Pipeline) (error) {
  spec := p.Trigger
  if spec.Type == "" {
    return fmt.Errorf("No Trigger Type")
  }

  if spec.Type == "manual" {
    e, err := engine.Build(p)
    if err != nil {
      return err
    }
    err = e.Start()
    if err != nil {
      return err
    }
    return nil  
  } else if spec.Type == "api" {
    if spec.Path == "" {
      return fmt.Errorf("No specified path for listener")
    }
    t.createServer(spec.Path, p)
    return nil
  }
  return nil

}


func (t *TriggerManager) createServer(path string, p *structures.Pipeline) {
  CallHandler := func(w http.ResponseWriter, r *http.Request) {
    e , err := engine.Build(p)
    if err != nil {
      return 
    }
    err = e.Start()
    if err != nil {
      return 
    }
  }

  http.HandleFunc(path, CallHandler)
  log.Fatal(http.ListenAndServe(":8080",nil))
}
