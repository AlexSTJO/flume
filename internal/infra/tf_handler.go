package infra

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
  "github.com/AlexSTJO/flume/internal/logging"
	"github.com/AlexSTJO/flume/internal/structures"
  "errors"
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

func (t *Terraform) Call(d structures.Deployment, l *logging.Config) (error) {
  switch d.Action {
  case "sync": {
    if err := TerraformInit(d.Key); err !=nil {
      return fmt.Errorf("Terraform Init Failed: %w", err)
    }
    l.InfoLogger("Terraform Initialization Succesful")

    changes, err := TerraformPlan(d.Key, d.VarFile)
    if err != nil{
      return fmt.Errorf("Terraform Plan Failed: %w", err)
    }

    if changes{
      l.InfoLogger("Changes can be made")
    } else {
      l.InfoLogger("Terraform Modules Up To Date With Infrastructure")
    }
    
  } 
  default:
    return fmt.Errorf("Unknown Action: %s", d.Action)
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

  _, err = cmd.CombinedOutput()

  if err != nil {
    return fmt.Errorf("terraform_init failed: %w", err)
  }
  
  return nil
}


func TerraformPlan(key string, var_file string) (bool, error) {
  home, err := os.UserHomeDir()
  if err != nil {
    return false, err
  }

  cmd := exec.Command("terraform", "plan", "-detailed-exitcode", "-input=false", "-no-color")
  cmd.Dir = filepath.Join(home, key)
  cmd.Env = append(os.Environ(), "TF_IN_AUTOMATION=1")

  _, err = cmd.CombinedOutput()

  if err == nil {
    return false, nil
  }
  var exitErr *exec.ExitError
  if errors.As(err, &exitErr) {
    switch exitErr.ExitCode() {
      case 0:
        return false, nil
      case 2:
        return true, nil
      default:
        return false, fmt.Errorf("Terraform Plan failed with exit code %d: %w", exitErr.ExitCode(), err)
    }
  }
  return false, fmt.Errorf("running terraform plan: %w", err)
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
