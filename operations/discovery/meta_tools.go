// Package discovery implements the progressive-discovery meta tools:
// search_tools, inspect_tool and invoke_tool.
package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"gitee.com/oschina/mcp-gitee-ent/catalog"
	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	SearchTools = "search_tools"
	InspectTool = "inspect_tool"
	InvokeTool  = "invoke_tool"

	defaultSearchLimit = 8
	maxSearchLimit     = 20
)

var SearchDomains = []string{
	"issues",
	"issue_links",
	"pulls",
	"repository",
	"documents",
	"milestones",
	"members",
	"groups",
	"workloads",
	"programs",
	"test_modules",
	"test_cases",
	"test_plans",
	"test_reports",
	"check_runs",
	"webhooks",
	"baselines",
	"views",
	"kanbans",
	"issue_recycle_bin",
	"labels",
	"issue_types",
	"issue_states",
	"scrum_sprints",
	"scrum_versions",
	"enterprises",
	"user",
}

// Hit is one search result. Only names are returned so that callers cannot
// fabricate executable handlers.
type Hit struct {
	Name   string
	Score  float64
	Source string
}

// Searcher retrieves candidate tool names for a query.
type Searcher interface {
	Search(ctx context.Context, query string, limit int) ([]Hit, error)
}

type ToolFilter interface {
	Allowed(ctx context.Context, spec catalog.Spec) bool
}

type ToolFilterFunc func(context.Context, catalog.Spec) bool

func (f ToolFilterFunc) Allowed(ctx context.Context, spec catalog.Spec) bool {
	return f(ctx, spec)
}

// SearchToolsTool discovers tools in the catalog.
var SearchToolsTool = mcp.NewTool(SearchTools,
	mcp.WithDescription("Discover available Gitee Enterprise tools by natural-language query. "+
		"Call this FIRST whenever the user asks for a capability you do not see a direct tool for "+
		"(for example milestones, work item relations, members, workloads, branches, check runs). "+
		"Before saying a capability is unavailable, search here. "+
		"Results contain tool names, short descriptions and required argument names; "+
		"use inspect_tool for the full schema and invoke_tool to execute."),
	mcp.WithString("query",
		mcp.Required(),
		mcp.Description("What the user wants to do, in natural language. Include domain words when known (issue, pull request, milestone, member, workload)."),
	),
	mcp.WithString("domain",
		mcp.Description("Optional exact domain filter. Allowed values: "+strings.Join(SearchDomains, ", ")+"."),
		mcp.Enum(SearchDomains...),
	),
	mcp.WithNumber("limit",
		mcp.Description("Maximum number of tools to return (default 8, max 20)."),
		mcp.Min(1),
		mcp.Max(maxSearchLimit),
	),
	mcp.WithReadOnlyHintAnnotation(true),
)

// InspectToolTool shows the full schema of one tool.
var InspectToolTool = mcp.NewTool(InspectTool,
	mcp.WithDescription("Show the full argument schema and a valid example for one tool. "+
		"Use after search_tools, before invoke_tool when argument names or formats are unclear."),
	mcp.WithString("name",
		mcp.Required(),
		mcp.Description("Exact tool name returned by search_tools."),
	),
	mcp.WithReadOnlyHintAnnotation(true),
)

// InvokeToolTool executes one catalog tool by exact name.
var InvokeToolTool = mcp.NewTool(InvokeTool,
	mcp.WithDescription("Execute one Gitee Enterprise tool by exact name. "+
		"Only tools returned by search_tools and enabled by the server configuration can be invoked. "+
		"Pass the tool arguments as a JSON object in the arguments field. "+
		"Do not use this to call invoke_tool itself."),
	mcp.WithString("name",
		mcp.Required(),
		mcp.Description("Exact tool name returned by search_tools."),
	),
	mcp.WithObject("arguments",
		mcp.Description("Arguments for the target tool, matching its input schema. May also be supplied as a JSON-encoded string."),
	),
	mcp.WithReadOnlyHintAnnotation(false),
	mcp.WithOpenWorldHintAnnotation(true),
)

// SearchToolsHandler handles search_tools.
type SearchToolsHandler struct {
	registry *catalog.Registry
	searcher Searcher
	filter   ToolFilter
}

// NewSearchToolsHandler creates a search handler. A nil searcher uses the
// built-in lexical search.
func NewSearchToolsHandler(registry *catalog.Registry, searcher Searcher) *SearchToolsHandler {
	if searcher == nil {
		searcher = NewLexicalSearcher(registry)
	}
	return &SearchToolsHandler{registry: registry, searcher: searcher}
}

func (h *SearchToolsHandler) WithFilter(filter ToolFilter) *SearchToolsHandler {
	h.filter = filter
	return h
}

func (h *SearchToolsHandler) eligible(ctx context.Context, spec catalog.Spec, domain string) bool {
	if !spec.Invokable {
		return false
	}
	if h.filter != nil && !h.filter.Allowed(ctx, spec) {
		return false
	}
	return domain == "" || strings.EqualFold(spec.Domain, domain)
}

func (h *SearchToolsHandler) lexicalBackfill(ctx context.Context, query, domain string) []Hit {
	normalizedQuery := normalizeText(strings.ToLower(strings.TrimSpace(query)))
	if normalizedQuery == "" {
		return nil
	}
	compactQuery := strings.ReplaceAll(normalizedQuery, " ", "")
	tokens := strings.Fields(normalizedQuery)

	hits := make([]Hit, 0)
	for _, spec := range h.registry.List() {
		if !h.eligible(ctx, spec, domain) {
			continue
		}
		score := scoreSpec(spec, normalizedQuery, compactQuery, tokens)
		if score < scoreAliasContains {
			continue
		}
		hits = append(hits, Hit{Name: spec.Tool.Name, Score: score, Source: "lexical"})
	}
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].Score == hits[j].Score {
			return hits[i].Name < hits[j].Name
		}
		return hits[i].Score > hits[j].Score
	})
	return hits
}

// Handle implements utils.OptionHandlerFunc.
func (h *SearchToolsHandler) Handle(ctx context.Context, request mcp.CallToolRequest, _ ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	query, err := requiredString(arguments, "query")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	domain, _ := optionalString(arguments, "domain")
	limit := optionalInt(arguments, "limit", defaultSearchLimit)
	if limit < 1 {
		limit = 1
	}
	if limit > maxSearchLimit {
		limit = maxSearchLimit
	}

	hits, err := h.searcher.Search(ctx, query, searchCandidateLimit(limit, domain))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("tool search failed: %v", err)), nil
	}

	seen := make(map[string]bool)
	selected := make([]catalog.Spec, 0, limit)
	appendCandidate := func(name string) {
		if len(selected) >= limit || seen[name] {
			return
		}
		spec, ok := h.registry.Get(name)
		if !ok || !h.eligible(ctx, spec, domain) {
			return
		}
		seen[name] = true
		selected = append(selected, spec)
	}
	for _, hit := range hits {
		appendCandidate(hit.Name)
	}
	if len(selected) == 0 {
		for _, hit := range h.lexicalBackfill(ctx, query, domain) {
			appendCandidate(hit.Name)
		}
	}

	if len(selected) == 0 {
		message := fmt.Sprintf("No tools matched %q.", query)
		if domain != "" {
			message += fmt.Sprintf(" Try removing the domain filter %q.", domain)
		}
		message += " Try broader keywords, or inspect an existing tool name with inspect_tool."
		return mcp.NewToolResultText(message), nil
	}

	var builder strings.Builder
	fmt.Fprintf(&builder, "Found %d tool(s) matching %q:\n\n", len(selected), query)
	for i, spec := range selected {
		fmt.Fprintf(&builder, "%d. %s\n", i+1, spec.Tool.Name)
		fmt.Fprintf(&builder, "   %s\n", spec.Summary)
		fmt.Fprintf(&builder, "   domain: %s | risk: %s\n", spec.Domain, spec.Risk)
		if required := spec.RequiredArguments(); len(required) > 0 {
			fmt.Fprintf(&builder, "   required: %s\n", strings.Join(required, ", "))
		}
		builder.WriteString("\n")
	}
	builder.WriteString("Use inspect_tool(name) for the full schema and an example.\n")
	builder.WriteString("Use invoke_tool(name, arguments) to execute.")
	return mcp.NewToolResultText(builder.String()), nil
}

// InspectToolHandler handles inspect_tool.
type InspectToolHandler struct {
	registry *catalog.Registry
	filter   ToolFilter
}

// NewInspectToolHandler creates an inspect handler.
func NewInspectToolHandler(registry *catalog.Registry) *InspectToolHandler {
	return &InspectToolHandler{registry: registry}
}

func (h *InspectToolHandler) WithFilter(filter ToolFilter) *InspectToolHandler {
	h.filter = filter
	return h
}

type inspectResponse struct {
	Name             string         `json:"name"`
	Summary          string         `json:"summary"`
	Description      string         `json:"description,omitempty"`
	Domain           string         `json:"domain"`
	Risk             catalog.Risk   `json:"risk"`
	Required         []string       `json:"required,omitempty"`
	InputSchema      any            `json:"input_schema"`
	ExampleArguments map[string]any `json:"example_arguments,omitempty"`
	Notes            []string       `json:"notes,omitempty"`
	Invokable        bool           `json:"invokable"`
}

// Handle implements utils.OptionHandlerFunc.
func (h *InspectToolHandler) Handle(ctx context.Context, request mcp.CallToolRequest, _ ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := requiredString(arguments, "name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	spec, ok := h.registry.Get(name)
	if !ok {
		return mcp.NewToolResultError(unknownToolMessage(ctx, h.registry, name, h.filter)), nil
	}
	if !spec.Invokable {
		return mcp.NewToolResultError(toolNotEnabledMessage(name)), nil
	}
	if h.filter != nil && !h.filter.Allowed(ctx, spec) {
		return mcp.NewToolResultError(toolNotAllowedMessage(name)), nil
	}

	response := inspectResponse{
		Name:             spec.Tool.Name,
		Summary:          spec.Summary,
		Description:      spec.Tool.Description,
		Domain:           spec.Domain,
		Risk:             spec.Risk,
		Required:         spec.RequiredArguments(),
		InputSchema:      spec.Tool.InputSchema,
		ExampleArguments: spec.ExampleArguments,
		Notes:            spec.Notes,
		Invokable:        spec.Invokable,
	}
	encoded, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to encode tool schema: %v", err)), nil
	}
	return mcp.NewToolResultText(string(encoded)), nil
}

// InvokeToolHandler handles invoke_tool.
type InvokeToolHandler struct {
	registry *catalog.Registry
	filter   ToolFilter
}

// NewInvokeToolHandler creates an invoke handler.
func NewInvokeToolHandler(registry *catalog.Registry) *InvokeToolHandler {
	return &InvokeToolHandler{registry: registry}
}

func (h *InvokeToolHandler) WithFilter(filter ToolFilter) *InvokeToolHandler {
	h.filter = filter
	return h
}

// Handle implements utils.OptionHandlerFunc.
func (h *InvokeToolHandler) Handle(ctx context.Context, request mcp.CallToolRequest, _ ...utils.Option) (result *mcp.CallToolResult, err error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := requiredString(arguments, "name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if name == InvokeTool {
		return mcp.NewToolResultError("invoke_tool cannot call itself; specify a concrete tool name."), nil
	}

	spec, ok := h.registry.Get(name)
	if !ok {
		return mcp.NewToolResultError(unknownToolMessage(ctx, h.registry, name, h.filter)), nil
	}
	if !spec.Invokable {
		return mcp.NewToolResultError(toolNotEnabledMessage(name)), nil
	}
	if h.filter != nil && !h.filter.Allowed(ctx, spec) {
		return mcp.NewToolResultError(toolNotAllowedMessage(name)), nil
	}

	toolArguments, err := decodeArguments(arguments["arguments"])
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid arguments for %q: %v", name, err)), nil
	}

	inner := mcp.CallToolRequest{}
	inner.Params.Name = name
	inner.Params.Arguments = toolArguments

	start := time.Now()
	defer func() {
		if recovered := recover(); recovered != nil {
			log.Printf("invoke_tool panic name=%s duration=%s panic=%v", name, time.Since(start), recovered)
			result = mcp.NewToolResultError(fmt.Sprintf("[%s] internal error while executing tool: %v", name, recovered))
			err = nil
		}
		log.Printf("invoke_tool name=%s duration=%s is_error=%t", name, time.Since(start), result != nil && result.IsError)
	}()

	result, err = spec.Handler(ctx, inner, spec.Options...)
	if err != nil {
		log.Printf("invoke_tool name=%s error=%v", name, err)
		if result == nil {
			return mcp.NewToolResultError(fmt.Sprintf("[%s] %v", name, err)), nil
		}
	}
	if result == nil {
		return mcp.NewToolResultError(fmt.Sprintf("[%s] tool returned no result", name)), nil
	}
	if result.IsError {
		return prefixToolError(result, name), nil
	}
	return result, nil
}

func prefixToolError(result *mcp.CallToolResult, name string) *mcp.CallToolResult {
	var messages []string
	for _, content := range result.Content {
		if text, ok := content.(mcp.TextContent); ok {
			messages = append(messages, text.Text)
		}
	}
	if len(messages) == 0 {
		return mcp.NewToolResultError(fmt.Sprintf("[%s] tool returned an error", name))
	}
	return mcp.NewToolResultError(fmt.Sprintf("[%s] %s", name, strings.Join(messages, "\n")))
}

func searchCandidateLimit(limit int, domain string) int {
	if domain != "" {
		return 50
	}
	return limit * 2
}

func requiredString(arguments map[string]any, key string) (string, error) {
	value, ok := arguments[key]
	if !ok || value == nil {
		return "", fmt.Errorf("Missing required parameter: %s. Use inspect_tool to see the expected arguments.", key)
	}
	text, ok := value.(string)
	if !ok || strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("Parameter %q must be a non-empty string.", key)
	}
	return strings.TrimSpace(text), nil
}

func optionalString(arguments map[string]any, key string) (string, bool) {
	value, ok := arguments[key]
	if !ok || value == nil {
		return "", false
	}
	text, ok := value.(string)
	if !ok {
		return "", false
	}
	return strings.TrimSpace(text), true
}

func optionalInt(arguments map[string]any, key string, fallback int) int {
	value, ok := arguments[key]
	if !ok || value == nil {
		return fallback
	}
	converted, err := utils.SafelyConvertToInt(value)
	if err != nil {
		log.Printf("ignoring invalid %s=%v: %v", key, value, err)
		return fallback
	}
	return converted
}

func decodeArguments(value any) (map[string]any, error) {
	if value == nil {
		return map[string]any{}, nil
	}
	switch typed := value.(type) {
	case map[string]any:
		return typed, nil
	case string:
		trimmed := strings.TrimSpace(typed)
		if trimmed == "" {
			return map[string]any{}, nil
		}
		decoded := map[string]any{}
		if err := json.Unmarshal([]byte(trimmed), &decoded); err != nil {
			return nil, fmt.Errorf("arguments must be a JSON object: %w", err)
		}
		return decoded, nil
	default:
		return nil, fmt.Errorf("arguments must be an object or a JSON object string, got %T", value)
	}
}

func toolNotEnabledMessage(name string) string {
	return fmt.Sprintf(
		"Tool %q is registered but not enabled in this server configuration. "+
			"Enable it with ENABLED_TOOLSETS or remove it from DISABLED_TOOLSETS.", name)
}

func toolNotAllowedMessage(name string) string {
	return fmt.Sprintf("Tool %q is not allowed for this request.", name)
}

func unknownToolMessage(ctx context.Context, registry *catalog.Registry, name string, filter ToolFilter) string {
	message := fmt.Sprintf("Unknown tool %q.", name)
	suggestions := suggestNames(ctx, registry, name, 3, filter)
	if len(suggestions) > 0 {
		message += " Did you mean: " + strings.Join(suggestions, ", ") + "?"
	}
	message += " Use search_tools to discover available tools, then inspect_tool for arguments."
	return message
}

func suggestNames(ctx context.Context, registry *catalog.Registry, query string, limit int, filter ToolFilter) []string {
	candidateLimit := limit * 4
	if candidateLimit < 8 {
		candidateLimit = 8
	}
	hits, err := NewLexicalSearcher(registry).Search(ctx, query, candidateLimit)
	if err != nil {
		return nil
	}
	names := make([]string, 0, limit)
	for _, hit := range hits {
		if len(names) >= limit {
			break
		}
		spec, ok := registry.Get(hit.Name)
		if !ok || !spec.Invokable {
			continue
		}
		if filter != nil && !filter.Allowed(ctx, spec) {
			continue
		}
		names = append(names, hit.Name)
	}
	return names
}
