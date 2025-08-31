package infra

import (
  "context"
  "fmt"

  "github.com/AlexSTJO/flume/internal/structures"
  "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
  "github.com/aws/aws-sdk-go-v2/service/ec2/types" 
)


type EC2Service struct {
  client *ec2.Client 
} 


func (e *EC2Service) Name() string {
  return "ec2"
}

func (e *EC2Service) Call(d structures.Deployment) ([]string, error) {
  switch d.Action {
    case "check": {
      ids, err := e.check(d)
      if err != nil {
        return nil, err
      }
      return ids, nil 
    }
  }

  return nil, fmt.Errorf("Action not found")
}

func Configure(cfg aws.Config) (Service){
  return &EC2Service{
    client: ec2.NewFromConfig(cfg),
  }
}

func (e *EC2Service) check(d structures.Deployment) ([]string, error) {
  filters := []types.Filter{}

  for k, p := range(d.Tags) {
    k := fmt.Sprintf("tag:%s", k)
    t := types.Filter{
        Name: aws.String(k),
        Values: p,
    }

    filters = append(filters, t)
  }

  if d.States != nil {
    s := types.Filter {
      Name:  aws.String("instance-state-name"),
      Values: d.States,
    }
    filters = append(filters, s)
  }

  o, err := e.client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{
    Filters: filters,
  })

  if err != nil {
    return nil,err
  }

  var results []string

  for _, reservation := range o.Reservations {
    for _, instance := range reservation.Instances {
      results = append(results, *instance.InstanceId)
    }
  }
      
  return results, nil    

}



func init() {
  registry["ec2"] = Configure
}

