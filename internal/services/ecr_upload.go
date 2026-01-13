package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"os/exec"
	"strings"

	"github.com/AlexSTJO/flume/internal/logging"
	"github.com/AlexSTJO/flume/internal/resolver"
	"github.com/AlexSTJO/flume/internal/structures"
	"github.com/AlexSTJO/flume/internal/utils"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
)

type EcrUploadService struct{
  client *ecr.Client
  region string
  account_id string
}

func (s EcrUploadService) Name() string{
  return "ecr_upload"
}

func (s EcrUploadService) Parameters() []string{
  return []string{"local_image", "registry", "tag"}
}

func NewEcrUploadService() (*EcrUploadService, error) {
  cfg, err := config.LoadDefaultConfig(context.Background())
  ctx := context.Background()
  if err != nil {
    return nil, fmt.Errorf("loading aws config: %w", err)
  }

  account_id, region, err := utils.AccountAndRegion(ctx, cfg)
  if err != nil {
    return nil, fmt.Errorf("getting account details: %w", err)
  }
  return &EcrUploadService{
    client: ecr.NewFromConfig(cfg),
    account_id: account_id,
    region: region,
  }, nil
}

func (s EcrUploadService) Run(t structures.Task, n string, ctx *structures.Context, infra_outputs *map[string]map[string]string, l*logging.Config, r *structures.RunInfo) error {
  runCtx := make(map[string]string, 2)
  defer ctx.SetEventValues(n, runCtx)
  runCtx["success"] = "false"
  aws_context := context.Background()

  raw_local_image, err:= t.StringParam("local_image")
  if err != nil { return err }
  local_image, err := resolver.ResolveStringParam(raw_local_image, ctx, infra_outputs)
  if err != nil { return err }

  raw_registry, err := t.StringParam("registry")
  if err != nil { return err }
  registry, err := resolver.ResolveStringParam(raw_registry, ctx, infra_outputs)
  if err != nil { return err }

  raw_tag, err := t.StringParam("tag")
  if err != nil { return err }
  tag, err := resolver.ResolveStringParam(raw_tag, ctx, infra_outputs)
  if err != nil { return err }


  auth, err := s.client.GetAuthorizationToken(aws_context, &ecr.GetAuthorizationTokenInput{})
  if err != nil {
    return err
  } 

  token, _ := base64.StdEncoding.DecodeString(
		*auth.AuthorizationData[0].AuthorizationToken,
	)
	parts := strings.SplitN(string(token), ":", 2)

  username := parts[0]
  password := parts[1]

  l.InfoLogger(fmt.Sprintf("Logging into registry: %s", registry))
  

  login := exec.Command(
		"docker", "login",
		"--username", username,
		"--password-stdin",
		registry,
	)

  login.Stdin = strings.NewReader(password)
  login.Stdout = nil
	login.Stderr = nil

  if err := login.Run(); err != nil {
		return fmt.Errorf("docker login failed: %w", err)
	}

  remote_image := fmt.Sprintf("%s:%s", registry, tag)
  l.InfoLogger(fmt.Sprintf("Tagging Image: %s", remote_image))

  if err := exec.Command("docker", "tag", local_image, remote_image).Run(); err != nil {
    err = fmt.Errorf("docker tag failed: %w", err)
    l.ErrorLogger(err)
    return err
	}

  cmd := exec.Command("docker", "push", remote_image)
  _, err = cmd.CombinedOutput()
  if err != nil {
    err = fmt.Errorf("docker push failed: %w", err)
    l.ErrorLogger(err)
    return err
	}

  l.InfoLogger(fmt.Sprintf("Image succesfully pushed to: %s", remote_image))
  
  runCtx["remote_image"] = remote_image
  runCtx["success"] = "true"
  return nil

}



func init() {
  ecr_s, err := NewEcrUploadService()
  if err != nil {
    fmt.Printf("Error registering 'ecr_upload' service: %w\n", err)
  }
  structures.Registry["ecr_upload"] = ecr_s
}
