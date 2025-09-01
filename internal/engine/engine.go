package engine

import (
	"fmt"
	"sync"
	"time"

	"github.com/AlexSTJO/flume/internal/logging"
	"github.com/AlexSTJO/flume/internal/structures"
  "github.com/AlexSTJO/flume/internal/infra"
	"github.com/fatih/color"
  "github.com/joho/godotenv"
)


type Engine struct{
  FlumeName string
  RunID string
  Flume *structures.Pipeline
  LogPath string
  Context *structures.Context
}

func Build(p *structures.Pipeline) (*Engine, error) {  
  e := &Engine {
    FlumeName: p.Name,
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
  
  
  fmt.Printf("%s %s\n", label("Flume:"), value(e.FlumeName))
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

  err := godotenv.Load() 
	if err != nil {
		logger.ErrorLogger(fmt.Errorf("Error loading .env file"))
	}
  
  infrastructure, err := infra.Build()
  if err != nil {
    logger.ErrorLogger(err)
  }

  if err = infrastructure.CreateReferences(e.Flume.Infrastructure); err != nil {
    logger.ErrorLogger(err)
  }

  
  logger.InfoLogger("Graphing Runtime")
  fmt.Println()
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

    var wg sync.WaitGroup
    errCh := make(chan error, len(l))

    for _, n := range(l) {
      wg.Add(1)
      go func() {
        defer wg.Done()
        logger.InfoLogger(fmt.Sprintf("Running Task: %s", n))  
        t, ok := g.Nodes[n]
        if !ok {
          errCh <- fmt.Errorf("Unknown Task: %s", n)
          return
        }
        
        svc, ok := structures.Registry[t.Service]
        if !ok {
          errCh <- fmt.Errorf("Unknown service %s for task %s", t.Service, n)
          return
        }
      
        err = svc.Run(t, n, e.Context, &logger, infrastructure.TaskReferences)
        if err != nil {
          errCh <-fmt.Errorf("Error in task '%s':%v", n , err)
          return
        } 
        logger.InfoLogger("Task Ran Succesfully")
        
      }()
    }


    wg.Wait()
    close(errCh)

    for err := range errCh {
      if err != nil{
        logger.ErrorLogger(err)
        return err 
      }
    }
    
  }

  logger.SuccessLogger("Flume Completed")
  return nil
}  



