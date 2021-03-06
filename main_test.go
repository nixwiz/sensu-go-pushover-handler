package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckArgs(t *testing.T) {
	assert := assert.New(t)
	event := corev2.FixtureEvent("entity1", "check1")
	assert.Error(CheckArgs(event))
	config.PushoverToken = "abc123"
	assert.Error(CheckArgs(event))
	config.PushoverUserKey = "abc123"
	assert.Error(CheckArgs(event))
	config.MessageTitleTemplate = "{{.Entity.Name}}/{{.Check.Name}}"
	assert.Error(CheckArgs(event))
	config.MessageBodyTemplate = "{{.Check.Output}}"
	assert.NoError(CheckArgs(event))
}

func TestSendPushover(t *testing.T) {
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
			body, err := ioutil.ReadAll(r.Body)
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
		assert.NoError(SendPushover(event))
	}
}

func TestMain(t *testing.T) {
	assert := assert.New(t)
	file, _ := ioutil.TempFile(os.TempDir(), "sensu-go-pushover-handler")
	defer func() {
		_ = os.Remove(file.Name())
	}()

	event := corev2.FixtureEvent("entity1", "check1")
	event.Metrics = corev2.FixtureMetrics()
	eventJSON, _ := json.Marshal(event)
	_, err := file.WriteString(string(eventJSON))
	require.NoError(t, err)
	require.NoError(t, file.Sync())
	_, err = file.Seek(0, 0)
	require.NoError(t, err)
	os.Stdin = file
	requestReceived := false

	var test = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived = true
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"ok": true}`))
		require.NoError(t, err)
	}))

	_, err = url.ParseRequestURI(test.URL)
	require.NoError(t, err)
	oldArgs := os.Args
	os.Args = []string{"sensu-go-pushover-handler", "--pushoverAPIURL", test.URL, "--pushoverToken", "123", "--pushoverUserKey", "abc"}
	defer func() { os.Args = oldArgs }()

	main()
	assert.True(requestReceived)
}
