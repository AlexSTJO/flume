package infra

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
  "github.com/AlexSTJO/flume/internal/logging"
	"github.com/AlexSTJO/flume/internal/structures"
  "github.com/AlexSTJO/flume/internal/utils"
  "errors"
  "path"
  "strings"
)


type State struct{
  Resources []Resource `json:"resources"`
}

type Resource struct {
  Type string `json:"type"`
  Name string `json:"name"`
  Instances []Instance `json:"instances"`
}

type Instance struct {
  Attributes Attributes `json:"attributes"`
}

type Attributes struct {
  Id string `json:"id"`
  Tags map[string]string `json:"tags"`
  State string `json:"instance_state"`
}

type tfOutputValue struct {
    Sensitive bool            `json:"sensitive"`
    Type      json.RawMessage `json:"type"`
    Value     string             `json:"value"`
}

type Terraform struct {}


func (t *Terraform) Name() string {
  return "terraform"
}

func (t *Terraform) Call(d structures.Deployment, l *logging.Config) (map[string]string, error) {
  l.InfoLogger(fmt.Sprintf("Pulling Terraform Remote Repo: %s", d.Repo))
  key, err := TerraformPull(d.Repo, l)
  if err != nil {
    return nil, fmt.Errorf("Error pulling terraform repo: %w", err)
  }

  switch d.Action {
  case "sync": {
    if err := TerraformInit(key); err !=nil {
      return nil, fmt.Errorf("Terraform Init Failed: %w", err)
    }
    l.InfoLogger("Terraform Initialization Succesful")

    changes, err := TerraformPlan(key, d.VarFile)
    if err != nil{
      return nil, fmt.Errorf("Terraform Plan Failed: %w", err)
    }

    if changes{
      l.InfoLogger("Changes can be made. Running Apply")
      err := TerraformApply(key, d.VarFile)
      if err != nil {
        l.ErrorLogger(fmt.Errorf("Error Applying Terraform Deployment"))
        return nil, err
      }
      l.SuccessLogger("Succesful Terraform Apply") 
    } else {
      l.InfoLogger("Terraform Modules Up To Date With Infrastructure")
    }
    
  } 
  default:
    return nil, fmt.Errorf("Unknown Action: %s", d.Action)
  }
  tf_outputs, err := TerraformOutputs(key)
  if err != nil{
    l.ErrorLogger(fmt.Errorf("Error Parsing Terraform Outputs"))
  }
  return tf_outputs, nil
}


func TerraformState(key string) (*State, error) {
  cmd := exec.Command("terraform", "state", "pull")
  dir := filepath.Join(".", "terraform", key)
  cmd.Dir = dir

  out, err := cmd.Output()
  if err != nil{
    return nil, err
  }
  return ParseState(out)
}

func TerraformPull(repo string, l *logging.Config) (string, error) {
  clean := strings.TrimSuffix(repo, ".git")
  name := path.Base(clean)

  repo_folder := filepath.Join(".", "terraform", name) 

  _ = os.MkdirAll(filepath.Join(".", "terraform"), 0755)

   
  keyDir := filepath.Join("terraform", ".keys", name)
  b, err := utils.KeyExists(keyDir)
  if err != nil {
    return "", fmt.Errorf("checking key existence: %w", err)
  }

  if !b {
    if err := utils.GenerateDeployKey(keyDir); err != nil{
      return "", fmt.Errorf("generating deploy key: %w", err)
    }
    s, err := utils.ReadKey(keyDir)
    if err != nil {
      return "", fmt.Errorf("error reading key: %w", err)
    }
    l.InfoLogger(fmt.Sprintf("Deploy key generated for '%s': %s",repo, s))
    l.InfoLogger("To add key -> Log into GitHub Repo -> Settings -> Deploy Keys -> Add Deploy Key")   
  }

  if b := utils.HasRepoAccess(repo, keyDir); !b {
    l.ErrorLogger(fmt.Errorf("deploy key for '%s' does not have repo access", repo))
    s, err := utils.ReadKey(keyDir)
    if err != nil {
      return "", fmt.Errorf("error reading key: %w", err)
    }
    l.InfoLogger(fmt.Sprintf("Key: %s", s))
    l.InfoLogger("To add key -> Log into GitHub Repo -> Settings -> Deploy Keys -> Add Deploy Key") 
    return "", fmt.Errorf("git error")
  }

  private_key_path := filepath.Join(keyDir, "id_ed25519")
  ssh_env := "GIT_SSH_COMMAND=ssh -i "+private_key_path+" -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new"

 
  exists := true
  _, err = os.Stat(repo_folder)
  if err != nil {
    if os.IsNotExist(err) {
      exists = false
    } else {
      return "", fmt.Errorf("Error checking git repo: %w", err)
    }
  }
  if exists { 
    cmd := exec.Command("git", "pull")
    cmd.Dir = repo_folder
    cmd.Env = append(os.Environ(), ssh_env)
    if err := cmd.Run(); err != nil {
      return "", fmt.Errorf("Error pulling terraform repo from existing folder: %w", err)
    }
  } else { 
    cmd := exec.Command("git", "clone", repo, repo_folder)
    cmd.Env = append(os.Environ(), ssh_env)
    _, err = cmd.CombinedOutput()
    if err != nil {
      return "", fmt.Errorf("Error cloning repo: %w", err)
    } 
  }
  
  return name, nil
}

func TerraformInit(key string) error {
  cmd := exec.Command("terraform", "init", "-input=false", "-no-color")
 
  cmd.Dir = filepath.Join(".", "terraform", key)
  cmd.Env = append(os.Environ(), "TF_IN_AUTOMATION=1")

  _, err := cmd.CombinedOutput()

  if err != nil {
    return fmt.Errorf("terraform_init failed: %w", err)
  }
  
  return nil
}


func TerraformPlan(key string, var_file string) (bool, error) {   
  cmd := exec.Command("terraform", "plan", "-detailed-exitcode", "-input=false", "-var-file="+var_file, "-no-color")
  cmd.Dir = filepath.Join(".", "terraform", key)
  cmd.Env = append(os.Environ(), "TF_IN_AUTOMATION=1")

  _, err := cmd.CombinedOutput()

  if err == nil {
    return false, nil
  }
  var exitErr *exec.ExitError
  if errors.As(err, &exitErr) {
    switch exitErr.ExitCode() {
      case 0:
        return false, nil
      case 2:
        return true, nil
      default:
        return false, fmt.Errorf("Terraform Plan failed with exit code %d: %w", exitErr.ExitCode(), err)
    }
  }
  return false, fmt.Errorf("running terraform plan: %w", err)
}

func ParseState(data []byte) (*State, error) {
  var s State
  if err := json.Unmarshal(data, &s); err != nil {
    return nil, err
  }
  return &s, nil
}

func TerraformApply(key string, var_file string) (error) {
  cmd := exec.Command("terraform", "apply", "-auto-approve", "-input=false", "-var-file="+var_file)
  cmd.Dir = filepath.Join(".", "terraform", key)
  
  _, err := cmd.CombinedOutput() // stdout + stderr together

	if err != nil {
		return fmt.Errorf(
      "terraform apply failed status: %w",err)
	}
  return nil
}

func TerraformOutputs(key string) (map[string]string, error) { 
  cmd := exec.Command("terraform", "output", "-json")
  cmd.Dir = filepath.Join(".", "terraform", key)

  out, err := cmd.CombinedOutput()
  if err != nil {
    return nil, fmt.Errorf("decode terraform outputs: %w", err)
  }
  var parsed map[string]tfOutputValue
  if err := json.Unmarshal(out, &parsed); err != nil {
    return nil, fmt.Errorf("Unparsable terraform output: %w", err)
  }

  outputs := make(map[string]string, len(parsed))
  for name, v := range parsed {
    outputs[name] = v.Value
  }

  return outputs, nil
}


func init() {
  registry["terraform"] = &Terraform{}
}
