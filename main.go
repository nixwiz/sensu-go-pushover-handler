package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	"github.com/sensu-community/sensu-plugin-sdk/templates"
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
	emergencyRetry   = "emergencyRetry"
	emergencyExpire  = "emergencyExpire"
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
	EmergencyRetry       uint64
	EmergencyExpire      uint64
	PushoverAPIURL       string
}

type PushoverResponse struct {
	Status  int      `json:"status"`
	Request string   `json:"request"`
	Errors  []string `json:"errors,omitempty"`
	Receipt string   `json:"receipt,omitempty"`
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
			Secret:    true,
			Usage:     "The Pushover API token",
			Value:     &config.PushoverToken,
		},
		{
			Path:      pushoverUserKey,
			Env:       "SENSU_PUSHOVER_USERKEY",
			Argument:  pushoverUserKey,
			Shorthand: "u",
			Default:   "",
			Secret:    true,
			Usage:     "The Pushover API token",
			Value:     &config.PushoverUserKey,
		},
		{
			Path:      messageTitle,
			Argument:  messageTitle,
			Shorthand: "m",
			Default:   "{{.Entity.Name}}/{{.Check.Name}}",
			Usage:     "The message title template",
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
			Usage:     "The message body template",
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
			Path:      emergencyRetry,
			Argument:  emergencyRetry,
			Shorthand: "R",
			Default:   uint64(60),
			Usage:     "How often, in seconds, to send the same notification to the user, only relevant to Priority 2 messages",
			Value:     &config.EmergencyRetry,
		},
		{
			Path:      emergencyExpire,
			Argument:  emergencyExpire,
			Shorthand: "E",
			Default:   uint64(3600),
			Usage:     "How long, in seconds, to continue sending the same notification to the user, only relevant to Priority 2 messages",
			Value:     &config.EmergencyExpire,
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
	if config.EmergencyExpire > 10800 {
		return errors.New("expire argument too large, > 10800")
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

	messageTitle, titleErr := templates.EvalTemplate("messageTitle", config.MessageTitleTemplate, event)
	if titleErr != nil {
		return titleErr
	}

	messageBody, bodyErr := templates.EvalTemplate("messageBody", config.MessageBodyTemplate, event)
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
	pushoverForm.Add("retry", fmt.Sprintf("%d", config.EmergencyRetry))
	pushoverForm.Add("expire", fmt.Sprintf("%d", config.EmergencyExpire))

	resp, err := http.PostForm(config.PushoverAPIURL, pushoverForm)
	if err != nil {
		return fmt.Errorf("Post to %s failed: %s", config.PushoverAPIURL, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("POST to %s failed with %v", config.PushoverAPIURL, resp.Status)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Failed to read response body from %s: %v", config.PushoverAPIURL, err)
	}

	pushoverResponse := PushoverResponse{}
	err = json.Unmarshal(body, &pushoverResponse)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal resonse from Pushover: %v", err)
	}

	// FUTURE: send to AH
	fmt.Printf("Submitted request ID %s to Pushover\n", pushoverResponse.Request)

	return nil
}
