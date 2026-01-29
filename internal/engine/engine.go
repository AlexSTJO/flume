package engine

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/AlexSTJO/flume/internal/condition"
	"github.com/AlexSTJO/flume/internal/infra"
	"github.com/AlexSTJO/flume/internal/logging"
	"github.com/AlexSTJO/flume/internal/structures"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
)

type Engine struct {
	FlumeName      string
	RunInfo        *structures.RunInfo
	Flume          *structures.Pipeline
	DisableLogging bool
	Context        *structures.Context
}

func Build(p *structures.Pipeline, r *structures.RunInfo) (*Engine, error) {
	e := &Engine{
		FlumeName:      p.Name,
		RunInfo:        r,
		Flume:          p,
		DisableLogging: p.DisableLogging,
		Context:        structures.NewContext(),
	}

	return e, nil
}

func (e *Engine) Start() error {
	label := color.New(color.FgGreen, color.Bold).SprintFunc()
	value := color.New(color.FgCyan).SprintFunc()
	warn := color.New(color.FgYellow).SprintFunc()

	fmt.Printf("%s %s\n", label("Flume:"), value(e.FlumeName))
	fmt.Printf("%s %s\n", label("ID:"), value(e.RunInfo.RunID))

	logger := logging.New(e.DisableLogging, e.FlumeName, e.RunInfo.RunID, e.RunInfo.RunDir)
	defer logger.Close()
	if logger.LogPath != "" {
		fmt.Printf("%s %s\n", label("Logs:"), value(logger.LogPath))
	} else {
		fmt.Printf("%s %s\n", label("Logs:"), warn("No log file specified"))
	}

	err := godotenv.Load()
	if err != nil {
		logger.ErrorLogger(fmt.Errorf("Error loading .env file"))
	}

	infra_outputs, err := infra.Deploy(e.Flume.Infrastructure, e.RunInfo, logger)
	if err != nil {
		logger.ErrorLogger(err)
		return err
	}

	ctx := structures.NewContext()

	logger.InfoLogger("Graphing Runtime")
	fmt.Println("-------------------")
	g, err := structures.BuildGraph(e.Flume)
	if err != nil {
		logger.ErrorLogger(err)
	}

	if g == nil || len(g.Nodes) == 0 {
		logger.ErrorLogger(fmt.Errorf("Graph is empty"))
	}

	// Will have to set up maxparallel specification in Config

	in := make(map[string]int, len(g.InDeg))
	for n, v := range g.InDeg {
		in[n] = v
	}

	ready := make(chan string, len(g.Nodes))
	var wg sync.WaitGroup

	for n, v := range in {
		if v == 0 {
			ready <- n
		}
	}

	var (
		mu        sync.Mutex
		completed int
		closeOnce sync.Once
	)

	markDone := func(u string) {
		for _, v := range g.Adj[u] {
			mu.Lock()
			in[v]--
			if in[v] == 0 {
				ready <- v
			}
			mu.Unlock()
		}
	}

	worker := func() {
		defer wg.Done()
		for name := range ready {
			logger.InfoLogger(fmt.Sprintf("Worker recieved task: %s", name))
			task := g.Nodes[name]

			result, err := condition.Evaluate(task, ctx, infra_outputs)
			if err != nil {
				logger.WarnLogger(fmt.Sprintf("Condition evaluation failed for '%s': %v, running anyway", name, err))
			}
			if !result.ShouldRun {
				logger.InfoLogger(fmt.Sprintf("Skipping task '%s'. Reason: %s", name, result.Reason))
				ctx.SetEventValues(name, map[string]string{
					"success":     "skipped",
					"skipped":     "true",
					"skip_reason": result.Reason,
				})
				markDone(name)

				mu.Lock()
				completed++
				done := completed == len(g.Nodes)
				mu.Unlock()

				if done {
					closeOnce.Do(func() { close(ready) })
				}
				continue
			}
			svc, ok := structures.Registry[task.Service]
			if !ok {
				logger.ErrorLogger(fmt.Errorf("Unrecognized service"))
			}

			maxAttempts := 1
			if task.Retry.MaxAttempts > 0 {
				maxAttempts = task.Retry.MaxAttempts
			}
			delay := 5 * time.Second
			if task.Retry.Delay != "" {
				if d, parseErr := time.ParseDuration(task.Retry.Delay); parseErr == nil {
					delay = d
				}
			}

			var lastErr error
			for attempt := 1; attempt <= maxAttempts; attempt++ {
				if attempt > 1 {
					logger.WarnLogger(fmt.Sprintf("Retrying task '%s' (attempt %d/%d) after %v", name, attempt, maxAttempts, delay))
					time.Sleep(delay)
				}

				lastErr = svc.Run(task, name, ctx, infra_outputs, logger, e.RunInfo)
				if lastErr == nil {
					break
				}

				if attempt < maxAttempts {
					logger.WarnLogger(fmt.Sprintf("Task '%s' failed (attempt %d/%d): %v", name, attempt, maxAttempts, lastErr))
				}
			}

			if lastErr != nil {
				close(ready)
				logger.ErrorLogger(lastErr)
			}
			markDone(name)

			mu.Lock()
			completed++
			done := completed == len(g.Nodes)
			mu.Unlock()

			if done {
				closeOnce.Do(func() { close(ready) })
			}

		}
	}

	maxParallel := runtime.NumCPU()

	wg.Add(maxParallel)
	for i := 0; i < maxParallel; i++ {
		go worker()
	}

	wg.Wait()

	if completed != len(g.Nodes) {
		logger.ErrorLogger(fmt.Errorf("cycle detected: only completed %d of %d tasks", completed, len(g.Nodes)))
	}

	logger.SuccessLogger("Flume Completed")
	return nil
}
