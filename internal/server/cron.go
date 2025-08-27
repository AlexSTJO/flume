package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)


type CronManager struct {
  CronIDs map[string]cron.EntryID
  Manager *cron.Cron
}

func CronInit() (*CronManager, error) {
  err := godotenv.Load()
  if err != nil {
    return nil, err
  }
  port := os.Getenv("PORT")
  url := os.Getenv("URL")

  requestUrl := fmt.Sprintf("http://%s:%s/run", url, port) 
  fmt.Println(requestUrl)
  home, err := os.UserHomeDir()
  if err != nil {
    return nil, err
  }

  c := cron.New()
  
  root := filepath.Join(home, ".flume")
  cronIDs := make(map[string]cron.EntryID)
  err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
    if filepath.Base(path) == "meta.json" {
      b, err := os.ReadFile(path)
      if err != nil {
        return err
      }
      
      pm := &PipelineMeta{}
      if err = json.Unmarshal(b, pm); err != nil {
        return  err
      }

      _, err, pm = SyncMeta(pm.YamlPath)
      if err != nil {
        return err
      }
      
      if pm.Trigger.Type == "cron" {
        if err = validateCron(pm.Trigger.Cron); err != nil {
          return err
        }
        fmt.Printf("Adding Flume: %s with expression: %s\n", pm.Name, pm.Trigger.Cron )
        id, err := c.AddFunc(pm.Trigger.Cron, addEntry(pm, requestUrl))
        if err != nil {
          return err
        }

        cronIDs[pm.Name] = id 
      }
    }     
    return nil
  })
  if err != nil {
    return nil, err
  }
  fmt.Println("Started")
  c.Start()

  cm := &CronManager{
    CronIDs: cronIDs,
    Manager: c,
  }
  return cm, nil
}


//fw detects meta change, check trigger if croncheck pm name -> c[name] id -> c.remove(id) -> c.addfunc -> update the map after

func (cm *CronManager) ResyncCron(p *PipelineMeta) error {
  port := os.Getenv("PORT")
  url := os.Getenv("URL")

  requestUrl := fmt.Sprintf("http://%s:%s/run", url, port)
  if err := validateCron(p.Trigger.Cron); err != nil {
    return err
  }
  id, ok := cm.CronIDs[p.Name]
  if ok {
    newID, err := cm.Manager.AddFunc(p.Trigger.Cron, addEntry(p, requestUrl))
    if err != nil {
      return err
    }
    cm.CronIDs[p.Name] = newID
    cm.Manager.Remove(id)
    fmt.Printf("Resynced %s's Cron to %s\n",p.Name,  p.Trigger.Cron)
  } 
  return nil 
}

func validateCron(expression string) error {
  _, err := cron.ParseStandard(expression)
  return err
}


func addEntry(pm *PipelineMeta, requestUrl string) func(){ 
  return func() {
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
  }
}

