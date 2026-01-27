package server

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/AlexSTJO/flume/internal/structures"
)

type Trigger struct {
	Type string `json:"type"`
	Cron string `json:"cron,omitempty"`
}

type PipelineMeta struct {
	Name          string  `json:"name"`
	YamlPath      string  `json:"yaml_path"`
	Trigger       Trigger `json:"trigger"`
	Enabled       bool    `json:"enabled"`
	YamlSha256    string  `json:"yaml_sha256"`
	YamlMtimeUnix int64   `json:"yaml_mtime_unix"`
}

func generateMeta(absPath string) (*PipelineMeta, error) {
	p, err := structures.Initialize(absPath)
	if err != nil {
		return nil, fmt.Errorf("Error creating flume: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("Error stat yaml: %w", err)
	}

	sum, err := fileSHA256(absPath)
	if err != nil {
		return nil, fmt.Errorf("Error hashing yaml: %w", err)
	}

	t := Trigger{
		Type: p.Trigger.Type,
		Cron: p.Trigger.CronExpression,
	}
	pm := &PipelineMeta{
		Name:          p.Name,
		YamlPath:      absPath,
		Trigger:       t,
		Enabled:       true,
		YamlSha256:    sum,
		YamlMtimeUnix: info.ModTime().Unix(),
	}

	return pm, nil
}

func writeMeta(pm *PipelineMeta) error {
	dir := filepath.Dir(pm.YamlPath)
	metaPath := filepath.Join(dir, "meta.json")

	if err := os.MkdirAll(filepath.Dir(metaPath), 0o755); err != nil {
		return fmt.Errorf("Error creating metadata dir: %w", err)
	}

	data, err := json.MarshalIndent(pm, "", "  ")
	if err != nil {
		return fmt.Errorf("Error Marshalling metadata: %w", err)
	}

	tmp := metaPath + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write tmp: %w", err)
	}
	if err := os.Rename(tmp, metaPath); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename meta: %w", err)
	}
	return nil

}

func SyncMeta(yamlPath string) (bool, error, *PipelineMeta) {
	newPM, err := generateMeta(yamlPath)
	if err != nil {
		return false, fmt.Errorf("Error generating metadata: %v", err), nil
	}

	dir := filepath.Dir(yamlPath)
	metaPath := filepath.Join(dir, "meta.json")

	oldPM, _ := ReadMeta(metaPath)

	if oldPM != nil &&
		oldPM.YamlSha256 == newPM.YamlSha256 &&
		oldPM.YamlMtimeUnix == newPM.YamlMtimeUnix &&
		oldPM.Trigger == newPM.Trigger &&
		oldPM.Enabled == newPM.Enabled &&
		oldPM.Name == newPM.Name &&
		oldPM.YamlPath == newPM.YamlPath {
		return false, nil, oldPM
	}

	if err := writeMeta(newPM); err != nil {
		return false, err, nil
	}

	return true, nil, newPM

}

func ReadMeta(metaPath string) (*PipelineMeta, error) {
	b, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}
	var pm PipelineMeta
	if err := json.Unmarshal(b, &pm); err != nil {
		return nil, err
	}
	return &pm, nil
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
