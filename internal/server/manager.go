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
  if err := http.ListenAndServe(":8080", mux); err != nil {
    return fmt.Errorf("Error creating server: %w", err)
  }

  return nil
  
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
    http.Error(w, "Engine Run Failute: " + err.Error(), http.StatusBadRequest)
    return
  }
}
