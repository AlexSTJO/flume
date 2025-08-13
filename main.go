package main

import (
	"fmt"

	"github.com/AlexSTJO/flume/internal/logging"
)

func main(){
  c := logging.Config{
    NoColor: false,
    Json: true,
    RunID: "",
    Flume: "",
  }

  c.ErrorLogger(fmt.Errorf("Balerina cappucina"))
  c.InfoLogger("Hello")
  c.SuccessLogger("Bye")

}
