# **Flume**
A declarative workflow engine for hybrid cloud automation — powered by Go.

Flume is a lightweight, extensible workflow orchestrator designed to automate builds, deployments, and infrastructure using a simple YAML format. It integrates with Terraform, AWS services, Git, and custom Go services, enabling hybrid workflows that unify application deployment and infrastructure provisioning.

---

## **Features**

- Declarative YAML pipelines
- Automatic DAG execution with parallel task workers
- Dynamic resolver engine
- Terraform integration
- AWS services: S3 upload, CloudFront invalidation, ECR push, SSM operations
- Git clone with GitHub App authentication
- Docker image building
- Shell execution
- JSON writer
- SMTP email notifications
- Modular service registry
- API-triggered pipelines
- Cron scheduling
- Remote pipelines from S3 (`s3://<bucket>/<key>`)
- Hash-based file change detection

With more to come...

---

## Getting Started

### Prerequisites

- Go (1.22+ recommended)
- Terraform (if using the `terraform` service)
- AWS credentials configured (if using AWS services)
- `.env` file for server configuration

### Configuration

Flume reads its server configuration from a `.env` file in the project root.

Example:

```env
URL=localhost
PORT=8080
```

Flume will use these values to construct its server address as:

```text
http://$URL:$PORT
```

---

## Installation

Clone the repository and run Flume directly with Go:

```bash
git clone https://github.com/AlexSTJO/flume.git
cd flume
go run .
```

Flume will start the HTTP server using the URL and PORT defined in your `.env` file.

---

## Triggering a Pipeline via API

Pipelines are defined in YAML and referenced by name or S3 URI. To trigger a pipeline, send an HTTP POST request to the `/run` endpoint of the Flume server.

Example using `curl`:

```bash
# Local pipeline
curl -X POST "http://$URL:$PORT/run" \
  -H "Content-Type: application/json" \
  -d '{"pipeline_ref": "sample-flume"}'

# Remote pipeline from S3
curl -X POST "http://$URL:$PORT/run" \
  -H "Content-Type: application/json" \
  -d '{"pipeline_ref": "s3://my-bucket/pipelines/my-pipeline"}'
```

Where:

- `$URL` and `$PORT` come from your `.env` file
- `pipeline_ref` is either a local pipeline name (e.g., `sample-flume`) or an S3 URI

---

## **Architecture**

### Services

All task execution is handled through a unified Service interface:

```go
type Service interface {
    Name() string
    Parameters() []string
    Run(t Task, n string, ctx *Context, infra_outputs *map[string]map[string]string, l *logging.Config, r *RunInfo) error
}
```

Services register themselves in `init()`:

```go
func init() {
    structures.Registry["my_service"] = &MyService{}
}
```

### Available Services

| Service | Description |
|---------|-------------|
| `git` | Clone repositories (uses GitHub App auth) |
| `shell` | Execute shell commands |
| `s3_upload` | Upload files to S3 |
| `cloudfront_invalidate` | Invalidate CloudFront distribution |
| `json_writer` | Write JSON data to file |
| `docker_build` | Build Docker images |
| `ecr_upload` | Push images to ECR |
| `smtp` | Send emails |
| `ssm_service` | AWS SSM operations |

### Resolver Patterns

| Pattern | Meaning |
|--------|---------|
| `${context:task.value}` | Task output |
| `${infra:terraform.output}` | Terraform output |
| `${env:VAR}` | Environment variable |
| `${timestamp}` | Execution timestamp |

---

## **Pipeline Structure**

Pipelines are stored in `.flume/<pipeline-name>/<pipeline-name>.yaml`:

```
.flume/
├── my-pipeline/
│   └── my-pipeline.yaml
└── another-pipeline/
    └── another-pipeline.yaml
```

### Pipeline Schema

```yaml
name: "pipeline-name"
trigger:
  type: "api"           # or "cron" with cron_expression
  cron_expression: ""   # e.g., "0 0 * * *" (optional, for cron triggers)
log_path: ""

infrastructure:
  deployment_name:
    service: terraform
    action: sync
    repo: "git@github.com:user/terraform-repo.git"  # Git repo containing Terraform code
    var-file: "terraform.tfvars"

tasks:
  task_name:
    service: service_name
    dependencies: ["other_task"]  # Tasks to run before this one
    parameters:
      key: value
```

---

## **Example Pipeline**

A CI/CD pipeline that builds and deploys a website:

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

### Docker Build & Deploy Example

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

  ssm:
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

---

## **Custom Service Example**

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
    // Create output map for downstream tasks
    runCtx := make(map[string]string)
    defer ctx.SetEventValues(n, runCtx)

    // Get parameters from task
    myParam, err := t.StringParam("my_param")
    if err != nil {
        return err
    }

    // Do work here...

    // Set outputs for downstream tasks
    runCtx["success"] = "true"
    runCtx["result"] = myParam

    return nil
}

func init() {
    structures.Registry["my_service"] = MyService{}
}
```

---

## **Roadmap**

- Web UI  
- Remote workers  
- Retry strategies  
- Secrets management  
- Plugin system  

---

## **License**

MIT License.
