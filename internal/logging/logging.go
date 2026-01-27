package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
)

type Config struct {
	NoColor        bool
	DisableLogging bool
	LogPath        string
	logFile        *os.File
}

type LogLine struct {
	TS    string `json:"ts"`
	Level string `json:"level"`
	Msg   string `json:"msg"`
}

var red = color.New(color.FgRed)
var green = color.New(color.FgGreen)
var cyan = color.New(color.FgCyan)
var yellow = color.New(color.FgYellow)

func timeStamp() time.Time {
	t := time.Now()
	return t
}

func New(disable_logging bool, name string, run_id string, run_dir string) *Config {
	var c *Config
	if disable_logging {
		c = &Config{
			NoColor:        false,
			DisableLogging: disable_logging,
			LogPath:        "",
		}
	} else {
		log_path := filepath.Join(run_dir, "logs", run_id+".jsonl")
		c = &Config{
			NoColor:        false,
			DisableLogging: disable_logging,
			LogPath:        log_path,
			logFile:        nil,
		}
		logFile, err := c.open()
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: Failed to open log file, file logging disabled: %v\n", err)
			c.DisableLogging = true
		} else {
			c.logFile = logFile
		}
	}
	return c
}

func (c *Config) ErrorLogger(e error) {
	if c.NoColor {
		fmt.Printf(timeStamp().Format(time.TimeOnly)+"  ERROR    "+" %v\n", e)
	} else {
		red.Printf(timeStamp().Format(time.TimeOnly)+"  ERROR    "+" %v\n", e)
	}

	if !c.DisableLogging {
		c.PipeLogsToFile("ERROR", e.Error())
	}
}

func (c *Config) InfoLogger(s string) {
	fmt.Printf(timeStamp().Format(time.TimeOnly)+"  INFO     "+" %v\n", s)
	if !c.DisableLogging {
		c.PipeLogsToFile("INFO", s)
	}
}
func (c *Config) SuccessLogger(s string) {
	if c.NoColor {
		fmt.Printf(timeStamp().Format(time.TimeOnly)+"  SUCCESS  "+" %v\n", s)
	} else {
		green.Printf(timeStamp().Format(time.TimeOnly)+"  SUCCESS  "+" %v\n", s)
	}

	if !c.DisableLogging {
		c.PipeLogsToFile("SUCCESS", s)
	}

}

func (c *Config) ShellLogger(s string) {
	if c.NoColor {
		fmt.Printf(timeStamp().Format(time.TimeOnly)+"  SHELL    "+" %v", s)
	} else {
		cyan.Printf(timeStamp().Format(time.TimeOnly)+"  SHELL    "+" %v", s)
	}

	if !c.DisableLogging {
		c.PipeLogsToFile("SHELL", s)
	}

}

func (c *Config) WarnLogger(s string) {
	if c.NoColor {
		fmt.Printf(timeStamp().Format(time.TimeOnly)+"  WARN     "+" %v\n", s)
	} else {
		yellow.Printf(timeStamp().Format(time.TimeOnly)+"  WARN     "+" %v\n", s)
	}

	if !c.DisableLogging {
		c.PipeLogsToFile("WARN", s)
	}
}

func (c *Config) PipeLogsToFile(level string, msg string) {
	if c.logFile == nil {
		return
	}

	entry := LogLine{
		TS:    time.Now().Format(time.TimeOnly),
		Level: level,
		Msg:   msg,
	}

	enc := json.NewEncoder(c.logFile)
	if err := enc.Encode(entry); err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: Failed to write to log file: %v\n", err)
	}
}

func (c *Config) open() (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(c.LogPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	f, err := os.OpenFile(c.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return f, nil
}

func (c *Config) Close() {
	if c.logFile != nil {
		c.logFile.Close()
	}
}
