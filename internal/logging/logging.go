package logging

import (
	"fmt"
	"time"
  "os"
  "encoding/json"

	"github.com/fatih/color"
)

type Config struct {
  NoColor bool
  LogPath string
}

type LogLine struct {
	TS    string `json:"ts"`
	Level string `json:"level"`
	Msg   string `json:"msg"`
}

var red = color.New(color.FgRed)
var green = color.New(color.FgGreen)


func timeStamp() time.Time {
  t := time.Now()
  return t
}
  

func (c *Config) ErrorLogger(e error) {
  if c.NoColor {
    fmt.Printf(timeStamp().Format(time.TimeOnly) + "  ERROR    " + " %v\n", e)
  } else {
    red.Printf(timeStamp().Format(time.TimeOnly) + "  ERROR    " + " %v\n", e)
  }

  if c.LogPath != "" {
    f, err := os.OpenFile(c.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
      panic(err)
    }
    defer f.Close()

    entry := LogLine{
      TS: time.Now().Format(time.TimeOnly),
      Level: "ERROR",
      Msg: e.Error(),
    }

    enc := json.NewEncoder(f)
    if err := enc.Encode(entry); err != nil {
      panic(err)
    }
  }
}

func (c *Config) InfoLogger(s string) {
  fmt.Printf(timeStamp().Format(time.TimeOnly) + "  INFO     " + " %v\n", s)

  if c.LogPath != ""{
    f, err := os.OpenFile(c.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
      panic(err)
    }
    defer f.Close()

    entry := LogLine{
      TS: time.Now().Format(time.TimeOnly),
      Level: "INFO",
      Msg: s,
    }

    enc := json.NewEncoder(f)
    if err := enc.Encode(entry); err != nil {
      panic(err)
    }
  }
}
func (c *Config) SuccessLogger(s string) {
  if c.NoColor {
    fmt.Printf(timeStamp().Format(time.TimeOnly) + "  SUCCESS  " + " %v\n", s)
  } else {
    green.Printf(timeStamp().Format(time.TimeOnly) + "  SUCCESS  " + " %v\n", s)
  }


  if c.LogPath != "" {
    f, err := os.OpenFile(c.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
      panic(err)
    }
    defer f.Close()

    entry := LogLine{
      TS: time.Now().Format(time.TimeOnly),
      Level: "SUCCESS",
      Msg: s,
    }

    enc := json.NewEncoder(f)
    if err := enc.Encode(entry); err != nil {
      panic(err)
    }
  }
}





  
