package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/AlexSTJO/flume/internal/logging"
	"github.com/AlexSTJO/flume/internal/resolver"
	"github.com/AlexSTJO/flume/internal/structures"
)

type SlackService struct{}

type slackMessage struct {
	Text        string `json:"text"`
	Channel     string `json:"channel,omitempty"`
	Username    string `json:"username,omitempty"`
	IconEmoji   string `json:"icon_emoji,omitempty"`
}

func (s SlackService) Name() string {
	return "slack"
}

func (s SlackService) Parameters() []string {
	return []string{"webhook_url", "message"}
}

func (s SlackService) Run(t structures.Task, n string, ctx *structures.Context, infra_outputs *map[string]map[string]string, l *logging.Config, r *structures.RunInfo) error {
	runCtx := make(map[string]string)
	var err error
	defer func() {
		runCtx["success"] = fmt.Sprintf("%t", err == nil)
		ctx.SetEventValues(n, runCtx)
	}()

	rawWebhookURL, err := t.StringParam("webhook_url")
	if err != nil {
		return err
	}
	webhookURL, err := resolver.ResolveStringParam(rawWebhookURL, ctx, infra_outputs, r)
	if err != nil {
		return fmt.Errorf("resolving webhook_url: %w", err)
	}

	rawMessage, err := t.StringParam("message")
	if err != nil {
		return err
	}
	message, err := resolver.ResolveStringParam(rawMessage, ctx, infra_outputs, r)
	if err != nil {
		return fmt.Errorf("resolving message: %w", err)
	}

	msg := slackMessage{
		Text: message,
	}

	// Optional parameters
	if rawChannel, err := t.StringParam("channel"); err == nil {
		if channel, err := resolver.ResolveStringParam(rawChannel, ctx, infra_outputs, r); err == nil && channel != "" {
			msg.Channel = channel
		}
	}

	if rawUsername, err := t.StringParam("username"); err == nil {
		if username, err := resolver.ResolveStringParam(rawUsername, ctx, infra_outputs, r); err == nil && username != "" {
			msg.Username = username
		}
	}

	if rawIconEmoji, err := t.StringParam("icon_emoji"); err == nil {
		if iconEmoji, err := resolver.ResolveStringParam(rawIconEmoji, ctx, infra_outputs, r); err == nil && iconEmoji != "" {
			msg.IconEmoji = iconEmoji
		}
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshaling slack message: %w", err)
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	l.InfoLogger(fmt.Sprintf("Sending Slack message to webhook"))

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("sending slack message: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	runCtx["status_code"] = fmt.Sprintf("%d", resp.StatusCode)
	runCtx["response"] = string(respBody)

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("slack API returned %d: %s", resp.StatusCode, string(respBody))
		return err
	}

	l.InfoLogger("Slack message sent successfully")
	return nil
}

func init() {
	structures.Registry["slack"] = SlackService{}
}
