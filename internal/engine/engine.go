package engine

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
  "path/filepath"

	"github.com/AlexSTJO/flume/internal/infra"
	"github.com/AlexSTJO/flume/internal/logging"
	"github.com/AlexSTJO/flume/internal/structures"
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
  
  infra_outputs, err := infra.Deploy(e.Flume.Infrastructure, &logger)
  if err != nil {
    logger.ErrorLogger(err)
  }


  ctx := structures.NewContext()
  home, err := os.UserHomeDir()
  if err != nil {
    logger.ErrorLogger(fmt.Errorf("Failed to get home dir: %w", err))
  }
  flume_info := make(map[string]string, 1)
  flume_path := filepath.Join(home, ".flume", e.FlumeName)
  flume_info["path"] = flume_path 
  ctx.SetEventValues("flume_info", flume_info)

  
  logger.InfoLogger("Graphing Runtime")
  fmt.Println()
  g, err := structures.BuildGraph(e.Flume) 
  if err != nil {
    logger.ErrorLogger(err)
  }
  
  if g == nil || len(g.Nodes) == 0 {
    logger.ErrorLogger(fmt.Errorf("Graph is empty"))
  }

  // Will have to set up maxparallel specification in Config

  in := make(map[string]int, len(g.InDeg))
  for n, v := range(g.InDeg) {
    in[n] = v
  }

  ready := make(chan string, len(g.Nodes))
  var wg sync.WaitGroup

  for n,v := range(in) {
    if v == 0 {
      ready <- n
    }
  }


  var (
    mu sync.Mutex
    completed int
    closeOnce sync.Once
  )

  markDone := func(u string) {
    for _, v := range(g.Adj[u]) {
      mu.Lock()
      in[v]--
      if in[v] == 0 {
        ready <- v
      }
    mu.Unlock()
    }
  }


  worker := func() {
    defer wg.Done()
    for name := range(ready){
      logger.InfoLogger(fmt.Sprintf("Worker recieved task: %s", name))
      task := g.Nodes[name]
      svc, ok := structures.Registry[task.Service]
      if !ok {
        logger.ErrorLogger(fmt.Errorf("Unrecognized service"))
      }
      if err = svc.Run(task, name, ctx, infra_outputs, &logger); err != nil {
        close(ready)
        logger.ErrorLogger(err)
      } 
      markDone(name)
      
      mu.Lock()
      completed++
      done := completed == len(g.Nodes) 
      mu.Unlock()

      if done {
       closeOnce.Do(func() { close(ready) })
      }

    }
  }

  maxParallel := runtime.NumCPU()
  
  wg.Add(maxParallel)
  for i := 0; i < maxParallel; i++ {
    go worker()
  }
  
  wg.Wait()

  if completed != len(g.Nodes) {
    logger.ErrorLogger(fmt.Errorf("cycle detected: only completed %d of %d tasks", completed, len(g.Nodes)))
	}

  logger.SuccessLogger("Flume Completed")
  return nil
}  



