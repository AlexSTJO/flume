package services

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/AlexSTJO/flume/internal/logging"
	"github.com/AlexSTJO/flume/internal/resolver"
	"github.com/AlexSTJO/flume/internal/structures"
)

type HttpRequest struct {

}

func (s HttpRequest) Name() string {
  return "http_request"
}

func (s HttpRequest) Parameters() []string {
  return []string{"url", "method"}
}

func (s HttpRequest) Run(t structures.Task, n string, ctx *structures.Context, infra_outputs *map[string]map[string]string, l *logging.Config, r *structures.RunInfo) error {
  runCtx := make(map[string]string)
  defer ctx.SetEventValues(n, runCtx)

  rawUrl, err := t.StringParam("url")
  if err != nil { return err }
  url, err := resolver.ResolveStringParam(rawUrl, ctx, infra_outputs)
  if err != nil {
    return fmt.Errorf("resolving url: %w", err)
  }

  rawMethod, err := t.StringParam("method")
  if err != nil { return err }
  method, err := resolver.ResolveStringParam(rawMethod, ctx, infra_outputs)
  if err != nil {
    return fmt.Errorf("resolving url: %w", err)
  }

  rawBody, err := t.StringParam("body")
  if err != nil {
    return err
  }
  body, err := resolver.ResolveStringParam(rawBody, ctx, infra_outputs)
  if err != nil {
    return fmt.Errorf("resolving body: %w", err)
  }

  method = strings.ToUpper(method)

  ValidMethods := map[string]bool{
    "GET": true, "POST": true, "PUT": true, "DELETE": true, "PATCH": true, "HEAD": true, "OPTIONS": true,
  }

  if !ValidMethods[method]{
    return fmt.Errorf("http method invalid: %s", method)
  }

  var reqBody io.Reader
  if body != "" {
    reqBody = bytes.NewBufferString(body)
  }

  req, err := http.NewRequest(method, url, reqBody)
  if err != nil { return fmt.Errorf("Creating Request: %w", err)}

  headersRaw, ok := t.Parameters["headers"].(map[string]any)
  if !ok {
    return fmt.Errorf("header parameters need to be a map")
  }

  res_headers, err := resolver.ResolveAny(headersRaw,ctx, infra_outputs)
  if err != nil { return fmt.Errorf("resolving headers: %w", err) }

  headers, ok := res_headers.(map[string]any)
  if ok {
    for key, val := range headers {
      if strVal, ok := val.(string); ok {
        req.Header.Set(key,strVal)
      }
    }
  }

  client := &http.Client{
    Timeout: 30 & time.Second,
  } 

  l.InfoLogger(fmt.Sprintf("HTTP %s %s", method, url))
  resp, err := client.Do(req)
  if err != nil {
    return fmt.Errorf("executing request: %w", err)
  }
  defer resp.Body.Close()

  respBody, err := io.ReadAll(resp.Body)
  if err != nil {
     return fmt.Errorf("reading response body: %w", err)
  }
  
  runCtx["status_code"] = fmt.Sprintf("%d", resp.StatusCode)
  runCtx["body"] = string(respBody)
  runCtx["content_type"] = resp.Header.Get("Content-Type")

  l.InfoLogger(fmt.Sprintf("Response: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode)))

  if resp.StatusCode >= 200 && resp.StatusCode < 300 {
    runCtx["success"] = "true"
  } else {
    return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
  } 

  return nil 
}

func init() {
  structures.Registry["http_request"] = HttpRequest{}
}
