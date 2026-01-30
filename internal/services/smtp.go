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

type EmailService struct{}

func (s EmailService) Name() string {
	return "send_email"
}

func (s EmailService) Parameters() []string {
	return []string{"username", "password", "host", "recipient", "subject", "body"}
}

func (s EmailService) Run(t structures.Task, n string, ctx *structures.Context, infra_outputs *map[string]map[string]string, l *logging.Config, r *structures.RunInfo) error {
	tContext := make(map[string]string)
	var err error
	defer func() {
		tContext["success"] = strconv.FormatBool(err == nil)
		ctx.SetEventValues(n, tContext)
	}()

	raw_username, err := t.StringParam("username")
	if err != nil {
		return err
	}
	username, err := resolver.ResolveStringParam(raw_username, ctx, infra_outputs, r)
	if err != nil {
		return err
	}

	raw_password, err := t.StringParam("password")
	if err != nil {
		return err
	}
	password, err := resolver.ResolveStringParam(raw_password, ctx, infra_outputs, r)
	if err != nil {
		return err
	}

	raw_host, err := t.StringParam("host")
	if err != nil {
		return err
	}
	host, err := resolver.ResolveStringParam(raw_host, ctx, infra_outputs, r)
	if err != nil {
		return err
	}

	raw_subject, err := t.StringParam("subject")
	if err != nil {
		return err
	}
	subject, err := resolver.ResolveStringParam(raw_subject, ctx, infra_outputs, r)
	if err != nil {
		return err
	}

	raw_body, err := t.StringParam("body")
	if err != nil {
		return err
	}
	body, err := resolver.ResolveStringParam(raw_body, ctx, infra_outputs, r)
	if err != nil {
		return err
	}

	raw_recipient, err := t.StringParam("recipient")
	if err != nil {
		return err
	}
	recipient, err := resolver.ResolveStringParam(raw_recipient, ctx, infra_outputs, r)
	if err != nil {
		return err
	}

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
