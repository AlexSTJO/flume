package services

import (
	"fmt"
  "time"
  "context"

	"github.com/AlexSTJO/flume/internal/logging"
	"github.com/AlexSTJO/flume/internal/resolver"
	"github.com/AlexSTJO/flume/internal/structures"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
  "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
)

type CloudfrontInvalidateService struct {
  client *cloudfront.Client
}

func NewCloudfrontInvalidateService() (*CloudfrontInvalidateService, error) {
  cfg, err := config.LoadDefaultConfig(context.Background())
  if err != nil {
      return nil, fmt.Errorf("loading aws config: %w", err)
  }

  return &CloudfrontInvalidateService{
    client: cloudfront.NewFromConfig(cfg),
  }, nil
}

func (s CloudfrontInvalidateService) Name() string {
  return "cloudfront_invalidate"
}

func (s CloudfrontInvalidateService) Parameters() []string {
  return []string{"dist_id", "paths"}
}

func (s CloudfrontInvalidateService) Run(t structures.Task, n string, ctx *structures.Context, infra_outputs *map[string]map[string]string,l *logging.Config) error { 
  runCtx := make(map[string]string, 1)
  defer ctx.SetEventValues(n, runCtx)
  runCtx["success"] = "false"
  raw_dist_id, err := resolver.ResolveAny(t.Parameters["dist_id"], ctx, infra_outputs)
  if err != nil { return err }

  dist_id, ok := raw_dist_id.(string)
  if !ok {return fmt.Errorf("Parameter 'dist_id' has to be a string")}


  raw_paths, err := resolver.ResolveAny(t.Parameters["paths"], ctx, infra_outputs)
  if err != nil {return err}

  rawList, ok := raw_paths.([]any)
  if !ok {
      return fmt.Errorf("Parameter 'paths' has to be a list")
  }

  paths := make([]string, len(rawList))
  for i, v := range rawList {
      s, ok := v.(string)
      if !ok {
          return fmt.Errorf("all paths must be strings, got %T", v)
      }
      paths[i] = s
  }
  awsCtx := context.Background()

  callerRef := fmt.Sprintf("flume-%d", time.Now().UnixNano())
  _, err = s.client.CreateInvalidation(awsCtx, &cloudfront.CreateInvalidationInput{
		DistributionId: aws.String(dist_id),
		InvalidationBatch: &types.InvalidationBatch{
			CallerReference: aws.String(callerRef),
			Paths: &types.Paths{
				Quantity: aws.Int32(int32(len(paths))),
				Items:    paths,
			},
		},
	})

  runCtx["success"] = "true"
  return nil
}


func init() {
  cf_i, err := NewCloudfrontInvalidateService()
  if err != nil { fmt.Println("Error Registering Cloudfront Service")}
  structures.Registry["cloudfront_invalidate"] = cf_i
}
