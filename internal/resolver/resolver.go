package resolver

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/AlexSTJO/flume/internal/structures"
)

var placeholderRE = regexp.MustCompile(`\$\{([^}]+)\}`)


func ResolveString(s string, ctx *structures.Context) (string, error) {
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

