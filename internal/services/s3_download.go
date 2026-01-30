package services

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlexSTJO/flume/internal/logging"
	"github.com/AlexSTJO/flume/internal/resolver"
	"github.com/AlexSTJO/flume/internal/structures"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3DownloadService struct {
	client *s3.Client
}

func (s S3DownloadService) Name() string {
	return "s3_download"
}

func (s S3DownloadService) Parameters() []string {
	return []string{"bucket", "destination"}
}

func NewS3DownloadService() (*S3DownloadService, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("loading aws config: %w", err)
	}
	return &S3DownloadService{
		client: s3.NewFromConfig(cfg),
	}, nil
}

func (s S3DownloadService) Run(t structures.Task, n string, ctx *structures.Context, infra_outputs *map[string]map[string]string, l *logging.Config, r *structures.RunInfo) error {
	runCtx := make(map[string]string)
	defer ctx.SetEventValues(n, runCtx)
	runCtx["success"] = "false"

	rawBucket, err := t.StringParam("bucket")
	if err != nil {
		return err
	}
	bucket, err := resolver.ResolveStringParam(rawBucket, ctx, infra_outputs, r)
	if err != nil {
		return err
	}

	rawDestination, err := t.StringParam("destination")
	if err != nil {
		return err
	}
	destination, err := resolver.ResolveStringParam(rawDestination, ctx, infra_outputs, r)
	if err != nil {
		return err
	}

	var key, prefix string
	if rawKey, err := t.StringParam("key"); err == nil {
		key, _ = resolver.ResolveStringParam(rawKey, ctx, infra_outputs, r)
	}
	if rawPrefix, err := t.StringParam("prefix"); err == nil {
		prefix, _ = resolver.ResolveStringParam(rawPrefix, ctx, infra_outputs, r)
	}

	if key == "" && prefix == "" {
		return fmt.Errorf("s3_download: must provide either 'key' or 'prefix' parameter")
	}

	awsCtx := context.Background()

	if key != "" {
		l.InfoLogger(fmt.Sprintf("Downloading s3://%s/%s to %s", bucket, key, destination))

		if err := s.downloadFile(awsCtx, bucket, key, destination); err != nil {
			return err
		}

		runCtx["success"] = "true"
		runCtx["downloaded_files"] = "1"
		return nil
	}

	l.InfoLogger(fmt.Sprintf("Downloading files from s3://%s/%s to %s", bucket, prefix, destination))

	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})

	downloadCount := 0
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(awsCtx)
		if err != nil {
			return fmt.Errorf("listing objects: %w", err)
		}

		for _, obj := range page.Contents {
			objKey := aws.ToString(obj.Key)

			if strings.HasSuffix(objKey, "/") {
				continue
			}

			relPath := strings.TrimPrefix(objKey, prefix)
			relPath = strings.TrimPrefix(relPath, "/")

			destPath := filepath.Join(destination, relPath)

			if err := s.downloadFile(awsCtx, bucket, objKey, destPath); err != nil {
				return err
			}
			downloadCount++
		}
	}

	l.InfoLogger(fmt.Sprintf("Downloaded %d files", downloadCount))
	runCtx["success"] = "true"
	runCtx["downloaded_files"] = fmt.Sprintf("%d", downloadCount)
	return nil
}

func (s S3DownloadService) downloadFile(ctx context.Context, bucket, key, destPath string) error {
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("creating directory %s: %w", destDir, err)
	}

	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("getting s3://%s/%s: %w", bucket, key, err)
	}
	defer resp.Body.Close()

	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("creating file %s: %w", destPath, err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("writing to %s: %w", destPath, err)
	}

	return nil
}

func init() {
	s3Download, err := NewS3DownloadService()
	if err != nil {
		fmt.Println("Error Registering S3 Download Service")
	}
	structures.Registry["s3_download"] = s3Download
}
