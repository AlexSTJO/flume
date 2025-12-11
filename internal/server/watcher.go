package server

import (
	"fmt"
	"io/fs"
	"path/filepath"


	"github.com/fsnotify/fsnotify"
)



func FileWatcher(c *CronManager) error {
  w, err := fsnotify.NewWatcher()
  if err != nil {
    return fmt.Errorf("Error creating file watcher: %w" , err)
  }

  defer w.Close()

  go watchLoop(w, c)
    filepath := filepath.Join(".",  ".flume")
  if err = w.Add(filepath); err != nil {
    return fmt.Errorf("Error adding directory to watcher: %w", err)
  }

  addTree(w, filepath)
  <-make(chan struct{})
  return nil

}


func watchLoop(w *fsnotify.Watcher, c *CronManager) {
	for {
		select {
      case err, ok := <-w.Errors:
        if !ok { 
          return
        }
        fmt.Printf("ERROR: %s", err)
      case e, ok := <-w.Events:
        if !ok { 
          return
        }
        
        if e.Op&fsnotify.Write == fsnotify.Write || e.Op&fsnotify.Create == fsnotify.Create { 
          ext := filepath.Ext(e.Name)
          if ext == ".yaml" {
            change, err, pm := SyncMeta(e.Name)
            if err != nil {
              fmt.Printf("ERROR: %s", err)
            }
            if change {
              if pm.Trigger.Type == "cron"{
                c.ResyncCron(pm)
              } 
              fmt.Printf("Synced file: %s\n", e.Name)
            }
          }
      }
	  }
  }
}

func addTree(w *fsnotify.Watcher, root string) error {
  return filepath.WalkDir(root, func (p string, d fs.DirEntry, err error) error {
    if err != nil { return err}
    if d.IsDir() {
      if err := w.Add(p); err != nil {
        return fmt.Errorf("Error adding directory to watcher: %v", err)
      }
    }

    return nil
  })
}


