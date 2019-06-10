package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"text/template"

	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

const pushoverAPIURL string = "https://api.pushover.net/1/messages"

type HandlerConfigOption struct {
	Value string
	Path  string
	Env   string
}

type HandlerConfig struct {
	PushoverToken        HandlerConfigOption
	PushoverUserKey      HandlerConfigOption
	MessageBodyTemplate  HandlerConfigOption
	MessageTitleTemplate HandlerConfigOption
	Keyspace             string
}

var (
	okPriority       int8
	warningPriority  int8
	criticalPriority int8
	unknownPriority  int8
	stdin            *os.File

	config = HandlerConfig{
		PushoverToken:        HandlerConfigOption{Path: "token", Env: "SENSU_PUSHOVER_TOKEN"},
		PushoverUserKey:      HandlerConfigOption{Path: "user-key", Env: "SENSU_PUSHOVER_USERKEY"},
		MessageBodyTemplate:  HandlerConfigOption{Value: "{{.Check.Output}}", Path: "body-template"},
		MessageTitleTemplate: HandlerConfigOption{Value: "{{.Entity.Name}}/{{.Check.Name}}", Path: "title-template"},
		Keyspace:             "sensu.io/plugins/pushover/config",
	}
	options = []*HandlerConfigOption{
		&config.PushoverToken,
		&config.PushoverUserKey,
		&config.MessageBodyTemplate,
		&config.MessageTitleTemplate,
	}
)

func main() {

	cmd := &cobra.Command{
		Use:   "sensu-go-pushover-handler",
		Short: "The Sensu Pushover handler for sending notifications",
		RunE:  run,
	}

	cmd.Flags().StringVarP(&config.PushoverToken.Value, "pushoverToken", "t", os.Getenv("SENSU_PUSHOVER_TOKEN"), "The Pushover API token, if not in env SENSU_PUSHOVER_TOKEN")
	cmd.Flags().StringVarP(&config.PushoverUserKey.Value, "pushoverUserKey", "u", os.Getenv("SENSU_PUSHOVER_USERKEY"), "The Pushover User Key, if not in env SENSU_PUSHOVER_USERKEY")
	cmd.Flags().StringVarP(&config.MessageTitleTemplate.Value, "messageTitle", "m", config.MessageTitleTemplate.Value, "The message title, in token substitution format")
	cmd.Flags().StringVarP(&config.MessageBodyTemplate.Value, "messageBody", "b", config.MessageBodyTemplate.Value, "The message body, in token substitution format")
	cmd.Flags().Int8VarP(&okPriority, "okPriority", "O", 0, "The priority for OK status messages (default 0)")
	cmd.Flags().Int8VarP(&warningPriority, "warningPriority", "W", 0, "The priority for Warning status messages (default 0)")
	cmd.Flags().Int8VarP(&criticalPriority, "criticalPriority", "C", 1, "The priority for Critical status messages")
	cmd.Flags().Int8VarP(&unknownPriority, "unknownPriority", "U", 1, "The priority for Unknown status messages")
	cmd.Execute()

}

func run(cmd *cobra.Command, args []string) error {

	if stdin == nil {
		stdin = os.Stdin
	}

	eventJSON, err := ioutil.ReadAll(stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %s", err)
	}

	event := &types.Event{}
	err = json.Unmarshal(eventJSON, event)
	if err != nil {
		return fmt.Errorf("failed to unmarshal stdin data: %s", err)
	}

	if err = event.Validate(); err != nil {
		return fmt.Errorf("failed to validate event: %s", err)
	}

	if !event.HasCheck() {
		return fmt.Errorf("event does not contain check")
	}

	configurationOverrides(&config, options, event)

	validationError := checkArgs()
	if validationError != nil {
		return validationError
	}

	pushoverError := sendPushover(event)
	if pushoverError != nil {
		return fmt.Errorf("failed to send to Pushover: %s", pushoverError)
	}

	return nil

}

func checkArgs() error {

	if len(config.PushoverToken.Value) == 0 {
		return errors.New("missing Pushover token")
	}
	if len(config.PushoverUserKey.Value) == 0 {
		return errors.New("missing Pushover user key")
	}
	if len(config.MessageTitleTemplate.Value) == 0 {
		return errors.New("missing message title template")
	}
	if len(config.MessageBodyTemplate.Value) == 0 {
		return errors.New("missing message body template")
	}

	return nil
}

func sendPushover(event *types.Event) error {

	var (
		priority string
	)

	switch event.Check.Status {
	case 0:
		priority = fmt.Sprint(okPriority)
	case 1:
		priority = fmt.Sprint(warningPriority)
	case 2:
		priority = fmt.Sprint(criticalPriority)
	default:
		priority = fmt.Sprint(unknownPriority)
	}

	messageTitle, titleErr := resolveTemplate(config.MessageTitleTemplate.Value, event)
	if titleErr != nil {
		return titleErr
	}

	messageBody, bodyErr := resolveTemplate(config.MessageBodyTemplate.Value, event)
	if bodyErr != nil {
		return bodyErr
	}

	pushoverForm := url.Values{}
	pushoverForm.Add("token", config.PushoverToken.Value)
	pushoverForm.Add("user", config.PushoverUserKey.Value)
	pushoverForm.Add("priority", priority)
	pushoverForm.Add("title", messageTitle)
	pushoverForm.Add("message", messageBody)

	resp, err := http.PostForm(pushoverAPIURL, pushoverForm)
	if err != nil {
		return fmt.Errorf("Post to %s failed: %s", pushoverAPIURL, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("POST to %s failed with %v", pushoverAPIURL, resp.Status)
	}

	return nil
}

func resolveTemplate(templateValue string, event *types.Event) (string, error) {
	var resolved bytes.Buffer
	tmpl, err := template.New("test").Parse(templateValue)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(&resolved, *event)
	if err != nil {
		panic(err)
	}

	return resolved.String(), nil
}

func configurationOverrides(config *HandlerConfig, options []*HandlerConfigOption, event *types.Event) {
	if config.Keyspace == "" {
		return
	}
	for _, opt := range options {
		if opt.Path != "" {
			// compile the Annotation keyspace to look for configuration overrides
			k := path.Join(config.Keyspace, opt.Path)
			switch {
			case event.Check.Annotations[k] != "":
				opt.Value = event.Check.Annotations[k]
				log.Printf("Overriding default handler configuration with value of \"Check.Annotations.%s\" (\"%s\")\n", k, event.Check.Annotations[k])
			case event.Entity.Annotations[k] != "":
				opt.Value = event.Entity.Annotations[k]
				log.Printf("Overriding default handler configuration with value of \"Entity.Annotations.%s\" (\"%s\")\n", k, event.Entity.Annotations[k])
			}
		}
	}
}
