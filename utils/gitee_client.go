package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"io"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	DefaultApiBase = "https://api.gitee.com/enterprises"
)

var (
	giteeAccessToken string
	apiBase          string
)

func SetGiteeAccessToken(token string) {
	giteeAccessToken = token
}

func SetApiBase(url string) {
	apiBase = url
}

func GetGiteeAccessToken() string {
	if giteeAccessToken != "" {
		return giteeAccessToken
	}
	return os.Getenv("GITEE_ENT_MCP_ACCESS_TOKEN")
}

func GetApiBase() string {
	if apiBase != "" {
		return apiBase
	}
	if envApiBase := os.Getenv("GITEE_ENT_API_BASE"); envApiBase != "" {
		return envApiBase
	}
	return DefaultApiBase
}

type Authorizer interface {
	Authorize(req *http.Request) error
}

type BearerTokenAuthorizer struct{}

func (b *BearerTokenAuthorizer) Authorize(req *http.Request) error {
	accessToken := GetGiteeAccessToken()
	if accessToken == "" {
		return NewAuthError()
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	return nil
}

type CookieAuthorizer struct {
	Cookie string
}

func (c *CookieAuthorizer) Authorize(req *http.Request) error {
	if c.Cookie == "" {
		return NewAuthError()
	}
	req.Header.Set("Cookie", c.Cookie)
	return nil
}

type GiteeClient struct {
	Url        string
	Method     string
	Payload    interface{}
	Headers    map[string]string
	Response   *http.Response
	parsedUrl  *url.URL
	Query      map[string]string
	httpClient *http.Client
	authorizer Authorizer
	apiBase    string // Added apiBase field
}

type Option func(client *GiteeClient)

func NewGiteeClient(method, urlPath string, opts ...Option) *GiteeClient {
	// Initialize client with defaults, including apiBase from GetApiBase()
	client := &GiteeClient{
		Method:     method,
		httpClient: http.DefaultClient,
		authorizer: &BearerTokenAuthorizer{},
		apiBase:    GetApiBase(),
	}

	// Apply options. This allows WithApiBase to override the default apiBase,
	// and WithQuery to populate client.Query.
	for _, opt := range opts {
		opt(client)
	}

	// Construct the full URL using the client's potentially updated apiBase
	fullURL := client.apiBase + urlPath
	parsedUrl, err := url.Parse(fullURL)
	if err != nil {
		panic(fmt.Errorf("failed to parse URL '%s': %w", fullURL, err))
	}
	client.parsedUrl = parsedUrl // Store parsed URL object

	// Apply query parameters from client.Query (populated by WithQuery option)
	if client.Query != nil {
		queryParams := client.parsedUrl.Query()
		for k, v := range client.Query {
			queryParams.Set(k, v)
		}
		client.parsedUrl.RawQuery = queryParams.Encode()
	}

	// Set the final URL string including any query parameters
	client.Url = client.parsedUrl.String()

	return client
}

func WithQuery(query map[string]interface{}) Option {
	delete(query, "enterprise_id")
	return func(client *GiteeClient) {
		parsedQuery := make(map[string]string)
		if query != nil {
			for k, v := range query {
				parsedValue := ""
				switch val := v.(type) {
				case string:
					parsedValue = val
				case int:
					parsedValue = strconv.Itoa(val)
				case float32, float64:
					parsedValue = fmt.Sprintf("%v", val)
				case bool:
					parsedValue = strconv.FormatBool(val)
				}
				if parsedValue != "" {
					parsedQuery[k] = parsedValue
				}
			}
		}
		client.Query = parsedQuery
	}
}

func WithPayload(payload interface{}) Option {
	return func(client *GiteeClient) {
		client.Payload = payload
	}
}

func WithHeaders(headers map[string]string) Option {
	return func(client *GiteeClient) {
		client.Headers = headers
	}
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(client *GiteeClient) {
		if httpClient != nil {
			client.httpClient = httpClient
		}
	}
}

func WithAuthorizer(authorizer Authorizer) Option {
	return func(client *GiteeClient) {
		if authorizer != nil {
			client.authorizer = authorizer
		}
	}
}

// WithApiBase now modifies the client's internal apiBase field
func WithApiBase(url string) Option {
	return func(client *GiteeClient) {
		if url != "" {
			client.apiBase = url
		}
	}
}

func (g *GiteeClient) SetHeaders(headers map[string]string) *GiteeClient {
	g.Headers = headers
	return g
}

func (g *GiteeClient) Do() (*GiteeClient, error) {
	g.Response = nil
	var requestBody []byte
	var err error

	if g.Payload != nil {
		requestBody, err = json.Marshal(g.Payload)
		if err != nil {
			return nil, NewInternalError(fmt.Errorf("failed to marshal payload: %w", err))
		}
	}

	req, err := http.NewRequest(g.Method, g.Url, bytes.NewReader(requestBody))
	if err != nil {
		return nil, NewInternalError(fmt.Errorf("failed to create request: %w", err))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "mcp-gitee-ent "+Version+" Go/"+runtime.GOOS+"/"+runtime.GOARCH+"/"+runtime.Version())

	for key, value := range g.Headers {
		req.Header.Set(key, value)
	}

	// Apply authorization
	if g.authorizer != nil {
		if err := g.authorizer.Authorize(req); err != nil {
			if IsAuthError(err) {
				return nil, err
			}
			return nil, NewInternalError(fmt.Errorf("authorization failed: %w", err))
		}
	} else {
		return nil, NewAuthError()
	}

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return g, NewNetworkError(err)
	}

	g.Response = resp

	// 检查响应状态码
	if !g.IsSuccess() {
		body, _ := io.ReadAll(resp.Body)
		return g, NewAPIError(resp.StatusCode, body)
	}

	return g, nil
}

func (g *GiteeClient) IsSuccess() bool {
	if g.Response == nil {
		return false
	}

	successMap := map[int]struct{}{
		http.StatusOK:          struct{}{},
		http.StatusCreated:     struct{}{},
		http.StatusNoContent:   struct{}{},
		http.StatusFound:       struct{}{},
		http.StatusNotModified: struct{}{},
	}

	if _, ok := successMap[g.Response.StatusCode]; ok {
		return true
	}
	return false
}

func (g *GiteeClient) IsFail() bool {
	return !g.IsSuccess()
}

func (g *GiteeClient) GetRespBody() ([]byte, error) {
	return io.ReadAll(g.Response.Body)
}

func (g *GiteeClient) HandleMCPResult(object any) (*mcp.CallToolResult, error) {
	_, err := g.Do()
	if err != nil {
		switch {
		case IsAuthError(err):
			return mcp.NewToolResultError("Authentication failed: Please check your Gitee Ent MCP access token"), err
		case IsNetworkError(err):
			return mcp.NewToolResultError("Network error: Unable to connect to Gitee Enterprise API"), err
		case IsAPIError(err):
			giteeErr := err.(*GiteeError)
			return mcp.NewToolResultError(fmt.Sprintf("API error (%d): %s", giteeErr.Code, giteeErr.Details)), err
		default:
			return mcp.NewToolResultError(err.Error()), err
		}
	}

	// Handle no content case when object is nil
	if object == nil {
		return mcp.NewToolResultText("Operation completed successfully"), nil
	}

	body, err := g.GetRespBody()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read response body: %s", err.Error())),
			NewInternalError(err)
	}

	if _, ok := object.(*types.PullDiff); ok {
		return mcp.NewToolResultText(string(body)), nil
	}

	if err = json.Unmarshal(body, object); err != nil {
		errorMessage := fmt.Sprintf("Failed to parse response: %v", err)
		return mcp.NewToolResultError(errorMessage), NewInternalError(errors.New(errorMessage))
	}

	result, err := json.MarshalIndent(object, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format response: %s", err.Error())),
			NewInternalError(err)
	}

	return mcp.NewToolResultText(string(result)), nil
}
