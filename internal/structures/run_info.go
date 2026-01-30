package structures

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"
)

func newID() string {
	ts := time.Now().UTC().Format("20060102T150405Z")

	b := make([]byte, 4)
	_, _ = rand.Read(b)

	return fmt.Sprintf("%s_%s", ts, hex.EncodeToString(b))
}

type RunInfo struct {
	RunID    string
	RunDir   string
	Pipeline string
	Remote   bool
	FileRef  string
	S3       *RemotePipeline
	Params   map[string]string
}

type RemotePipeline struct {
	Bucket string
	Key    string
}

func GenerateRunInfo(fileRef string, params map[string]string) (*RunInfo, error) {
	remote := false
	pipeline := ""
	var remote_pipeline RemotePipeline
	if strings.HasPrefix(fileRef, "s3://") {
		remote = true
		path := strings.TrimPrefix(fileRef, "s3://")

		parts := strings.SplitN(path, "/", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("Invalid s3 uri: %s", fileRef)
		}

		remote_pipeline.Bucket = parts[0]
		remote_pipeline.Key = parts[1]
		segments := strings.Split(strings.TrimSuffix(remote_pipeline.Key, "/"), "/")
		if len(segments) != 3 {
			return nil, fmt.Errorf("Invalid object key in S3 Uri: %s", remote_pipeline.Key)
		}

		pipeline = segments[1]
	} else {
		fmt.Println("hey")
		raw_fileRef := strings.TrimSuffix(fileRef, ".yaml")
		pipeline = strings.TrimPrefix(raw_fileRef, "local://")
	}
	run_id := newID()
	run_dir, err := os.MkdirTemp("", fmt.Sprintf("flume-run-%s", run_id))
	if err != nil {
		return nil, fmt.Errorf("Couldnt make tmp run directory: %w", err)
	}
	if params == nil {
		params = make(map[string]string)
	}
	return &RunInfo{
		RunID:    run_id,
		RunDir:   run_dir,
		Pipeline: pipeline,
		Remote:   remote,
		FileRef:  fileRef,
		S3:       &remote_pipeline,
		Params:   params,
	}, nil
}
