package scrum_versions

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestGetScrumVersionAndStages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/7/programs/9/scrum_versions/5", "/7/programs/9/scrum_versions/5/scrum_stages":
		default:
			t.Errorf("unexpected path %s", request.URL.Path)
		}
		_, _ = io.WriteString(w, `{"data":[],"total_count":0}`)
	}))
	defer server.Close()
	previousBase := utils.GetApiBase()
	previousToken := utils.GetGiteeAccessToken()
	utils.SetApiBase(server.URL)
	utils.SetGiteeAccessToken("test-token")
	t.Cleanup(func() { utils.SetApiBase(previousBase); utils.SetGiteeAccessToken(previousToken) })

	args := map[string]any{"enterprise_id": 7, "program_id": 9, "scrum_version_id": 5}
	for _, handler := range []utils.OptionHandlerFunc{GetScrumVersionHandleFunc, ListScrumStagesHandleFunc} {
		request := mcp.CallToolRequest{}
		request.Params.Arguments = args
		result, err := handler(context.Background(), request)
		if err != nil || result.IsError {
			t.Fatalf("handler error = %v result = %#v", err, result)
		}
	}
}
