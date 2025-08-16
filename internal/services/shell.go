package services

import (
  "os/exec"
  "os"
  "fmt"
  "bytes"

  "github.com/AlexSTJO/flume/internal/structures"
  "github.com/AlexSTJO/flume/internal/logging"

)

type ShellService struct {}


func (s ShellService) Name() string {
  return "shell"
}

func (s ShellService) Parameters() []string {
  return []string{"command"}
} 

func (s ShellService) Run(t structures.Task, n string,  ctx *structures.Context, l *logging.Config) error { 
  rContext := make(map[string]string, 2)
  command := t.Parameters["command"]

  cmd := exec.Command("sh", "-c", command)

  var outBuf, errBuf bytes.Buffer
  
  cmd.Stdout = &outBuf
  cmd.Stderr = &errBuf
  cmd.Stdin = os.Stdin
  
  err := cmd.Run()

  if (err != nil) {
    rContext["success"] = "false"
    ctx.SetEventValues(n, rContext )
    return fmt.Errorf("Shell Error Occurred: %v", err)
  }


  if outBuf.Len() > 0 {
    l.ShellLogger(outBuf.String())
  }

  if errBuf.Len() > 0 {
    l.ShellLogger(errBuf.String())
  }


  rContext["success"] = "true"
  ctx.SetEventValues(n, rContext ) 
  return nil 

}

func init() {
  structures.Registry["shell"] = ShellService{}
}




