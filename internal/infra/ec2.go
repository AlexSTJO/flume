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
    case "fs": {
      ids, err := e.search(d)
      if err != nil {
        return nil, err
      }
      return ids, nil 
    }

    case "tfs": {
      ids, err := e.terraform_search(d)
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

func (e *EC2Service) search(d structures.Deployment) ([]string, error) {
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

func (e *EC2Service) terraform_search(d structures.Deployment)  ([]string, error) {
  var ids []string
  state, err := TerraformState(d.Key)
  if err != nil{
    return nil,err
  }
  
  for _,r := range(state.Resources) {
    if r.Type != "aws_instance" {
      continue
    } else {
      for _,i := range(r.Instances) {
        match := true
        for k, vs := range(d.Tags) {
          for _, v := range(vs) { 
            match = false
            if i.Attributes.Tags[k] == v {
              match = true
              break
            }
          }
        } 
        if match != false && len(d.States) > 0 {
          match = false
          for _, s := range(d.States) {
            if i.Attributes.State == s {
              match = true
              break
            }
          }
        }

        if match {
          ids = append(ids, i.Attributes.Id)
        }     
      }             
    }
  }
  return ids, nil
}


func init() {
  registry["ec2"] = Configure
}

