package condition

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/AlexSTJO/flume/internal/resolver"
	"github.com/AlexSTJO/flume/internal/structures"
)

type Result struct {
	ShouldRun bool
	Reason    string
}

var conditionRE = regexp.MustCompile(`^(.+?)\s*(==|!=)\s*(.+)$`)

func Evaluate(t structures.Task, c *structures.Context, i *map[string]map[string]string) (Result, error) {
	var result Result
	if t.RunIf == "" && t.SkipIf == "" {
		return Result{ShouldRun: true, Reason: ""}, nil
	}
	if t.RunIf != "" {
		boo, err := evaluateCondition(t.RunIf, c, i)
		if err != nil {
			return result, err
		}
		if boo {
			result.ShouldRun = true
			result.Reason = "'run_if' condition evaluated to 'true'"
		} else {
			result.ShouldRun = false
			result.Reason = "'run_if' condition evaluated to 'false'"
		}
	} else if t.SkipIf != "" {
		boo, err := evaluateCondition(t.SkipIf, c, i)
		if err != nil {
			return result, err
		}
		if boo {
			result.ShouldRun = false
			result.Reason = "'skip_if' condition evaluated to 'true'"
		} else {
			result.ShouldRun = true
			result.Reason = "'skip_if' condition evaluated to 'false'"
		}
	}
	return result, nil
}

func evaluateCondition(s string, c *structures.Context, i *map[string]map[string]string) (bool, error) {
	matches := conditionRE.FindStringSubmatch(s)
	if matches == nil {
		return false, fmt.Errorf("Invalid format for condition: %s", s)
	}

	left := strings.TrimSpace(matches[1])
	resolved_left, err := resolver.ResolveString(left, c, i)
	if err != nil {
		return false, err
	}

	op := strings.TrimSpace(matches[2])

	right := strings.TrimSpace(matches[3])
	resolved_right, err := resolver.ResolveString(right, c, i)
	if err != nil {
		return false, err
	}

	if op == "==" {
		if resolved_left == resolved_right {
			return true, nil
		}
	} else {
		if resolved_right != resolved_left {
			return true, nil
		}
	}

	return false, nil

}
