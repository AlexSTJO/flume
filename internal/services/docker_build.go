package services

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/AlexSTJO/flume/internal/logging"
	"github.com/AlexSTJO/flume/internal/resolver"
	"github.com/AlexSTJO/flume/internal/structures"
	"github.com/AlexSTJO/flume/internal/utils"
)

type DockerBuildService struct {
}

func (s DockerBuildService) Name() string {
	return "docker_build"
}

func (s DockerBuildService) Parameters() []string {
	return []string{"build_path", "image_name", "tag", "attachments", "build_args"}
}

func (s DockerBuildService) Run(t structures.Task, n string, ctx *structures.Context, infra_outputs *map[string]map[string]string, l *logging.Config, r *structures.RunInfo) error {
	runCtx := make(map[string]string, 2)
	runCtx["success"] = "false"
	defer ctx.SetEventValues(n, runCtx)
	raw_build_path, err := t.StringParam("build_path")
	if err != nil {
		return err
	}
	build_path, err := resolver.ResolveStringParam(raw_build_path, ctx, infra_outputs, r)
	if err != nil {
		return err
	}

	raw_attachments := t.Parameters["attachments"]
	resolved, err := resolver.ResolveAny(raw_attachments, ctx, infra_outputs, r)
	if err != nil {
		return err
	}

	attachments, err := resolver.ToStringSlice(resolved)
	if err != nil {
		return err
	}
	raw_image_name, err := t.StringParam("image_name")
	if err != nil {
		return err
	}
	image_name, err := resolver.ResolveString(raw_image_name, ctx, infra_outputs, r)
	if err != nil {
		return err
	}

	raw_tag, err := t.StringParam("tag")
	if err != nil {
		return err
	}
	tag, err := resolver.ResolveStringParam(raw_tag, ctx, infra_outputs, r)
	if err != nil {
		return err
	}

	raw_build_args := t.Parameters["build_args"]
	build_args := make(map[string]string)
	if raw_build_args != nil {
		a_build_args, err := resolver.ResolveAny(raw_build_args, ctx, infra_outputs, r)
		if err != nil {
			return err
		}
		t_build_args, ok := a_build_args.(map[string]string)
		if !ok {
			return fmt.Errorf("build args must be key value pairs")
		}
		build_args = t_build_args
	}

	for _, v := range attachments {
		name := filepath.Base(filepath.Clean(v))
		if err := utils.CopyDir(v, filepath.Join(build_path, name)); err != nil {
			return err
		}

	}

	imageRef := fmt.Sprintf("%s:%s", image_name, tag)

	args := []string{
		"build",
		"-t", imageRef,
		"-f", "Dockerfile",
		".",
	}

	flat_build_args := []string{}

	for k, v := range build_args {
		kp := []string{k, v}
		flat_build_args = append(flat_build_args, kp...)
	}

	args = append(args, flat_build_args...)

	cmd := exec.Command("docker", args...)
	cmd.Dir = build_path

	_, err = cmd.CombinedOutput()
	if err != nil {
		return err
	}

	l.InfoLogger(fmt.Sprintf("Image created: %s", imageRef))

	runCtx["success"] = "true"
	runCtx["image"] = imageRef
	return nil

}

func init() {
	structures.Registry["docker_build"] = DockerBuildService{}
}
