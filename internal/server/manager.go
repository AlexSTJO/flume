package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/AlexSTJO/flume/internal/engine"
	"github.com/AlexSTJO/flume/internal/structures"
)

type runRequest struct{
  Pipeline string `json:"pipeline"`
}


func CreateServer() error {
  mux := http.NewServeMux()
  mux.HandleFunc("/run", runPipeline)
  go func() {
    if err := http.ListenAndServe(":8080", mux); err != nil {
      fmt.Printf("Error creating server: %w", err)
    }
  }()

  CronManager, err := CronInit()
  if err != nil {
    return fmt.Errorf("cron init failed: %w", err)
  }

  go func() { 
    if err := FileWatcher(CronManager); err != nil {
      fmt.Printf("Error creating file watcher: %w", err)
    }
  }()
      

  
  select {}
}

func runPipeline(w http.ResponseWriter, r *http.Request) {
  defer r.Body.Close()
  dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))

  var req runRequest
  if err := dec.Decode(&req); err != nil {
    http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
    return
  }
  
  home, err := os.UserHomeDir()
  if err != nil {
    http.Error(w, "failed to get home directory: "+err.Error(), http.StatusBadRequest)
    return
  }

  filepath := filepath.Join(home, ".flume", req.Pipeline, req.Pipeline+".yaml" )
  p, err := structures.Initialize(filepath)
  if err != nil {
    http.Error(w, "Pipeline Initalization Failure:" + err.Error(), http.StatusBadRequest)
    return
  }

  e, err := engine.Build(p)
  if err != nil {
    http.Error(w, "Engine Build Failure: " +err.Error(), http.StatusBadRequest)
    return
  }
  if err = e.Start(); err != nil {
    http.Error(w, "Engine Run Failure: " + err.Error(), http.StatusBadRequest)
    return
  }
}
