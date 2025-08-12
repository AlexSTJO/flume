package services

import (
  "os/exec"
  "os"
  "fmt"

  "github.com/AlexSTJO/flume/internal/structures"
)

type ShellService struct {}


func (s ShellService) Name() string {
  return "shell"
}

func (s ShellService) Parameters() []string {
  return []string{"command"}
} 

func (s ShellService) Run(t structures.Task) error { 
  command := t.Parameters["command"]

  cmd := exec.Command("sh", "-c", command)
  
  cmd.Stdout = os.Stdout
  cmd.Stderr = os.Stderr
  cmd.Stdin = os.Stdin
  
  err := cmd.Run()

  if (err != nil) {
    return fmt.Errorf("Shell Error Occurred: %v", err)
  }
 

  return nil 

}

func init() {
  structures.Registry["shell"] = ShellService{}
}




