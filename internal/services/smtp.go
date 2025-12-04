package services

import (
	"fmt"
	"net/smtp"
	"strconv"

	"github.com/AlexSTJO/flume/internal/logging"
	"github.com/AlexSTJO/flume/internal/resolver"
	"github.com/AlexSTJO/flume/internal/structures"
	"github.com/jordan-wright/email"
)

type EmailService struct {}


func (s EmailService) Name() string {
  return "send_email"
}

func (s EmailService) Parameters() []string {
  return []string{"username", "password", "host", "recipient", "subject", "body"}
}

func (s EmailService) Run(t structures.Task, n string, ctx *structures.Context, infra_outputs *map[string]map[string]string, l *logging.Config) error {
  tContext := make(map[string]string)
  var err error
  defer func() {
    tContext["success"] = strconv.FormatBool(err == nil)
    ctx.SetEventValues(n, tContext)
  }()
 
  username, err := resolver.ResolveParam("username", t.Parameters, ctx, infra_outputs)
  if err != nil { return err }

  password, err := resolver.ResolveParam("password", t.Parameters, ctx, infra_outputs)
  if err != nil { return err }

  host, err := resolver.ResolveParam("host", t.Parameters, ctx, infra_outputs)
  if err != nil { return err }

  subject, err := resolver.ResolveParam("subject", t.Parameters, ctx, infra_outputs)
  if err != nil { return err }

  body, err := resolver.ResolveParam("body", t.Parameters, ctx, infra_outputs)
  if err != nil { return err }

  recipient, err := resolver.ResolveParam("recipient", t.Parameters, ctx, infra_outputs)
  if err != nil { return err } 

  e := email.NewEmail()
  e.From = username
  e.To = []string{recipient}
  e.Subject = subject
  e.Text = []byte(body)

  addr := fmt.Sprintf("%s:%d", host, 587)
  auth := smtp.PlainAuth("", username, password, host)

  if err := e.Send(addr, auth); err != nil {
		l.ErrorLogger(err)
    return err
	}

  l.InfoLogger("Email Successfully Sent")


  return nil 
  
}

func init() {
  structures.Registry["send_email"] = EmailService{}
}
