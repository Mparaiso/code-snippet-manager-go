package smartsnippets_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"net/http"
	"net/http/httptest"

	app "github.com/Mparaiso/code-snippet-manager-go"
	"github.com/Mparaiso/expect-go"
	"google.golang.org/appengine/aetest"
)

func SetUpApp(t *testing.T) (aetest.Instance, *app.App, func()) {
	instance, err := aetest.NewInstance(nil)
	expect.Expect(t, err, nil, "Instance creation shouldn't return an error")
	App := app.NewApp()
	return instance, App, func() { instance.Close() }

}

func TestIndex(t *testing.T) {
	t.Log("GET /")
	instance, App, done := SetUpApp(t)
	defer done()
	response := httptest.NewRecorder()
	request, err := instance.NewRequest("GET", "/", nil)
	expect.Expect(t, err, nil, "Request error should be nil")
	App.ServeHTTP(response, request)
	expect.Expect(t, response.Code, 200, "Status should be 200")
	SubTestPostSnippets(t, instance, App)
}

func SubTestPostSnippets(t *testing.T, instance aetest.Instance, App *app.App) {
	t.Log("POST /snippets/")
	response := httptest.NewRecorder()
	snippet := &app.Snippet{Title: "Snippet Title", Description: "Snippet Description"}
	buffer := new(bytes.Buffer)
	err := json.NewEncoder(buffer).Encode(snippet)
	expect.Expect(t, err, nil)
	request, err := instance.NewRequest("POST", "/snippets", buffer)
	expect.Expect(t, err, nil)
	App.ServeHTTP(response, request)
	expect.Expect(t, response.Code, 303, "Status")
	location := response.HeaderMap.Get("Location")
	expect.Expect(t, strings.HasPrefix(location, "/snippets/"), true, fmt.Sprintf("%s", location))
	SubTestGetSnippet(t, instance, App, location, snippet)
}

func SubTestGetSnippet(t *testing.T, instance aetest.Instance, App *app.App, location string, snippet *app.Snippet) {
	t.Logf("GET %s", location)
	request, err := instance.NewRequest("GET", location, nil)
	expect.Expect(t, err, nil)
	response := httptest.NewRecorder()
	App.ServeHTTP(response, request)
	expect.Expect(t, response.Code, http.StatusOK, "Status")
	result := &app.Snippet{}
	err = json.NewDecoder(response.Body).Decode(result)
	expect.Expect(t, err, nil)
	expect.Expect(t, result.Title, snippet.Title, "Snippet.Title")
	expect.Expect(t, result.Version, int64(1), "Snippet.Version")
	SubTestListSnippets(t, instance, App)
}

func SubTestListSnippets(t *testing.T, instance aetest.Instance, App *app.App) {
	t.Log("GET /snippets/")
	response := httptest.NewRecorder()
	request, err := instance.NewRequest("GET", "/snippets", nil)
	expect.Expect(t, err, nil)
	App.ServeHTTP(response, request)
	expect.Expect(t, response.Code, http.StatusOK, "Status")
}
