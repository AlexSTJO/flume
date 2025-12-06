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

func ResolveStringParam(v string, ctx *structures.Context, infra *map[string]map[string]string, ) (string, error) {
    v, err := ResolveString(v, ctx, infra)
    return v, err
}

func ResolveAny(v any, ctx *structures.Context, infra *map[string]map[string]string, ) (any, error) {
  switch typed := v.(type) {
  case string:
    return ResolveString(typed, ctx, infra)
  
  case map[string]any:
    out := make(map[string]any, len(typed))
    for k, val := range typed  {
      rv, err := ResolveAny(val, ctx, infra)
      if err != nil {
        return nil, err
      }
      out[k] = rv
    }
    return out, nil
  case [] any:
    out := make([]any, len(typed))
    for i, val := range typed {
      rv, err := ResolveAny(val,ctx,infra)
      if err != nil {
        return nil, err
      }
      out[i] =  rv
    }
    return out, nil
  default:
    return v, nil
  }
}

