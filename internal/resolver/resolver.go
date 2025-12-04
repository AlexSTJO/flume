package resolver

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/AlexSTJO/flume/internal/structures"
)

var placeholderRE = regexp.MustCompile(`\$\{([^}]+)\}`)


func ResolveString(s string, ctx *structures.Context, infra_outputs *map[string]map[string]string) (string, error) {
  var e error
  result := placeholderRE.ReplaceAllStringFunc(s, func(m string)(string) {
    key := strings.TrimSpace(m[2:len(m) -1])
    
    parts := strings.SplitN(key, ":" , 2)
    // context:build.status
    // env:dbname
    switch parts[0] {
      case "context": {
        keys := strings.SplitN(parts[1], ".", 2)
        values := ctx.GetEventValues(keys[0])
        result := values[keys[1]]
      
        return result
      }
      case "infra":
        keys := strings.SplitN(parts[1], ".", 2)
        return ((*infra_outputs)[keys[0]][keys[1]])
      case "env": {
        return os.Getenv(parts[1])
      }
    }

    e = fmt.Errorf("Invalid Reference: %s", s)
    return "ERROR"
  })

  if e != nil {
    return "", e
  }

  return result, nil
}

func ResolveParam(key string, params map[string]string, ctx *structures.Context, infra *map[string]map[string]string, ) (string, error) {

    v, err := ResolveString(params[key], ctx, infra)
    return v, err
}

