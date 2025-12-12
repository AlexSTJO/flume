package services

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlexSTJO/flume/internal/logging"
	"github.com/AlexSTJO/flume/internal/structures"
  "github.com/AlexSTJO/flume/internal/utils"
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
  repo_url, err := t.StringParam("repo_url")
  if err != nil { return err } 
  
  keyDir := filepath.Join(flume_folder, ".keys", n)
  b, err := utils.KeyExists(keyDir)
  if err != nil {
    return fmt.Errorf("checking key existence: %w", err)
  }

  if !b {
    if err := utils.GenerateDeployKey(keyDir); err != nil{
      return fmt.Errorf("generating deploy key: %w", err)
    }
    s, err := utils.ReadKey(keyDir)
    if err != nil {
      return fmt.Errorf("error reading key: %w", err)
    }
    l.InfoLogger(fmt.Sprintf("Deploy key generated for '%s': %s",repo_url, s))
    l.InfoLogger("To add key -> Log into GitHub Repo -> Settings -> Deploy Keys -> Add Deploy Key")   
  }



  if b := utils.HasRepoAccess(repo_url, keyDir); !b {
    l.ErrorLogger(fmt.Errorf("Deploy key for '%s' does not have repo access", repo_url))
    s, err := utils.ReadKey(keyDir)
    if err != nil {
      return fmt.Errorf("error reading key: %w", err)
    }
    l.InfoLogger(fmt.Sprintf("Key is: %s", s))
    l.InfoLogger("To add key -> Log into GitHub Repo -> Settings -> Deploy Keys -> Add Deploy Key") 
    return fmt.Errorf("git error")
  }
  
  private_key_path := filepath.Join(keyDir, "id_ed25519")
  ssh_env := "GIT_SSH_COMMAND=ssh -i "+private_key_path+" -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new"

  exists := true
  _, err = os.Stat(repo_folder)
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
    cmd.Env = append(os.Environ(), ssh_env)
    cmd.Dir = repo_folder
    if err := cmd.Run(); err != nil {
      return fmt.Errorf("Error pulling repo from existing folder: %w", err)
    }
  } else {
    l.InfoLogger(fmt.Sprintf("Cloning into %s", repo_folder))
    
    cmd := exec.Command("git", "clone", repo_url, repo_folder)
    cmd.Env = append(os.Environ(), ssh_env)
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





