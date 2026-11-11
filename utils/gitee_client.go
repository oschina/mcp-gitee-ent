package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"time"

	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	DefaultApiBase = "https://api.gitee.com/enterprises"
)

var (
	giteeAccessToken string
	apiBase          string
	giteeHTTPClient  = &http.Client{Timeout: HTTPTimeout()}
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
	Url         string
	Method      string
	ApiPath     string
	Payload     interface{}
	Body        io.Reader
	ContentType string
	Headers     map[string]string
	Response    *http.Response
	parsedUrl   *url.URL
	Query       map[string]string
	httpClient  *http.Client
	authorizer  Authorizer
	apiBase     string // Added apiBase field
}

type Option func(client *GiteeClient)

func NewGiteeClient(method, urlPath string, opts ...Option) *GiteeClient {
	// Initialize client with defaults, including apiBase from GetApiBase()
	client := &GiteeClient{
		Method:     method,
		ApiPath:    urlPath,
		httpClient: giteeHTTPClient,
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
				if k == "per_page" {
					if limit, err := SafelyConvertToInt(v); err == nil {
						if limit < 1 {
							limit = 1
						}
						if limit > MaxPerPage() {
							limit = MaxPerPage()
						}
						v = limit
					}
				}
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

func WithBody(body io.Reader, contentType string) Option {
	return func(client *GiteeClient) {
		client.Body = body
		client.ContentType = contentType
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
	return g.DoContext(context.Background())
}

func (g *GiteeClient) DoContext(ctx context.Context) (*GiteeClient, error) {
	g.Response = nil
	var payload []byte
	var err error
	if g.Body == nil && g.Payload != nil {
		payload, err = json.Marshal(g.Payload)
		if err != nil {
			return nil, NewInternalError(fmt.Errorf("failed to marshal payload: %w", err))
		}
	}

	attempts := 1
	if g.Method == http.MethodGet {
		attempts = 2
	}
	start := time.Now()

	for attempt := 1; attempt <= attempts; attempt++ {
		var requestBody io.Reader
		if g.Body != nil {
			requestBody = g.Body
		} else if payload != nil {
			requestBody = bytes.NewReader(payload)
		}

		req, err := http.NewRequestWithContext(ctx, g.Method, g.Url, requestBody)
		if err != nil {
			return nil, NewInternalError(fmt.Errorf("failed to create request: %w", err))
		}

		contentType := g.ContentType
		if contentType == "" {
			contentType = "application/json"
		}
		req.Header.Set("Content-Type", contentType)
		req.Header.Set("User-Agent", "mcp-gitee-ent "+Version+" Go/"+runtime.GOOS+"/"+runtime.GOARCH+"/"+runtime.Version())
		for key, value := range g.Headers {
			req.Header.Set(key, value)
		}
		if g.authorizer == nil {
			return nil, NewAuthError()
		}
		if err := g.authorizer.Authorize(req); err != nil {
			if IsAuthError(err) {
				return nil, err
			}
			return nil, NewInternalError(fmt.Errorf("authorization failed: %w", err))
		}

		resp, err := g.httpClient.Do(req)
		duration := time.Since(start)
		if err != nil {
			if attempt < attempts && ctx.Err() == nil {
				log.Printf("gitee request retry method=%s path=%s attempt=%d error=%v", g.Method, g.parsedUrl.Path, attempt, err)
				time.Sleep(200 * time.Millisecond)
				continue
			}
			log.Printf("gitee request failed method=%s path=%s duration=%s error=%v", g.Method, g.parsedUrl.Path, duration, err)
			return g, NewNetworkError(err)
		}
		g.Response = resp

		if !g.IsSuccess() {
			body, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if attempt < attempts && (resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500) {
				log.Printf("gitee request retry method=%s path=%s status=%d attempt=%d", g.Method, g.parsedUrl.Path, resp.StatusCode, attempt)
				time.Sleep(200 * time.Millisecond)
				continue
			}
			log.Printf("gitee request error method=%s path=%s status=%d duration=%s", g.Method, g.parsedUrl.Path, resp.StatusCode, duration)
			return g, NewAPIError(resp.StatusCode, body)
		}

		if duration > SlowRequestThreshold() {
			log.Printf("gitee request slow method=%s path=%s status=%d duration=%s", g.Method, g.parsedUrl.Path, resp.StatusCode, duration)
		}
		return g, nil
	}
	return g, NewInternalError(fmt.Errorf("request failed after retries"))
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
	return g.HandleMCPResultContext(context.Background(), object)
}

func (g *GiteeClient) HandleMCPResultContext(ctx context.Context, object any) (*mcp.CallToolResult, error) {
	_, err := g.DoContext(ctx)
	if err != nil {
		return MCPErrorResult(err)
	}
	defer g.Response.Body.Close()

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
		if len(body) > MaxResponseBytes() {
			return mcp.NewToolResultError(fmt.Sprintf("Response too large (%d bytes, limit %d). Narrow the diff or reduce per_page.", len(body), MaxResponseBytes())), nil
		}
		return mcp.NewToolResultText(string(body)), nil
	}

	if err = json.Unmarshal(body, object); err != nil {
		errorMessage := fmt.Sprintf("Failed to parse response: %v", err)
		return mcp.NewToolResultError(errorMessage), NewInternalError(errors.New(errorMessage))
	}

	return MCPJSONResult(EnrichResult(g.ApiPath, object))
}

func MCPErrorResult(err error) (*mcp.CallToolResult, error) {
	switch {
	case IsAuthError(err):
		return mcp.NewToolResultError("Authentication failed: Please check your Gitee Ent MCP access token"), err
	case IsNetworkError(err):
		return mcp.NewToolResultError("Network error: Unable to connect to Gitee Enterprise API"), err
	case IsAPIError(err):
		giteeErr := err.(*GiteeError)
		hint := ""
		switch {
		case giteeErr.Code == http.StatusUnauthorized:
			hint = " Check the access token."
		case giteeErr.Code == http.StatusForbidden:
			hint = " Check token scopes, enterprise permissions, and whether the enterprise plan includes this feature."
		case giteeErr.Code == http.StatusNotFound:
			hint = " Check the resource IDs and project scope."
		case giteeErr.Code == http.StatusTooManyRequests:
			hint = " Rate limited; retry later or reduce request frequency."
		case giteeErr.Code >= 500:
			hint = " Upstream error; retry later."
		}
		return mcp.NewToolResultError(fmt.Sprintf("API error (%d): %s%s", giteeErr.Code, giteeErr.Details, hint)), err
	default:
		return mcp.NewToolResultError(err.Error()), err
	}
}

func MCPJSONResult(object any) (*mcp.CallToolResult, error) {
	result, err := json.Marshal(object)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format response: %s", err.Error())),
			NewInternalError(err)
	}

	if len(result) > MaxResponseBytes() {
		return mcp.NewToolResultError(fmt.Sprintf("Response too large (%d bytes, limit %d). Narrow filters or reduce per_page.", len(result), MaxResponseBytes())), nil
	}
	return mcp.NewToolResultText(string(result)), nil
}
