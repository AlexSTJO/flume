package infra

import (
	"github.com/AlexSTJO/flume/internal/structures"
  "fmt"
)

type Service interface{
  Name() string
  Call(d structures.Deployment) (error)  
}


var registry = map[string]Service{}        


func Deploy(i map[string]structures.Deployment) error {
    for _, deployment := range i {
        svc, ok := registry[deployment.Service]
        if !ok {
            return fmt.Errorf("unknown service %q", deployment.Service)
        }

        if err := svc.Call(deployment); err != nil {
            return err
        }
    }
    return nil
}
  
