package services

import (
	"context"
	"fmt"
  "strings"
  "time"

  "github.com/AlexSTJO/flume/internal/logging"
	"github.com/AlexSTJO/flume/internal/resolver"
	"github.com/AlexSTJO/flume/internal/structures"
  "github.com/aws/aws-sdk-go-v2/service/ssm/types"
  "github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/config"
  "github.com/aws/aws-sdk-go-v2/aws"
)

type SSMService struct {
  client *ssm.Client
}

func (s SSMService) Name() string{
  return "ssm"
}

func (s SSMService) Parameters() []string{
  return []string{"instance_id", "commands"}
}

func NewSSMService() (*SSMService, error) {
  cfg, err := config.LoadDefaultConfig(context.Background())
  if err != nil{
    return nil, fmt.Errorf("laoding aws config: %w", err)
  }
  return &SSMService{
    client: ssm.NewFromConfig(cfg),
  }, err
}


func (s SSMService) Run(t structures.Task, n string, ctx *structures.Context, infra_outputs *map[string]map[string]string, l*logging.Config, r *structures.RunInfo) error {
  runCtx := make(map[string]string, 2)
  defer ctx.SetEventValues(n, runCtx)
  runCtx["success"] = "false"

  aws_ctx := context.Background()

  raw_instance, err := t.StringParam("instance_id")
  if err != nil { return err }
  instance_id, err := resolver.ResolveStringParam(raw_instance, ctx, infra_outputs)
  if err != nil { return err}
  
  raw_commands := t.Parameters["commands"]
  res_commands, err := resolver.ResolveAny(raw_commands, ctx, infra_outputs)
  if err != nil { return err }
  commands, err := resolver.ToStringSlice(res_commands)
  if err != nil {
    return err
  }

  l.InfoLogger(fmt.Sprintf("Sending commands: '%d' to instance: '%s'", len(commands), instance_id))

  send_out, err := s.client.SendCommand(aws_ctx, &ssm.SendCommandInput{
		DocumentName:  aws.String("AWS-RunShellScript"),
		InstanceIds:   []string{instance_id},
		TimeoutSeconds: aws.Int32(600),
		Parameters: map[string][]string{
			"commands": commands,
		},
		CloudWatchOutputConfig: &types.CloudWatchOutputConfig{
			CloudWatchOutputEnabled: *aws.Bool(true),
		},
	})

  if err != nil {
		return fmt.Errorf("send command: %w", err)
	}
	cmd_id := *send_out.Command.CommandId

	deadline := time.Now().Add(8 * time.Minute)
  var lastStatus types.CommandInvocationStatus
	lastStdoutLen := 0
	lastStderrLen := 0

	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for command %s", cmd_id)
		}

		inv, err := s.client.GetCommandInvocation(aws_ctx, &ssm.GetCommandInvocationInput{
			CommandId:  aws.String(cmd_id),
			InstanceId: aws.String(instance_id),
		})

		if err != nil {
			if strings.Contains(err.Error(), "InvocationDoesNotExist") {
				time.Sleep(2 * time.Second)
				continue
			}
			return fmt.Errorf("get command invocation: %w", err)
		}

    if l != nil && inv.Status != lastStatus {
			l.InfoLogger(fmt.Sprintf("SSM %s status: %s", cmd_id, inv.Status))
			lastStatus = inv.Status
		}

		stdout := deref(inv.StandardOutputContent)
		stderr := deref(inv.StandardErrorContent)

		if l != nil {
			if len(stdout) > lastStdoutLen {
				lastStdoutLen = len(stdout)
			}
			if len(stderr) > lastStderrLen {
				lastStderrLen = len(stderr)
			}
		}	
		switch inv.Status {
		case types.CommandInvocationStatusPending,
			types.CommandInvocationStatusInProgress,
			types.CommandInvocationStatusDelayed:
			time.Sleep(2 * time.Second)
			continue
		default:	
			if inv.Status != types.CommandInvocationStatusSuccess {
				return fmt.Errorf("ssm command failed: %s", inv.Status)
			}
      runCtx["success"] = "true"
			return nil
		}
  }
 }

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func init() {
  ssm, err := NewSSMService()
  if err != nil {
    fmt.Printf("Error registering ssm service: %w\n", err)
  }
  structures.Registry["ssm"] = ssm
}
