# **Flume**
A declarative workflow engine for hybrid cloud automation — powered by Go.

Flume is a lightweight, extensible workflow orchestrator designed to automate builds, deployments, and infrastructure using a simple YAML format. It integrates with Terraform, AWS services, Git, and custom Go services, enabling hybrid workflows that unify application deployment and infrastructure provisioning.

---

## **Features**

- Declarative YAML pipelines  
- Automatic DAG execution  
- Dynamic resolver engine  
- Terraform integration  
- AWS services: S3 upload, CloudFront invalidation  
- Git clone & update service  
- Shell execution  
- JSON writer  
- Modular service registry  
- API-triggered pipelines 
- Cron Support
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

Pipelines are defined in YAML and referenced by name. To trigger a pipeline, send an HTTP POST request to the `/run` endpoint of the Flume server.

Example using `curl`:

```bash
curl -X POST "http://$URL:$PORT/run"   -H "Content-Type: application/json"   -d '{"pipeline": "sample-flume"}'
```

Where:

- `$URL` and `$PORT` come from your `.env` file  
- `pipeline` is the name of the pipeline (e.g., `sample-flume`)

---

## **Architecture**

### Services

```go
type Service interface {
    Name() string
    Parameters() []string
    Call(t structures.Task) error
}
```

```go
func init() {
    registry["json_writer"] = &JSONWriter{}
}
```

### Resolver Patterns

| Pattern | Meaning |
|--------|---------|
| `${context:task.value}` | Task output |
| `${infra:terraform.output}` | Terraform output |
| `${env:VAR}` | Environment variable |
| `${timestamp}` | Execution timestamp |

---

## **Example Pipeline**

```yaml
name: "sample-flume"
trigger:
  type: "api"

log_path: ""

infrastructure:
  tfdeploy:
    service: terraform
    action: sync
    key: "/terraform/portfolio-website"
    var-file: "terraform.tfvars"

tasks:

  git_pull:
    version: 1
    service: git
    dependencies: []
    parameters:
      repo_url: "https://github.com/AlexSTJO/portfolio-website"

  build:
    version: 1
    service: shell
    dependencies: ["git_pull"]
    parameters:
      command: |
        cd ${context:git_pull.repo_folder}
        npm install
        npm run build  

  upload:
    version: 1
    service: s3_upload
    dependencies: ["build"]
    parameters:
      bucket: ${infra:terraform.site_bucket_name}
      source: ${context:git_pull.repo_folder}/out
      prefix: "build/"

  meta:
    version: 1
    service: json_writer
    dependencies: ["upload"]
    parameters:
      data:
        infra_status: ${infra:terraform.deploy_success}
        pull_status: ${context:git_pull.success}
        build_status: ${context:build.success}
        upload_status: ${context:upload.success}

  meta_upload:
    version: 1
    service: s3_upload
    dependencies: ["meta"]
    parameters:
      bucket: ${infra:terraform.site_bucket_name}
      source: ${context:meta.json_path}
      prefix: "build/meta/"

  cf_invalidate:
    version: 1 
    service: cloudfront_invalidate
    dependencies: ["meta_upload"]
    parameters:
      dist_id: ${infra:terraform.dist_id}
      paths: ["/*"]

  end_message:
    version: 1 
    service: shell
    dependencies: ["cf_invalidate"]
    parameters:
      command: "echo ${infra:terraform.deploy_success} with url of ${infra:terraform.portfolio_url}"
```

---

## **Custom Service Example**

```go
type MyService struct{}

func (m *MyService) Name() string { return "my_service" }

func (m *MyService) Parameters() []string { return []string{"sample_p"}}

func (m *MyService) Call(d structures.Deployment) error {
    return nil
}

func init() {
    registry["my_service"] = &MyService{}
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
