package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
  "io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)


func CronInit() error {
  err := godotenv.Load()
  if err != nil {
    return err
  }
  port := os.Getenv("PORT")
  url := os.Getenv("URL")

  requestUrl := fmt.Sprintf("http://%s:%s/run", url, port) 
  fmt.Println(requestUrl)
  home, err := os.UserHomeDir()
  if err != nil {
    return err
  }

  c := cron.New()
  
  root := filepath.Join(home, ".flume")
  err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
    if filepath.Base(path) == "meta.json" {
      b, err := os.ReadFile(path)
      if err != nil {
        return err
      }
      
      var pm PipelineMeta
      if err = json.Unmarshal(b, &pm); err != nil {
        return err
      }
      
      if pm.Trigger.Type == "cron" {
        if err = validateCron(pm.Trigger.Cron); err != nil {
          return err
        }
        fmt.Printf("Adding Flume: %s with expression: %s\n", pm.Name, pm.Trigger.Cron )
        c.AddFunc(pm.Trigger.Cron, func(){
          jsonBody := []byte(fmt.Sprintf(`{"pipeline": "%s"}`, pm.Name))
        
          req, err := http.NewRequest(http.MethodPost, requestUrl, bytes.NewBuffer(jsonBody))
          if err != nil {
            return
          }

          resp, err := http.DefaultClient.Do(req)
          if err != nil {
            return 
          }
          defer resp.Body.Close()
          b, _ := io.ReadAll(resp.Body)
          fmt.Println(string(b))
        })
      }
    }     
    return nil
  })
  if err != nil {
    return err
  }
  fmt.Println("Started")
  c.Start()
  return nil
}

func validateCron(expression string) error {
  _, err := cron.ParseStandard(expression)
  return err
}

