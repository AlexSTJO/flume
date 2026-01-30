package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlexSTJO/flume/internal/logging"
	"github.com/AlexSTJO/flume/internal/resolver"
	"github.com/AlexSTJO/flume/internal/structures"
)

type JsonWriterService struct{}

func (s JsonWriterService) Name() string {
	return "json_writer"
}

func (s JsonWriterService) Parameters() []string {
	return []string{"file_name", "data"}
}

func (s JsonWriterService) Run(t structures.Task, n string, ctx *structures.Context, infra_outputs *map[string]map[string]string, l *logging.Config, r *structures.RunInfo) error {
	runCtx := make(map[string]string, 1)
	defer ctx.SetEventValues(n, runCtx)
	runCtx["success"] = "false"

	u_file_name, err := t.StringParam("file_name")
	if err != nil {
		return err
	}
	file_name, err := resolver.ResolveStringParam(u_file_name, ctx, infra_outputs, r)
	if err != nil {
		return err
	}

	if !strings.HasSuffix(file_name, ".json") {
		file_name = file_name + ".json"
	}

	json_file := filepath.Join(r.RunDir, "job_outputs", n, file_name)

	payload := make(map[string]any)
	data, ok := t.Parameters["data"].(map[string]any)
	if !ok {
		return fmt.Errorf("json data is bad")
	}
	for k, v := range data {
		payload[k] = v
	}

	resolved_payload, err := resolver.ResolveAny(payload, ctx, infra_outputs, r)
	if err != nil {
		return fmt.Errorf("Error resolving data: %w", err)
	}

	dir := filepath.Dir(json_file)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating directory @ %s: %w", json_file, err)
	}
	f, err := os.Create(json_file)
	if err != nil {
		return fmt.Errorf("creating file @ %s: %w", json_file, err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")

	if err := enc.Encode(resolved_payload); err != nil {
		return fmt.Errorf("encoding json: %w", err)
	}

	runCtx["json_path"] = filepath.Dir(json_file)
	runCtx["success"] = "true"

	return nil

}

func init() {
	structures.Registry["json_writer"] = JsonWriterService{}
}
