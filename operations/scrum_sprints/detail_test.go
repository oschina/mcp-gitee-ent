package scrum_sprints

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestGetScrumSprint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/7/programs/9/scrum_sprints/5" {
			t.Errorf("path = %q", request.URL.Path)
		}
		_, _ = io.WriteString(w, `{"id":5}`)
	}))
	defer server.Close()
	previousBase := utils.GetApiBase()
	previousToken := utils.GetGiteeAccessToken()
	utils.SetApiBase(server.URL)
	utils.SetGiteeAccessToken("test-token")
	t.Cleanup(func() { utils.SetApiBase(previousBase); utils.SetGiteeAccessToken(previousToken) })

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{"enterprise_id": 7, "program_id": 9, "scrum_sprint_id": 5}
	result, err := GetScrumSprintHandleFunc(context.Background(), request)
	if err != nil || result.IsError {
		t.Fatalf("handler error = %v result = %#v", err, result)
	}
}
