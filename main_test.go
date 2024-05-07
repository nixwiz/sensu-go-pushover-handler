package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	corev2 "github.com/sensu/core/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_checkArgs(t *testing.T) {
	assert := assert.New(t)
	event := corev2.FixtureEvent("entity1", "check1")
	assert.Error(checkArgs(event))
	config.PushoverToken = "abc123"
	assert.Error(checkArgs(event))
	config.PushoverUserKey = "abc123"
	assert.Error(checkArgs(event))
	config.MessageTitleTemplate = "{{.Entity.Name}}/{{.Check.Name}}"
	assert.Error(checkArgs(event))
	config.MessageBodyTemplate = "{{.Check.Output}}"
	assert.NoError(checkArgs(event))
}

func Test_sendPushover(t *testing.T) {
	testcases := []struct {
		status   uint32
		state    string
		priority uint64
	}{
		{0, "OK", 0},
		{1, "WARNING", 0},
		{2, "CRITICAL", 1},
		{127, "UNKNOWN", 1},
	}

	for _, tc := range testcases {
		assert := assert.New(t)
		event := corev2.FixtureEvent("entity1", "check1")
		event.Check.Status = tc.status
		event.Check.Output = tc.state

		var test = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			assert.NoError(err)
			expectedBody := fmt.Sprintf("expire=3600&message=%s&priority=%v&retry=60&sound=pushover&title=entity1%%2Fcheck1&token=123&user=abc", tc.state, tc.priority)
			assert.Equal(expectedBody, strings.Trim(string(body), "\n"))
			w.WriteHeader(http.StatusOK)
			response := PushoverResponse{
				Status:  1,
				Request: "8d4ab099-eac7-475f-ac60-640332ae0aa1",
			}
			responseBytes, _ := json.Marshal(response)
			_, err = w.Write(responseBytes)
			require.NoError(t, err)
		}))

		_, err := url.ParseRequestURI(test.URL)
		require.NoError(t, err)
		config.PushoverAPIURL = test.URL
		config.MessageTitleTemplate = "{{.Entity.Name}}/{{.Check.Name}}"
		config.MessageBodyTemplate = "{{.Check.Output}}"
		config.MessageSound = "pushover"
		config.OkPriority = 0
		config.WarningPriority = 0
		config.CriticalPriority = 1
		config.UnknownPriority = 1
		config.EmergencyRetry = 60
		config.EmergencyExpire = 3600
		config.PushoverToken = "123"
		config.PushoverUserKey = "abc"
		assert.NoError(sendPushover(event))
	}
}

func TestMain(m *testing.M) {
	dir, _ := os.MkdirTemp("", "sensu-go-pushover-handler")
	file, _ := os.CreateTemp(dir, "sensu-go-pushover-handler")

	defer func() {
		_ = os.RemoveAll(dir)
	}()

	event := corev2.FixtureEvent("entity1", "check1")
	event.Metrics = corev2.FixtureMetrics()
	eventJSON, _ := json.Marshal(event)
	if _, err := file.WriteString(string(eventJSON)); err != nil {
		fmt.Printf("Failed to create test file, err %v\n", err)
		os.Exit(1)
	}
	if _, err := file.Seek(0, 0); err != nil {
		fmt.Printf("Failed to seek on test file, err %v\n", err)
		os.Exit(1)
	}

	var test = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// requestReceived = true
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"ok": true}`)); err != nil {
			fmt.Printf("Failed to write HTTP response in test, err: %v", err)
			os.Exit(1)
		}
	}))

	oldStdin := os.Stdin
	os.Stdin = file
	defer func() { os.Stdin = oldStdin }()

	config.PushoverToken = "123"
	config.PushoverUserKey = "abc"
	config.PushoverAPIURL = test.URL

	os.Exit(m.Run())
}
