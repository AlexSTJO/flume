package utils

import (
  "os"
  "os/exec"
  "fmt"
  "path/filepath"
)

func KeyExists(keyDir string) (bool, error) {
  privPath := filepath.Join(keyDir, "id_ed25519")
  pubPath := privPath + ".pub"

  if _, err := os.Stat(pubPath); err != nil {
    if os.IsNotExist(err) {
      return false, nil
    }
    return false, err
  }

  return true, nil 
}

func GenerateDeployKey(keyDir string) error {
  if err := os.MkdirAll(keyDir, 0o700); err != nil {
    return fmt.Errorf("Creating key dir: %w", err)
  }

  keyPath := filepath.Join(keyDir, "id_ed25519")
  if _, err := os.Stat(keyPath); err == nil {
    return fmt.Errorf("deploy key already exists")
  }

  cmd := exec.Command("ssh-keygen", "-t", "ed25519", "-f", keyPath)
  return cmd.Run()
}

func HasRepoAccess(repoUrl, keyDir string) bool {
  privateKeyPath := filepath.Join(keyDir, "id_ed25519")
  cmd := exec.Command("git", "ls-remote", repoUrl)
  cmd.Env = append(os.Environ(),
		"GIT_SSH_COMMAND=ssh -i "+privateKeyPath+" -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new",
	)
  err := cmd.Run()
  return err == nil
}

func ReadKey(keyDir string) (string, error) {
  pubKey := filepath.Join(keyDir, "id_ed25519.pub")

  bytes, err := os.ReadFile(pubKey)
  if err != nil {
    return "",err
  }
  return string(bytes), nil
}
