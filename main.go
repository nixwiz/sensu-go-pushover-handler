package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"text/template"

	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

const (
	pushoverToken    = "pushoverToken"
	pushoverUserKey  = "pushoverUserKey"
	messageBody      = "messageBody"
	messageTitle     = "messageTitle"
	messageSound     = "messageSound"
	okPriority       = "okPriority"
	warningPriority  = "warningPriority"
	criticalPriority = "criticalPriority"
	unknownPriority  = "unknownPriority"
	pushoverAPIURL   = "pushoverAPIURL"
)

// HandlerConfig is needed for Sensu Go Handlers
type HandlerConfig struct {
	sensu.PluginConfig
	PushoverToken        string
	PushoverUserKey      string
	MessageBodyTemplate  string
	MessageTitleTemplate string
	MessageSound         string
	OkPriority           uint64
	WarningPriority      uint64
	CriticalPriority     uint64
	UnknownPriority      uint64
	PushoverAPIURL       string
}

var (
	config = HandlerConfig{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-go-pushover-handler",
			Short:    "The Sensu Go Pushover handler for sending notifications",
			Keyspace: "sensu.io/plugins/pushover/config",
		},
	}

	pushoverConfigOptions = []*sensu.PluginConfigOption{
		{
			Path:      pushoverToken,
			Env:       "SENSU_PUSHOVER_TOKEN",
			Argument:  pushoverToken,
			Shorthand: "t",
			Default:   "",
			Usage:     "The Pushover API token",
			Value:     &config.PushoverToken,
		},
		{
			Path:      pushoverUserKey,
			Env:       "SENSU_PUSHOVER_USERKEY",
			Argument:  pushoverUserKey,
			Shorthand: "u",
			Default:   "",
			Usage:     "The Pushover API token",
			Value:     &config.PushoverUserKey,
		},
		{
			Path:      messageTitle,
			Argument:  messageTitle,
			Shorthand: "m",
			Default:   "{{.Entity.Name}}/{{.Check.Name}}",
			Usage:     "The message title, in token substitution format",
			Value:     &config.MessageTitleTemplate,
		},
		{
			Path:      messageSound,
			Argument:  messageSound,
			Shorthand: "s",
			Default:   "pushover",
			Usage:     "The sound for the message",
			Value:     &config.MessageSound,
		},
		{
			Path:      messageBody,
			Argument:  messageBody,
			Shorthand: "b",
			Default:   "{{.Check.Output}}",
			Usage:     "The message body, in token substitution format",
			Value:     &config.MessageBodyTemplate,
		},
		{
			Path:      okPriority,
			Argument:  okPriority,
			Shorthand: "O",
			Default:   uint64(0),
			Usage:     "The priority for OK status messages (default 0)",
			Value:     &config.OkPriority,
		},
		{
			Path:      warningPriority,
			Argument:  warningPriority,
			Shorthand: "W",
			Default:   uint64(0),
			Usage:     "The priority for Warning status messages (default 0)",
			Value:     &config.WarningPriority,
		},
		{
			Path:      criticalPriority,
			Argument:  criticalPriority,
			Shorthand: "C",
			Default:   uint64(1),
			Usage:     "The priority for Critical status messages",
			Value:     &config.CriticalPriority,
		},
		{
			Path:      unknownPriority,
			Argument:  unknownPriority,
			Shorthand: "U",
			Default:   uint64(1),
			Usage:     "The priority for Unknown status messages",
			Value:     &config.UnknownPriority,
		},
		{
			Path:      pushoverAPIURL,
			Argument:  pushoverAPIURL,
			Shorthand: "a",
			Default:   "https://api.pushover.net/1/messages",
			Usage:     "The Pushover API URL",
			Value:     &config.PushoverAPIURL,
		},
	}
)

func main() {

	goHandler := sensu.NewGoHandler(&config.PluginConfig, pushoverConfigOptions, CheckArgs, SendPushover)
	goHandler.Execute()

}

func CheckArgs(_ *corev2.Event) error {

	if len(config.PushoverToken) == 0 {
		return errors.New("missing Pushover token")
	}
	if len(config.PushoverUserKey) == 0 {
		return errors.New("missing Pushover user key")
	}
	if len(config.MessageTitleTemplate) == 0 {
		return errors.New("missing message title template")
	}
	if len(config.MessageBodyTemplate) == 0 {
		return errors.New("missing message body template")
	}

	return nil
}

func SendPushover(event *corev2.Event) error {

	var (
		priority string
	)

	switch event.Check.Status {
	case 0:
		priority = fmt.Sprint(config.OkPriority)
	case 1:
		priority = fmt.Sprint(config.WarningPriority)
	case 2:
		priority = fmt.Sprint(config.CriticalPriority)
	default:
		priority = fmt.Sprint(config.UnknownPriority)
	}

	messageTitle, titleErr := resolveTemplate(config.MessageTitleTemplate, event)
	if titleErr != nil {
		return titleErr
	}

	messageBody, bodyErr := resolveTemplate(config.MessageBodyTemplate, event)
	if bodyErr != nil {
		return bodyErr
	}

	pushoverForm := url.Values{}
	pushoverForm.Add("token", config.PushoverToken)
	pushoverForm.Add("user", config.PushoverUserKey)
	pushoverForm.Add("priority", priority)
	pushoverForm.Add("sound", strings.ToLower(config.MessageSound))
	pushoverForm.Add("title", messageTitle)
	pushoverForm.Add("message", messageBody)

	resp, err := http.PostForm(config.PushoverAPIURL, pushoverForm)
	if err != nil {
		return fmt.Errorf("Post to %s failed: %s", config.PushoverAPIURL, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("POST to %s failed with %v", config.PushoverAPIURL, resp.Status)
	}

	return nil
}

func resolveTemplate(templateValue string, event *corev2.Event) (string, error) {
	var resolved bytes.Buffer
	tmpl, err := template.New("test").Parse(templateValue)
	if err != nil {
		return "", err
	}
	err = tmpl.Execute(&resolved, *event)
	if err != nil {
		return "", err
	}

	return resolved.String(), nil
}
