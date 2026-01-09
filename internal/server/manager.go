package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/AlexSTJO/flume/internal/engine"
	"github.com/AlexSTJO/flume/internal/structures"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type runRequest struct{
  PipelineRef string `json:"pipeline_ref"`
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

  var s3_client  *s3.Client

  run_info, err := structures.GenerateRunInfo(req.PipelineRef)
  if err != nil {
    http.Error(w, "Error Geerating Run Info: " + err.Error(), http.StatusBadRequest)
    return
  }

  path := filepath.Join(".", ".flume", run_info.Pipeline, run_info.Pipeline + "yaml")
  if run_info.Remote {
    aws_ctx := context.Background()
    cfg, err := config.LoadDefaultConfig(aws_ctx)
    if err != nil {
      http.Error(w, "Failed To Load AWS Config" + err.Error(), http.StatusBadRequest)
      return 
    }
    s3_client = s3.NewFromConfig(cfg)
    
    out, err := s3_client.GetObject(aws_ctx, &s3.GetObjectInput{
      Bucket: aws.String(run_info.S3.Bucket),
      Key: aws.String(run_info.S3.Key),
    })

    if err != nil {
      err_string := fmt.Sprintf("Failed to get key: '%s' from bucket: '%s' -> %s", run_info.S3.Key, run_info.S3.Bucket, err.Error())
      http.Error(w, err_string, http.StatusBadRequest)
      return
    }

    defer out.Body.Close()

    b, err := io.ReadAll(out.Body)
    if err != nil {
      http.Error(w, "Unable to read file body" + err.Error(), http.StatusBadRequest)
      return
    }

    dir := filepath.Dir(path)
    if err := os.MkdirAll(dir, 0o755); err != nil {
      http.Error(w, "Error creating directories for remote file",http.StatusBadRequest)
      return
    }

    if err := os.WriteFile(path, b, 0o644); err != nil {
      http.Error(w, "Writing to tmp failed", http.StatusBadRequest)
      return
    }   
  }

  p, err := structures.Initialize(path)
  if err != nil {
    http.Error(w, "Pipeline Initalization Failure:" + err.Error(), http.StatusBadRequest)
    return
  }

  e, err := engine.Build(p, run_info)
  if err != nil {
    http.Error(w, "Engine Build Failure: " +err.Error(), http.StatusBadRequest)
    return
  }
  if err = e.Start(); err != nil {
    http.Error(w, "Engine Run Failure: " + err.Error(), http.StatusBadRequest)
    return
  }
}

