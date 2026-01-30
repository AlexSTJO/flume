# Flume

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A declarative workflow engine for hybrid cloud automation.

Flume is a lightweight, extensible workflow orchestrator designed to automate builds, deployments, and infrastructure using a simple YAML format. It integrates with Terraform, AWS services, Git, and custom Go services, enabling hybrid workflows that unify application deployment and infrastructure provisioning.

## Table of Contents

- [Quick Start](#quick-start)
- [Features](#features)
- [Configuration](#configuration)
- [Pipeline Structure](#pipeline-structure)
- [Services](#services)
- [Resolver Patterns](#resolver-patterns)
- [Examples](#examples)
- [Creating Custom Services](#creating-custom-services)
- [Roadmap](#roadmap)

## Quick Start

```bash
# Clone and run
git clone https://github.com/AlexSTJO/flume.git
cd flume

# Create .env file
echo -e "URL=localhost\nPORT=8080" > .env

# Start the server
go run .

# Trigger a pipeline (in another terminal)
curl -X POST "http://localhost:8080/run" \
  -H "Content-Type: application/json" \
  -d '{"pipeline_ref": "sample-flume"}'

# Trigger with runtime parameters
curl -X POST "http://localhost:8080/run" \
  -H "Content-Type: application/json" \
  -d '{
    "pipeline_ref": "sample-flume",
    "parameters": {"environment": "staging", "version": "1.2.3"}
  }'
```

## Features

**Core**
- Declarative YAML pipelines with automatic DAG execution
- Parallel task workers with dependency resolution
- Dynamic resolver engine for variable substitution
- Runtime pipeline parameters via API
- Conditional task execution (`run_if`, `skip_if`)
- Task retry with configurable attempts and delay
- Task timeout support
- API-triggered and cron-scheduled pipelines
- Remote pipelines from S3 (`s3://<bucket>/<key>`)
- Hash-based file change detection

**Infrastructure**
- Terraform integration for infrastructure provisioning

**Services**
- Git clone with GitHub App authentication
- Shell command execution
- Docker image building
- HTTP requests (GET, POST, PUT, DELETE, PATCH)
- AWS: S3 upload/download, CloudFront invalidation, ECR push, SSM operations
- SMTP email notifications
- JSON file writing
- Modular service registry for custom extensions

## Configuration

### Prerequisites

- Go 1.22+
- Terraform (if using infrastructure provisioning)
- AWS credentials configured (if using AWS services)

### Environment

Create a `.env` file in the project root:

```env
URL=localhost
PORT=8080
```

### Build Options

```bash
# Run directly
go run .

# Build binary
go build -o flume .
./flume
```

## Pipeline Structure

Pipelines are stored in `.flume/<pipeline-name>/<pipeline-name>.yaml`:

```
.flume/
├── my-pipeline/
│   └── my-pipeline.yaml
└── another-pipeline/
    └── another-pipeline.yaml
```

### Schema

```yaml
name: "pipeline-name"
trigger:
  type: "api"           # or "cron"
  cron_expression: ""   # e.g., "0 0 * * *" (for cron triggers)
log_path: ""

infrastructure:
  deployment_name:
    service: terraform
    action: sync
    repo: "git@github.com:user/terraform-repo.git"
    var-file: "terraform.tfvars"

tasks:
  task_name:
    service: service_name
    dependencies: ["other_task"]
    run_if: "${param:env} == production"   # optional: run only if condition is true
    skip_if: "${context:prev.skip} == true" # optional: skip if condition is true
    timeout: "5m"                           # optional: task timeout (e.g., 30s, 5m, 1h)
    retry:                                  # optional: retry configuration
      max_attempts: 3
      delay: "10s"
    parameters:
      key: value
```

### Triggering Pipelines

```bash
# Local pipeline
curl -X POST "http://localhost:8080/run" \
  -H "Content-Type: application/json" \
  -d '{"pipeline_ref": "my-pipeline"}'

# Remote pipeline from S3
curl -X POST "http://localhost:8080/run" \
  -H "Content-Type: application/json" \
  -d '{"pipeline_ref": "s3://my-bucket/pipelines/my-pipeline"}'

# With runtime parameters
curl -X POST "http://localhost:8080/run" \
  -H "Content-Type: application/json" \
  -d '{
    "pipeline_ref": "my-pipeline",
    "parameters": {
      "environment": "production",
      "version": "2.0.0",
      "debug": "false"
    }
  }'
```

Parameters are accessible in your pipeline using `${param:name}` syntax.

## Services

| Service | Description | Required Parameters |
|---------|-------------|---------------------|
| `git` | Clone repositories (GitHub App auth) | `repo_url` |
| `shell` | Execute shell commands | `command` |
| `docker_build` | Build Docker images | `build_path`, `image_name`, `tag` |
| `http_request` | Make HTTP requests | `url`, `method`, `body`, `headers` |
| `s3_upload` | Upload files to S3 | `bucket`, `source`, `prefix` |
| `s3_download` | Download files from S3 | `bucket`, `destination`, `key` or `prefix` |
| `ecr_upload` | Push images to ECR | `local_image`, `registry`, `tag` |
| `cloudfront_invalidate` | Invalidate CloudFront cache | `dist_id`, `paths` |
| `ssm` | AWS SSM operations | `instance_id`, `commands` |
| `smtp` | Send emails | (see service file) |
| `json_writer` | Write JSON to file | (see service file) |

## Resolver Patterns

Use these patterns in task parameters to reference dynamic values:

| Pattern | Description | Example |
|---------|-------------|---------|
| `${context:<task>.<key>}` | Output from previous task | `${context:git_pull.repo_folder}` |
| `${infra:terraform.<output>}` | Terraform output value | `${infra:terraform.bucket_name}` |
| `${env:<VAR>}` | Environment variable | `${env:AWS_REGION}` |
| `${param:<name>}` | Runtime parameter from API | `${param:environment}` |
| `${timestamp}` | Execution timestamp | `${timestamp}` |

## Examples

### Website Deployment

```yaml
name: "portfolio-website"
trigger:
  type: "api"
log_path: ""

infrastructure:
  tfdeploy:
    service: terraform
    action: sync
    repo: "git@github.com:AlexSTJO/portfolio-website-architecture.git"
    var-file: "terraform.tfvars"

tasks:
  git_pull:
    service: git
    dependencies: []
    parameters:
      repo_url: "git@github.com:AlexSTJO/portfolio-website"

  build:
    service: shell
    dependencies: ["git_pull"]
    parameters:
      command: |
        cd ${context:git_pull.repo_folder}
        npm install
        npm run build

  upload:
    service: s3_upload
    dependencies: ["build"]
    parameters:
      bucket: ${infra:terraform.site_bucket_name}
      source: ${context:git_pull.repo_folder}/out
      prefix: ""

  cf_invalidate:
    service: cloudfront_invalidate
    dependencies: ["upload"]
    parameters:
      dist_id: ${infra:terraform.dist_id}
      paths: ["/*"]
```

### Docker Build & ECR Deploy

```yaml
name: "flume-deploy"
trigger:
  type: "api"
log_path: ""

infrastructure:
  tfdeploy:
    service: terraform
    action: sync
    repo: "git@github.com:AlexSTJO/flume-architecture.git"
    var-file: "terraform.tfvars"

tasks:
  git_pull:
    service: git
    dependencies: []
    parameters:
      repo_url: "git@github.com:AlexSTJO/flume.git"

  docker_build:
    service: docker_build
    dependencies: ["git_pull"]
    parameters:
      build_path: ${context:git_pull.repo_folder}
      image_name: "flume"
      tag: "latest"

  ecr_upload:
    service: ecr_upload
    dependencies: ["docker_build"]
    parameters:
      local_image: ${context:docker_build.image}
      registry: ${infra:terraform.ecr_repository_url}
      tag: "latest"

  deploy:
    service: ssm
    dependencies: ["ecr_upload"]
    parameters:
      instance_id: ${infra:terraform.flume_instance_id}
      commands:
        - |
          aws ecr get-login-password --region "us-east-2" \
            | docker login --username AWS --password-stdin \
              "${infra:terraform.ecr_repository_url}"
          docker pull "${context:ecr_upload.remote_image}"
          docker rm -f flume || true
          docker run -d --name flume -p 8080:8080 ${context:ecr_upload.remote_image}
```

### HTTP Webhook Notification

```yaml
tasks:
  notify:
    service: http_request
    dependencies: ["deploy"]
    parameters:
      url: "https://hooks.slack.com/services/xxx"
      method: "POST"
      body: '{"text": "Deployment complete"}'
      headers:
        Content-Type: "application/json"
```

### Parameterized Deployment

Pass runtime parameters to customize pipeline behavior:

```yaml
name: "parameterized-deploy"
trigger:
  type: "api"

tasks:
  deploy:
    service: shell
    parameters:
      command: |
        echo "Deploying version ${param:version} to ${param:environment}"
        ./deploy.sh --env ${param:environment} --version ${param:version}

  notify:
    service: http_request
    dependencies: ["deploy"]
    run_if: "${param:environment} == production"
    parameters:
      url: "https://hooks.slack.com/services/xxx"
      method: "POST"
      body: '{"text": "Deployed ${param:version} to production"}'
      headers:
        Content-Type: "application/json"
```

Trigger with:
```bash
curl -X POST "http://localhost:8080/run" \
  -H "Content-Type: application/json" \
  -d '{
    "pipeline_ref": "parameterized-deploy",
    "parameters": {"environment": "production", "version": "1.5.0"}
  }'
```

## Creating Custom Services

1. Create a file in `internal/services/`
2. Implement the `Service` interface
3. Register via `init()`

```go
package services

import (
    "github.com/AlexSTJO/flume/internal/logging"
    "github.com/AlexSTJO/flume/internal/structures"
)

type MyService struct{}

func (s MyService) Name() string { return "my_service" }

func (s MyService) Parameters() []string { return []string{"my_param"} }

func (s MyService) Run(t structures.Task, n string, ctx *structures.Context, infra_outputs *map[string]map[string]string, l *logging.Config, r *structures.RunInfo) error {
    runCtx := make(map[string]string)
    defer ctx.SetEventValues(n, runCtx)

    myParam, err := t.StringParam("my_param")
    if err != nil {
        return err
    }

    // Do work here...

    runCtx["result"] = myParam
    return nil
}

func init() {
    structures.Registry["my_service"] = MyService{}
}
```

## Roadmap

- [ ] Web UI
- [ ] Remote workers
- [x] Retry strategies
- [x] Task timeouts
- [x] Conditional execution
- [x] Pipeline parameters
- [ ] Secrets management
- [ ] Plugin system

## License

MIT
