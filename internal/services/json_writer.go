package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlexSTJO/flume/internal/logging"
	"github.com/AlexSTJO/flume/internal/resolver"
	"github.com/AlexSTJO/flume/internal/structures"
)


type JsonWriterService struct {}


func (s JsonWriterService) Name() string {
  return "json_writer"
}

func (s JsonWriterService) Parameters() []string {
  return []string{"data"}
}

func (s JsonWriterService) Run(t structures.Task, n string, ctx *structures.Context, infra_outputs *map[string]map[string]string, l*logging.Config) error{
  flume_info := ctx.GetEventValues("flume_info")
  flume_folder := flume_info["path"]
  json_file := filepath.Join(flume_folder, n, "meta.json")
  
  run_ctx := make(map[string]string, 1)

  payload := make(map[string]any)
  data, ok := t.Parameters["data"].(map[string]any)
  if !ok  { return fmt.Errorf("json data is bad")}
  for k, v := range data{
    payload[k] = v
  }

  resolved_payload, err := resolver.ResolveAny(payload, ctx, infra_outputs)
  if err != nil {return fmt.Errorf("Error resolving data: %w", err)}

  dir := filepath.Dir(json_file)
  if err := os.MkdirAll(dir, 0o755); err != nil {
    return fmt.Errorf("creating directory @ %s: %w", json_file, err)
  }
  f, err := os.Create(json_file)
  if err != nil {
    return fmt.Errorf("creating file @ %s: %w", json_file, err )
  }
  defer f.Close()

  enc := json.NewEncoder(f)
  enc.SetIndent("", "  ")

  if err := enc.Encode(resolved_payload); err != nil {
    return fmt.Errorf("encoding json: %w", err)
  }

  run_ctx["json_path"] = json_file
  ctx.SetEventValues(n, run_ctx)

  return nil

}


func init() {
  structures.Registry["json_writer"] = JsonWriterService{}
}
