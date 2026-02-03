package services

import (
	"fmt"
	"time"

	"github.com/AlexSTJO/flume/internal/logging"
	"github.com/AlexSTJO/flume/internal/resolver"
	"github.com/AlexSTJO/flume/internal/structures"
)

type WaitService struct{}

func (s WaitService) Name() string {
	return "wait"
}

func (s WaitService) Parameters() []string {
	return []string{"duration"}
}

func (s WaitService) Run(t structures.Task, n string, ctx *structures.Context, infra_outputs *map[string]map[string]string, l *logging.Config, r *structures.RunInfo) error {
	runCtx := make(map[string]string, 1)
	defer ctx.SetEventValues(n, runCtx)
	runCtx["success"] = "false"

	u_duration, err := t.StringParam("duration")
	if err != nil {
		return err
	}
	durationStr, err := resolver.ResolveStringParam(u_duration, ctx, infra_outputs, r)
	if err != nil {
		return err
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w (use format like 5s, 1m, 500ms)", durationStr, err)
	}

	l.Info(fmt.Sprintf("Waiting for %s", duration))
	time.Sleep(duration)

	runCtx["success"] = "true"
	return nil
}

func init() {
	structures.Registry["wait"] = WaitService{}
}
