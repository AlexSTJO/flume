package infra

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
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

func ParseState(data []byte) (*State, error) {
  var s State
  if err := json.Unmarshal(data, &s); err != nil {
    return nil, err
  }
  return &s, nil
}
