package services

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
  "os"
  "io"

	"github.com/AlexSTJO/flume/internal/logging"
	"github.com/AlexSTJO/flume/internal/resolver"
	"github.com/AlexSTJO/flume/internal/structures"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3UploadService struct {
  client *s3.Client
}


func (s S3UploadService) Name() string {
  return "s3_upload"
}

func (s S3UploadService) Parameters() []string {
  return []string{"bucket", "source", "prefix"} 
}

func NewS3SyncService() (*S3UploadService, error) {
    cfg, err := config.LoadDefaultConfig(context.Background())
    if err != nil {
        return nil, fmt.Errorf("loading aws config: %w", err)
    }
    return &S3UploadService{
        client: s3.NewFromConfig(cfg),
    }, nil
}

func (s S3UploadService) Run(t structures.Task, n string, ctx *structures.Context, infra_outputs *map[string]map[string]string, l*logging.Config) error {
  bucket, err := resolver.ResolveParam("bucket", t.Parameters, ctx , infra_outputs ) 
  if err != nil { return err }

  source, err := resolver.ResolveParam("source", t.Parameters, ctx, infra_outputs )
  if err != nil { return err }

  prefix, err := resolver.ResolveParam("prefix", t.Parameters, ctx, infra_outputs)
  if err != nil { return err }



  awsCtx := context.Background()

  err = filepath.WalkDir(source, func(path string, d os.DirEntry, walkErr error) error {
    if walkErr != nil{
      return walkErr
    }

    if d.IsDir() {
      return err
    }

    rel, err := filepath.Rel(source, path)
    if err != nil {
        return err
    }

    key := filepath.ToSlash(rel)
    if prefix != "" {
      key = strings.TrimSuffix(prefix, "/") + "/" + key
    }

    f, err := os.Open(path)
    if err != nil {
        return fmt.Errorf("open %s: %w", path, err)
    }
    defer f.Close()
    _, err = s.client.PutObject(awsCtx, &s3.PutObjectInput{
      Bucket: aws.String(bucket),
      Key:    aws.String(key),
      Body:   f,
    }) 
    if err != nil {
      return fmt.Errorf("put %s: %w", key, err)
    }

    l.InfoLogger(fmt.Sprintf("Uploaded %s -> s3://%s/%s", path, bucket, key))
    return nil
  })
  if err != nil {
    return err
  }


  return nil
}

func init() {
  structures.Registry["s3_upload"] = S3UploadService{}
}
