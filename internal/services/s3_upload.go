package services

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
  "os"
  "mime"

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

func (s S3UploadService) Run(t structures.Task, n string, ctx *structures.Context, infra_outputs *map[string]map[string]string, l*logging.Config, r *structures.RunInfo) error {
  runCtx := make(map[string]string, 1)
  defer ctx.SetEventValues(n, runCtx)
  runCtx["success"] = "false"
  

  raw_bucket, err := t.StringParam("bucket")
  if err != nil {return err}
  bucket, err := resolver.ResolveStringParam(raw_bucket, ctx , infra_outputs ) 
  if err != nil { return err }
  
  raw_source, err := t.StringParam("source")
  if err != nil {return err}
  source, err := resolver.ResolveStringParam(raw_source, ctx, infra_outputs )
  if err != nil { return err }
  
  raw_prefix, err := t.StringParam("prefix")
  if err != nil { return err}
  prefix, err := resolver.ResolveStringParam(raw_prefix, ctx, infra_outputs)
  if err != nil { return err }

  l.InfoLogger(fmt.Sprintf("Uploading contents of '%s' to bucket: '%s' with prefix of '%s'", source, bucket, prefix))

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
    ext := strings.ToLower(filepath.Ext(key))
    contentType := mime.TypeByExtension(ext)
    if contentType == "" {
        contentType = "application/octet-stream"
    }
    _, err = s.client.PutObject(awsCtx, &s3.PutObjectInput{
      Bucket: aws.String(bucket),
      Key:    aws.String(key),
      Body:   f,
      ContentType: aws.String(contentType),
    }) 
    if err != nil {
      return fmt.Errorf("put %s: %w", key, err)
    }

    return nil
  })
  if err != nil {
    return err
  }

  runCtx["success"] = "true"
  return nil
}

func init() {
  s3_upload, err := NewS3SyncService()
  if err != nil {
    fmt.Println("Error Registering S3 Service")
  }
  structures.Registry["s3_upload"] = s3_upload
}
