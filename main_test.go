package main

import (
	"encoding/json"
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

func TestExecuteHandler(t *testing.T) {
	assert := assert.New(t)
	event := corev2.FixtureEvent("entity1", "check1")
	event.Check.Output = "OK"

	var test = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		assert.NoError(err)
		expectedBody := "message=OK&priority=0&title=entity1%2Fcheck1&token=123&user=abc"
		assert.Equal(expectedBody, strings.Trim(string(body), "\n"))
		w.WriteHeader(http.StatusOK)
	}))

	_, err := url.ParseRequestURI(test.URL)
	require.NoError(t, err)
	config.PushoverAPIURL = test.URL
	config.MessageTitleTemplate = "{{.Entity.Name}}/{{.Check.Name}}"
	config.MessageBodyTemplate = "{{.Check.Output}}"
	config.OkPriority = 0
	config.PushoverToken = "123"
	config.PushoverUserKey = "abc"
	assert.NoError(sendPushover(event))
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
