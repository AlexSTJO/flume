package services

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"

	"github.com/AlexSTJO/flume/internal/logging"
	"github.com/AlexSTJO/flume/internal/resolver"
	"github.com/AlexSTJO/flume/internal/structures"
)

type TestRunnerService struct{}

func (s TestRunnerService) Name() string {
	return "test_runner"
}

func (s TestRunnerService) Parameters() []string {
	return []string{"command"}
}

func (s TestRunnerService) Run(t structures.Task, n string, ctx *structures.Context, infra_outputs *map[string]map[string]string, l *logging.Config, r *structures.RunInfo) error {
	rContext := make(map[string]string, 6)

	raw_command, err := t.StringParam("command")
	if err != nil {
		return err
	}

	command, err := resolver.ResolveStringParam(raw_command, ctx, infra_outputs, r)
	if err != nil {
		rContext["success"] = "false"
		ctx.SetEventValues(n, rContext)
		return err
	}

	outputDir := r.RunDir + "/job_outputs/" + n + "/"
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	outputPath := outputDir + "test_output.txt"

	cmd := exec.Command("sh", "-c", command)
	var combinedBuf bytes.Buffer
	cmd.Stdout = &combinedBuf
	cmd.Stderr = &combinedBuf

	runErr := cmd.Run()

	exitCode := 0
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			rContext["success"] = "false"
			ctx.SetEventValues(n, rContext)
			return fmt.Errorf("failed to execute test command: %w", runErr)
		}
	}

	output := combinedBuf.String()

	err = os.WriteFile(outputPath, []byte(output), 0644)
	if err != nil {
		return fmt.Errorf("failed to write test output: %w", err)
	}

	if len(output) > 0 {
		l.ShellLogger(output)
	}

	passed, failed, total := parseTestCounts(output)

	rContext["success"] = strconv.FormatBool(exitCode == 0)
	rContext["exit_code"] = strconv.Itoa(exitCode)
	rContext["passed"] = passed
	rContext["failed"] = failed
	rContext["total"] = total
	rContext["output_path"] = outputPath
	ctx.SetEventValues(n, rContext)

	return nil
}

func parseTestCounts(output string) (passed, failed, total string) {
	// Go test: "ok" and "FAIL" lines, or "--- PASS" / "--- FAIL"
	goPassRe := regexp.MustCompile(`(?m)^ok\s+`)
	goFailRe := regexp.MustCompile(`(?m)^FAIL\s+`)
	goPassMatches := goPassRe.FindAllString(output, -1)
	goFailMatches := goFailRe.FindAllString(output, -1)
	if len(goPassMatches) > 0 || len(goFailMatches) > 0 {
		p := len(goPassMatches)
		f := len(goFailMatches)
		return strconv.Itoa(p), strconv.Itoa(f), strconv.Itoa(p + f)
	}

	// pytest: "X passed, Y failed" or "X passed"
	pytestRe := regexp.MustCompile(`(\d+)\s+passed(?:.*?(\d+)\s+failed)?`)
	if m := pytestRe.FindStringSubmatch(output); m != nil {
		p := m[1]
		f := "0"
		if m[2] != "" {
			f = m[2]
		}
		pInt, _ := strconv.Atoi(p)
		fInt, _ := strconv.Atoi(f)
		return p, f, strconv.Itoa(pInt + fInt)
	}

	// Jest: "Tests: X passed, Y failed, Z total"
	jestRe := regexp.MustCompile(`Tests:\s+(?:(\d+)\s+passed)?[,\s]*(?:(\d+)\s+failed)?[,\s]*(\d+)\s+total`)
	if m := jestRe.FindStringSubmatch(output); m != nil {
		p := "0"
		f := "0"
		if m[1] != "" {
			p = m[1]
		}
		if m[2] != "" {
			f = m[2]
		}
		return p, f, m[3]
	}

	return "unknown", "unknown", "unknown"
}

func init() {
	structures.Registry["test_runner"] = TestRunnerService{}
}
