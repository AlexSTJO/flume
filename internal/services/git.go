package services

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlexSTJO/flume/internal/logging"
	"github.com/AlexSTJO/flume/internal/structures"

	"os/exec"
)

type GitService struct {}

func (s GitService) Name() string{
  return "git"
}


func (s GitService) Parameters() []string{
  return []string{"repo_url",}
}


func (s GitService) Run(t structures.Task, n string, ctx *structures.Context, infra_outputs *map[string]map[string]string, l*logging.Config) error {
  flume_info := ctx.GetEventValues("flume_info")
  flume_folder := flume_info["path"]
  repo_folder := filepath.Join(flume_folder, n)
  runCtx := make(map[string]string, 2)
  defer ctx.SetEventValues(n, runCtx)
  runCtx["success"] = "false"
  
  
  exists := true
  _, err := os.Stat(repo_folder)
  if err != nil {
    if os.IsNotExist(err) {
      exists = false
    } else {
      return fmt.Errorf("Error checking git repo: %w", err)
    }
  }

  if exists {
    l.InfoLogger("Updating Git Repo")
    cmd := exec.Command("git", "pull")
    cmd.Dir = repo_folder
    if err := cmd.Run(); err != nil {
      return fmt.Errorf("Error pulling repo from existing folder: %w", err)
    }
  } else {
    l.InfoLogger(fmt.Sprintf("Cloning into %s", repo_folder))
    repo_url, err := t.StringParam("repo_url")
    if err != nil { return err } 
    cmd := exec.Command("git", "clone", repo_url, repo_folder)
    _, err = cmd.CombinedOutput()
    if err != nil {
      return fmt.Errorf("Error cloning repo: %w", err)
    } 
  } 
  runCtx["success"] = "true"
  runCtx["repo_folder"] = repo_folder
  return nil
}

func init() {
  structures.Registry["git"] = GitService{}
}


