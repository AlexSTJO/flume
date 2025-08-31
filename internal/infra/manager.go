package infra

import (
	"context"
  "fmt"

	"github.com/AlexSTJO/flume/internal/structures"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

type DeploymentInfra struct {
  Config aws.Config
  TaskReferences map[string][]string
  Services map[string]Service
}

type Service interface{
  Name() string
  Call(d structures.Deployment) ([]string, error)  
}

type Factory func(aws.Config)Service

var registry = map[string]Factory{}

func Build() (*DeploymentInfra, error) {
  cfg, err := config.LoadDefaultConfig(context.TODO())
  if err != nil {
    return nil,  err
  }
  s := make(map[string]Service, len(registry))
  
  
  for k, config := range(registry) {
    fmt.Println(k)
    s[k] = config(cfg) 
  }


  return &DeploymentInfra{
    Config: cfg,
    Services: s,
    TaskReferences: make(map[string][]string),
  }, nil
}           


func (d *DeploymentInfra) CreateReferences(i map[string]structures.Deployment) (error) { 
  for k, deployment := range(i) {
    references, err := d.Services[deployment.Service].Call(deployment)
    if err != nil {
      return err
    }
    fmt.Println(references)
    d.TaskReferences[k] = references
  }
  return nil
}
  





