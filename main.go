package main

// build the form
// provide better usage (can i add extra text lines)
// better arg checking

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"text/template"

	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

const pushoverAPIURL string = "https://api.pushover.net/1/messages"

var (
	pushoverToken        string
	pushoverUserKey      string
	messageBodyTemplate  string
	messageTitleTemplate string
	stdin                *os.File
)

func main() {

	cmd := &cobra.Command{
		Use:   "sensu-pushover-handler",
		Short: "The Sensu Pushover handler for sending notifications",
		RunE:  run,
	}

	cmd.Flags().StringVarP(&pushoverToken, "pushoverToken", "t", os.Getenv("PUSHOVER_TOKEN"), "The Pushover API token, if not in env PUSHOVER_TOKEN")
	cmd.Flags().StringVarP(&pushoverUserKey, "pushoverUserKey", "u", os.Getenv("PUSHOVER_USERKEY"), "The Pushover User Key, if not in env PUSHOVER_USERKEY")
	cmd.Flags().StringVarP(&messageTitleTemplate, "messageTitle", "m", "{{.Entity.Name}}/{{.Check.Name}}", "The message title, in token substitution format")
	cmd.Flags().StringVarP(&messageBodyTemplate, "messageBody", "b", "{{.Check.Output}}", "The message body, in token substitution format")
	cmd.Execute()

}

func run(cmd *cobra.Command, args []string) error {

	validationError := checkArgs()
	if validationError != nil {
		return validationError
	}

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

	pushoverError := sendPushover(event)
	if pushoverError != nil {
		return fmt.Errorf("failed to send to Pushover: %s", pushoverError)
	}

	return nil

}

func checkArgs() error {

	if len(pushoverToken) == 0 {
		return errors.New("missing Pushover token")
	}
	if len(pushoverUserKey) == 0 {
		return errors.New("missing Pushover user key")
	}
	if len(messageTitleTemplate) == 0 {
		return errors.New("missing message title template")
	}
	if len(messageBodyTemplate) == 0 {
		return errors.New("missing message body template")
	}

	return nil
}

func sendPushover(event *types.Event) error {

	var (
		priority string
	)

	if event.Check.Status < 3 && event.Check.Status > 0 {
		priority = fmt.Sprint(event.Check.Status - 1)
	} else {
		priority = "0"
	}

	messageTitle, titleErr := resolveTemplate(messageTitleTemplate, event)
	if titleErr != nil {
		return titleErr
	}

	messageBody, bodyErr := resolveTemplate(messageBodyTemplate, event)
	if bodyErr != nil {
		return bodyErr
	}

	pushoverForm := url.Values{}
	pushoverForm.Add("token", pushoverToken)
	pushoverForm.Add("user", pushoverUserKey)
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
