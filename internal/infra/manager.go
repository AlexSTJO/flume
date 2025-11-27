package infra

import (
	"github.com/AlexSTJO/flume/internal/structures"
  "github.com/AlexSTJO/flume/internal/logging"
  
  "fmt"
)

type Service interface{
  Name() string
  Call(d structures.Deployment, l *logging.Config) (error)  
}


var registry = map[string]Service{}        


func Deploy(i map[string]structures.Deployment,l *logging.Config) error {
    for _, deployment := range i {
        svc, ok := registry[deployment.Service]
        if !ok {
            return fmt.Errorf("unknown service %q", deployment.Service)
        }

        if err := svc.Call(deployment, l); err != nil {
            return err
        }
    }
    return nil
}
  
