package engine

import (
	"fmt"
	"time"

	"github.com/AlexSTJO/flume/internal/logging"
	"github.com/AlexSTJO/flume/internal/structures"
	"github.com/fatih/color"
)


type Engine struct{
  TaskName string
  RunID string
  Flume *structures.Pipeline
  LogPath string
  Context *structures.Context
}

func Build(taskName string) (*Engine, error) {
  if taskName == "" {
    return nil, fmt.Errorf("How did you hit this")
  }


  filePath := fmt.Sprintf("%s.yaml", taskName)

  p, err := structures.Initialize(filePath)
  if err != nil{
    return nil, err
  }

  
  e := &Engine {
    TaskName: taskName,
    RunID: fmt.Sprintf("%d", time.Now().UnixNano()),
    Flume: p,
    LogPath: p.LogPath,
    Context: structures.NewContext(),
  }

  return e, nil
}


func (e *Engine) Start() error {
  label := color.New(color.FgGreen, color.Bold).SprintFunc()
  value := color.New(color.FgCyan).SprintFunc()
  warn := color.New(color.FgYellow).SprintFunc()
  
  
  fmt.Printf("%s %s\n", label("Flume:"), value(e.TaskName))
  fmt.Printf("%s %s\n", label("ID:"), value(e.RunID))

  if e.LogPath != "" {
    fmt.Printf("%s %s\n", label("Logs:"), value(e.LogPath))
  } else {
    fmt.Printf("%s %s\n", label("Logs:"), warn("No log file specified"))
  }

  logger := logging.Config{
    NoColor: false,
    LogPath: e.LogPath,
  } 
  
  logger.InfoLogger("Graphing Runtime")
  g, err := structures.BuildGraph(e.Flume) 
  if err != nil {
    logger.ErrorLogger(err)
  }


  levels, err := g.Levels()
  if err != nil {
    logger.ErrorLogger(err)
  }


  for _, l := range(levels) {
    if len(l) == 0 {
      continue
    }
    for _, n := range(l) {
      logger.InfoLogger(fmt.Sprintf("Running Node: %s", n))  
      t, ok := g.Nodes[n]
      if !ok {
        logger.ErrorLogger(fmt.Errorf("Weird Task Reference Error"))
      }
      
      svc, ok := structures.Registry[t.Service]
      if !ok {
        logger.ErrorLogger(fmt.Errorf("Weird Registry Reference Error"))
      }
    
      err = svc.Run(t)
      if err != nil {
        logger.ErrorLogger(fmt.Errorf("Error in task '%s':%v", n , err))
      } else {
        logger.InfoLogger("Task Ran Succesfully")
      }

    }
  }
  return nil
}  
  

  
  





