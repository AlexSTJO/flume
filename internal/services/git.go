package services

import (
	"fmt"
	"path/filepath"
  "context"
	"os/exec"

	"github.com/AlexSTJO/flume/internal/github"
	"github.com/AlexSTJO/flume/internal/logging"
	"github.com/AlexSTJO/flume/internal/structures"
	"github.com/AlexSTJO/flume/internal/utils"
)

type GitService struct {}

func (s GitService) Name() string{
  return "git"
}


func (s GitService) Parameters() []string{
  return []string{"repo_url",}
}


func (s GitService) Run(t structures.Task, n string, ctx *structures.Context, infra_outputs *map[string]map[string]string, l*logging.Config , r *structures.RunInfo) error {
  runCtx := make(map[string]string, 2)
  defer ctx.SetEventValues(n, runCtx)
  runCtx["success"] = "false"
  repo_url, err := t.StringParam("repo_url")
  if err != nil { return err } 

  repo_folder := filepath.Join(r.RunDir, "job_outputs", n)
  owner, repo,err := utils.ParseGitHubRepo(repo_url)
  if err != nil {
    return err
  }

  token, err := githubapp.InstallationTokenForRepo(context.Background(), owner, repo)
  if err != nil {
    return err
  }

  app_repo_url := fmt.Sprintf("https://x-access-token:%s@github.com/AlexSTJO/%s", token, repo)

  
  


  l.InfoLogger(fmt.Sprintf("Cloning repo '%s", repo_url))
    
  cmd := exec.Command("git", "clone", app_repo_url, repo_folder)
    _, err = cmd.CombinedOutput()
  if err != nil {
    return fmt.Errorf("Error cloning repo: %w", err)
  }
  runCtx["success"] = "true"
  runCtx["repo_folder"] = repo_folder
  return nil
}

func init() {
  structures.Registry["git"] = GitService{}
}





