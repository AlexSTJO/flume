package services

import (
	"fmt"
	"os/exec"

	"github.com/AlexSTJO/flume/internal/logging"
	"github.com/AlexSTJO/flume/internal/resolver"
	"github.com/AlexSTJO/flume/internal/structures"
)
type DockerBuildService struct {
}

func (s DockerBuildService) Name() string {
	return "docker_build"
}

func (s DockerBuildService) Parameters() []string {
	return []string{"build_path", "attach", "image_name", "tag", "build_args"}
}

func (s DockerBuildService) Run(t structures.Task, n string, ctx *structures.Context, infra_outputs *map[string]map[string]string, l*logging.Config) error{
  runCtx := make(map[string]string, 2)
  runCtx["success"] = "false"
  defer ctx.SetEventValues(n, runCtx)
  raw_build_path, err := t.StringParam("repo_url")
  if err != nil { return err }
  build_path, err := resolver.ResolveStringParam(raw_build_path, ctx, infra_outputs)
  if err != nil { return err }

  attach, err := resolver.ResolveAny(t.Parameters["attach"], ctx, infra_outputs)
  if err != nil { return err }

  raw_image_name, err := t.StringParam("image_name")
  if err != nil { return err}
  image_name, err := resolver.ResolveString(raw_image_name, ctx, infra_outputs)
  if err != nil {return err}

  raw_tag, err := t.StringParam("tag")
  if err != nil { return err}
  tag, err := resolver.ResolveStringParam(raw_tag, ctx, infra_outputs)

  raw_build_args := t.Parameters["build_args"]
  a_build_args, err := resolver.ResolveAny(raw_build_args, ctx, infra_outputs)
  build_args, ok := a_build_args.(map[string]string)
  if !ok { return fmt.Errorf("build args must be key value pairs")}

  imageRef := fmt.Sprintf("%s:%s", image_name, tag)

  args := []string{
    "build",
    "t", imageRef,
    "-f", "Dockerfile",
  }

  flat_build_args := []string{}

  for k,v := range(build_args) {
    kp := []string{k,v}
    flat_build_args = append(flat_build_args, kp...)
  }

  args = append(args, flat_build_args...)
  
  cmd := exec.Command("docker", args...)
  cmd.Dir = build_path

  _, err = cmd.CombinedOutput()
  if err != nil {
    return nil
  }

  runCtx["success"] = "true"
  runCtx["image"] = imageRef
  return nil 
  
}

func init() {
	structures.Registry["docker_build"] = DockerBuildService{}
}
