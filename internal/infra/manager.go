package infra

import (
	"context"
  "fmt"

  "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)


type DeploymentInfra struct {
  Config aws.Config
  TaskReferences map[string][]string
}


func New() (*DeploymentInfra, error) {
  cfg, err := config.LoadDefaultConfig(context.TODO())
  if err != nil {
    return nil,  err
  }
  return &DeploymentInfra{
    Config: cfg,
  }, nil
}


func (d *DeploymentInfra) Ec2Lookup() error {
  client := ec2.NewFromConfig(d.Config)

  o, err := client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
  if err != nil {
    return err
  }

  for _, reservation := range o.Reservations {
    for _, instance := range reservation.Instances {
      fmt.Printf("Instance ID: %s, State: %s\n", *instance.InstanceId, instance.State.Name)
    }
  }
      
  return nil   

}
