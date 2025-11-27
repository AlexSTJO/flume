package infra

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/AlexSTJO/flume/internal/structures"
)

type State struct{
  Resources []Resource `json:"resources"`
}

type Resource struct {
  Type string `json:"type"`
  Name string `json:"name"`
  Instances []Instance `json:"instances"`
}

type Instance struct {
  Attributes Attributes `json:"attributes"`
}

type Attributes struct {
  Id string `json:"id"`
  Tags map[string]string `json:"tags"`
  State string `json:"instance_state"`
}

type Terraform struct {}


func (t *Terraform) Name() string {
  return "terraform"
}

func (t *Terraform) Call(d structures.Deployment) (error) {
  switch d.Action {
    case "sync": {
    if err := TerraformInit(d.Key); err !=nil {
        return fmt.Errorf("Terraform Init Failed: %w", err)
      }
    } 
    default:
      return fmt.Errorf("Unknown Action: %w", d.Action)
  }

  return nil 
}


func TerraformState(key string) (*State, error) {
  cmd := exec.Command("terraform", "state", "pull")
  home, err := os.UserHomeDir()
  if err != nil {
    return nil, err  
  }

  dir := filepath.Join(home, key)
  cmd.Dir = dir

  out, err := cmd.Output()
  if err != nil{
    return nil, err
  }
  return ParseState(out)
}

func TerraformInit(key string) error {
  cmd := exec.Command("terraform", "init", "-input=false", "-no-color")
  home, err := os.UserHomeDir()
  if err != nil {
    return err
  }

  cmd.Dir = filepath.Join(home, key)
  cmd.Env = append(os.Environ(), "TF_IN_AUTOMATION=1")

  out, err := cmd.CombinedOutput()

  fmt.Println("terraform init output", "stdout", string(out))
  if err != nil {
    return fmt.Errorf("terraform_init failed: %w", err)
  }
  
  return nil 

}

func ParseState(data []byte) (*State, error) {
  var s State
  if err := json.Unmarshal(data, &s); err != nil {
    return nil, err
  }
  return &s, nil
}


func init() {
  registry["terraform"] = &Terraform{}
}
