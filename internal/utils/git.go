package utils

import (
  "fmt"
  "strings"
)

func ParseGitHubRepo(repo string) (owner string, name string, err error) {
  repo = strings.TrimSpace(repo)

  if strings.HasPrefix(repo, "git@github.com:") {
      repo = strings.TrimPrefix(repo, "git@github.com:")
  }

  if strings.HasPrefix(repo, "https://github.com/") {
      repo = strings.TrimPrefix(repo, "https://github.com/")
  }

  repo = strings.TrimSuffix(repo, ".git")

  parts := strings.Split(repo, "/")
  if len(parts) != 2 {
      return "", "", fmt.Errorf("invalid GitHub repo format: %s", repo)
  }

  return parts[0], parts[1], nil
}

