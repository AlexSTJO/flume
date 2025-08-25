package server

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"


	"github.com/fsnotify/fsnotify"
)



func FileWatcher() error {
  pm, err := generateMeta("sample/sample.yaml")
  if err != nil {
    return err
  }
  fmt.Println(pm)
  w, err := fsnotify.NewWatcher()
  if err != nil {
    return fmt.Errorf("Error creating file watcher: %w" , err)
  }

  defer w.Close()

  go watchLoop(w)
  home, err := os.UserHomeDir()
  if err != nil {
    return fmt.Errorf("Error finding home directory: %w", err)
  }
  filepath := filepath.Join(home,  ".flume")
  if err = w.Add(filepath); err != nil {
    return fmt.Errorf("Error adding directory to watcher: %w", err)
  }

  addTree(w, filepath)
  <-make(chan struct{})
  return nil

}


func watchLoop(w *fsnotify.Watcher) {
	i := 0
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

        i++
        fmt.Printf("%3d %s", i, e)
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


