package server

import (
	"crypto/sha256"
	"fmt"
  "io"
  "path/filepath"
  "os"

	"github.com/AlexSTJO/flume/internal/structures"
)

type Trigger struct {
  Type string `json:"type"`
  Cron string `json:"cron,omitempty"`
}

type PipelineMeta struct {
  Name string `json:"name"`
  YamlPath string `json:"yaml_path"`
  Trigger Trigger `json:"trigger"`
  Enabled bool `json:"enabled"`
  YamlSha256 string `json:"yaml_sha256"`
  YamlMtimeUnix int64 `json:"yaml_mtime_unix"`
}



func generateMeta(yamlPath string) (*PipelineMeta, error) {

  home, err := os.UserHomeDir()
  if err != nil {
    return nil, fmt.Errorf("Error getting home directory: %w", err)
  }

  absPath := filepath.Join(home, ".flume", yamlPath)

  p, err := structures.Initialize(absPath)
  if err != nil{
    return nil, fmt.Errorf("Error creating flume: %w", err)
  }

  info, err := os.Stat(absPath)
  if err != nil {
		return nil, fmt.Errorf("Error stat yaml: %w", err)
	}

  sum ,err := fileSHA256(absPath)
  if err != nil {
    return nil, fmt.Errorf("Error hashing yaml: %w", err)
  }
  

  t := Trigger{
    Type: p.Trigger.Type,
    Cron: p.Trigger.CronExpression,
  }
  pm := &PipelineMeta{
    Name: p.Name,
    YamlPath: yamlPath,
    Trigger: t,
    Enabled: true,
    YamlSha256: sum,
    YamlMtimeUnix: info.ModTime().Unix(),
  }

  return pm, nil
}



func fileSHA256(path string) (string, error){
  f, err := os.Open(path)
  if err != nil {
    return "", err
  }
  defer f.Close()
  h := sha256.New()
  if _, err := io.Copy(h, f); err != nil {return "", err } 
  return fmt.Sprintf("%x", h.Sum(nil)), nil
    

}

  
  



