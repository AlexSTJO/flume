package resolver

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/AlexSTJO/flume/internal/structures"
)

var placeholderRE = regexp.MustCompile(`\$\{([^}]+)\}`)


func ResolveString(s string, ctx *structures.Context, infra map[string][]string) (string, error) {
  var e error
  result := placeholderRE.ReplaceAllStringFunc(s, func(m string)(string) {
    key := strings.TrimSpace(m[2:len(m) -1])
    
    parts := strings.SplitN(key, ":" , 2)
    // context:build.status
    // infra:ec2check
    // env:dbname
    switch parts[0] {
      case "context": {
        keys := strings.SplitN(parts[1], ".", 2)
        values := ctx.GetEventValues(keys[0])
        result := values[keys[1]]
      
        return result
      }
      case "infra": {
        rawResult := infra[parts[1]]
        var result string
        for i, n := range(rawResult) {
          if i == 0 {
            result = n
          } else {
            result = result + "," + n
          }
        }
        return result
      }
      case "env": {
        return os.Getenv(parts[1])
      }
    }

    e = fmt.Errorf("Invalid Referenc: %s", s)
    return "ERROR"
  })

  if e != nil {
    return "", e
  }

  return result, nil
}

