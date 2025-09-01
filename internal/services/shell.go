package services

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/AlexSTJO/flume/internal/logging"
	"github.com/AlexSTJO/flume/internal/resolver"
	"github.com/AlexSTJO/flume/internal/structures"
)

type ShellService struct {}


func (s ShellService) Name() string {
  return "shell"
}

func (s ShellService) Parameters() []string {
  return []string{"command"}
} 

func (s ShellService) Run(t structures.Task, n string,  ctx *structures.Context, l *logging.Config, d map[string][]string) error { 
  ids := make([]string,0)
  if t.Resources != nil{
    for _, r := range(t.Resources) {
      ids = append(ids, d[r]...)
    }
    l.InfoLogger("Found resource IDs")
  }

      
  rContext := make(map[string]string, 2)
  command, err := resolver.ResolveString(t.Parameters["command"], ctx, d)
  
  if err != nil {
    rContext["success"] = "false"
    ctx.SetEventValues(n,rContext)
    return err
  }

  cmd := exec.Command("sh", "-c", command)

  var outBuf, errBuf bytes.Buffer
  
  cmd.Stdout = &outBuf
  cmd.Stderr = &errBuf
  cmd.Stdin = os.Stdin
  
  err = cmd.Run()

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




