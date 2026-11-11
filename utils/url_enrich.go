package utils

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sync"
	"sync/atomic"

	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"gitee.com/oschina/mcp-gitee-ent/utils/urlbuilder"
	"github.com/mark3labs/mcp-go/mcp"
)

var entPathCache sync.Map

var (
	entPathLocks       sync.Map
	autoResolveEntPath atomic.Bool
)

// EnableEnterprisePathAutoResolve makes EnterprisePath fetch an unknown
// enterprise on demand. Enabled by main so servers always render html_url even
// when the startup warm-up failed; disabled by default to keep unit tests off
// the network.
func EnableEnterprisePathAutoResolve(enabled bool) {
	autoResolveEntPath.Store(enabled)
}

// EnterprisePath returns the URL path for an enterprise ID. Known paths come
// from the cache, which is primed by WarmEnterprisePaths and refreshed by
// list_enterprises. Unknown IDs are resolved on demand when auto resolve is on.
func EnterprisePath(enterpriseID int) string {
	if enterpriseID <= 0 {
		return ""
	}
	if cached, ok := entPathCache.Load(enterpriseID); ok {
		value, _ := cached.(string)
		return value
	}
	if !autoResolveEntPath.Load() {
		return ""
	}
	return resolveEnterprisePath(enterpriseID)
}

func resolveEnterprisePath(enterpriseID int) string {
	lockValue, _ := entPathLocks.LoadOrStore(enterpriseID, &sync.Mutex{})
	lock, _ := lockValue.(*sync.Mutex)
	lock.Lock()
	defer lock.Unlock()

	if cached, ok := entPathCache.Load(enterpriseID); ok {
		value, _ := cached.(string)
		return value
	}
	path, err := fetchEnterprisePath(enterpriseID)
	if err != nil || path == "" {
		return ""
	}
	entPathCache.Store(enterpriseID, path)
	return path
}

func fetchEnterprisePath(enterpriseID int) (string, error) {
	client := NewGiteeClient("GET", fmt.Sprintf("/%d", enterpriseID))
	if _, err := client.Do(); err != nil {
		return "", err
	}
	defer client.Response.Body.Close()

	body, err := client.GetRespBody()
	if err != nil {
		return "", err
	}
	var enterprise struct {
		Path string `json:"path"`
	}
	if err = json.Unmarshal(body, &enterprise); err != nil {
		return "", err
	}
	return enterprise.Path, nil
}

// SetEnterprisePath records an enterprise ID to path mapping.
func SetEnterprisePath(enterpriseID int, path string) {
	if enterpriseID <= 0 || path == "" {
		return
	}
	entPathCache.Store(enterpriseID, path)
}

// CacheEnterprises records the ID to path mapping of the given enterprises.
func CacheEnterprises(enterprises []types.BasicEnterprise) {
	for _, enterprise := range enterprises {
		SetEnterprisePath(enterprise.Id, enterprise.Path)
	}
}

// WarmEnterprisePaths loads the caller's enterprises so html_url can be built
// without an extra request on every tool call.
func WarmEnterprisePaths() {
	client := NewGiteeClient("GET", "/list", WithQuery(map[string]interface{}{"per_page": 100}))
	if _, err := client.Do(); err != nil {
		return
	}
	defer client.Response.Body.Close()

	body, err := client.GetRespBody()
	if err != nil {
		return
	}
	var payload types.PagedResponse[types.BasicEnterprise]
	if err = json.Unmarshal(body, &payload); err != nil {
		return
	}
	CacheEnterprises(payload.Data)
}

// ResetEnterprisePathCache clears the resolver cache. Intended for tests.
func ResetEnterprisePathCache() {
	entPathCache.Range(func(key, _ any) bool {
		entPathCache.Delete(key)
		return true
	})
}

type enrichContext struct {
	builder urlbuilder.Builder
}

type enrichRule struct {
	pattern *regexp.Regexp
	apply   func(ctx enrichContext, root any, groups []string)
}

var enrichRules = []enrichRule{
	{mustRe(`^/(\d+)/issues$`), enrichIssueList},
	{mustRe(`^/(\d+)/issues/([^/]+)$`), enrichIssueDetail},
	{mustRe(`^/(\d+)/pull_requests$`), enrichPullList},
	{mustRe(`^/(\d+)/projects/([^/]+)/pull_requests$`), enrichPullList},
	{mustRe(`^/(\d+)/projects/([^/]+)/pull_requests/(\d+)$`), enrichPullDetail},
	{mustRe(`^/(\d+)/projects$`), enrichRepoList},
	{mustRe(`^/(\d+)/projects/([^/]+)/tree/.+$`), enrichRepoTree},
	{mustRe(`^/(\d+)/projects/([^/]+)/releases$`), enrichReleaseList},
	{mustRe(`^/(\d+)/programs$`), enrichProgramList},
	{mustRe(`^/(\d+)/programs/(\d+)$`), enrichProgramDetail},
	{mustRe(`^/(\d+)/programs/(\d+)/scrum_sprints$`), enrichSprintList},
	{mustRe(`^/(\d+)/programs/(\d+)/scrum_sprints/(\d+)$`), enrichSprintDetail},
	{mustRe(`^/(\d+)/programs/(\d+)/scrum_versions$`), enrichVersionList},
	{mustRe(`^/(\d+)/programs/(\d+)/scrum_versions/(\d+)$`), enrichVersionDetail},
	{mustRe(`^/(\d+)/programs/(\d+)/baselines$`), enrichBaselineList},
	{mustRe(`^/(\d+)/programs/(\d+)/baselines/(\d+)$`), enrichBaselineDetail},
	{mustRe(`^/(\d+)/milestones$`), enrichMilestoneList},
	{mustRe(`^/(\d+)/milestones/(\d+)$`), enrichMilestoneDetail},
	{mustRe(`^/(\d+)/test_plans$`), enrichTestPlanList},
	{mustRe(`^/(\d+)/test_plans/(\d+)$`), enrichTestPlanDetail},
	{mustRe(`^/(\d+)/test_cases$`), enrichTestCaseList},
	{mustRe(`^/(\d+)/test_cases/(\d+)$`), enrichTestCaseDetail},
	{mustRe(`^/(\d+)/doc_nodes/tile$`), enrichDocumentList},
	{mustRe(`^/(\d+)/doc_nodes/create$`), enrichDocumentList},
	{mustRe(`^/(\d+)/doc_nodes/(recent_doc|shared)$`), enrichDocumentList},
	{mustRe(`^/(\d+)/doc_nodes/(\d+)$`), enrichDocumentDetail},
	{mustRe(`^/(\d+)/groups$`), enrichGroupList},
	{mustRe(`^/(\d+)/groups/([^/]+)$`), enrichGroupDetail},
	{mustRe(`^/(\d+)/groups/[^/]+/projects$`), enrichRepoList},
	{mustRe(`^/(\d+)/groups/[^/]+/group_relate_programs$`), enrichProgramList},
	{mustRe(`^/(\d+)/programs/(\d+)/(overview|components)$`), enrichProgramSubPage},
	{mustRe(`^/(\d+)/issues/[^/]+/(children|ancestors|link_issues)$`), enrichIssueLinks},
	{mustRe(`^/(\d+)/issues/([^/]+)/overview$`), enrichIssueOverview},
	{mustRe(`^/(\d+)/kanbans$`), enrichKanbanList},
	{mustRe(`^/(\d+)/test_plans/(\d+)/test_cases$`), enrichTestPlanCases},
	{mustRe(`^/(\d+)/test_plans/(\d+)/test_cases/(\d+)$`), enrichTestPlanCases},
}

// EnrichJSONResult enriches the payload and renders it as an MCP text result.
// Operations that build their own response payload should use this instead of
// MCPJSONResult so html_url is added consistently.
func EnrichJSONResult(apiPath string, object any) (*mcp.CallToolResult, error) {
	return MCPJSONResult(EnrichResult(apiPath, object))
}

// EnrichResult adds html_url fields to a tool result based on the API path that
// produced it. Responses without a matching rule are returned unchanged.
func EnrichResult(apiPath string, object any) any {
	if apiPath == "" || object == nil || len(enrichRules) == 0 {
		return object
	}

	var rule *enrichRule
	var groups []string
	for i := range enrichRules {
		if match := enrichRules[i].pattern.FindStringSubmatch(apiPath); match != nil {
			rule = &enrichRules[i]
			groups = match
			break
		}
	}
	if rule == nil {
		return object
	}

	entPath := EnterprisePath(atoiSafe(groups[1]))
	if entPath == "" {
		return object
	}

	raw, err := json.Marshal(object)
	if err != nil {
		return object
	}
	var generic any
	if err = json.Unmarshal(raw, &generic); err != nil {
		return object
	}

	rule.apply(enrichContext{builder: urlbuilder.New(entPath)}, generic, groups)
	return generic
}

func enrichIssueList(ctx enrichContext, root any, _ []string) {
	eachItem(root, func(item map[string]any) {
		setHTMLURL(item, ctx.builder.Issue(stringField(item, "ident")))
	})
}

func enrichIssueDetail(ctx enrichContext, root any, groups []string) {
	item := singleObject(root)
	ident := stringField(item, "ident")
	if ident == "" {
		ident = group(groups, 2)
	}
	setHTMLURL(item, ctx.builder.Issue(ident))
}

func enrichPullList(ctx enrichContext, root any, groups []string) {
	fallback := pathParam(group(groups, 2))
	eachItem(root, func(item map[string]any) {
		pathWithNamespace := repoNamespace(item, fallback)
		setHTMLURL(item, ctx.builder.PullRequest(pathWithNamespace, intField(item, "iid")))
	})
}

func enrichPullDetail(ctx enrichContext, root any, groups []string) {
	item := singleObject(root)
	pathWithNamespace := repoNamespace(item, pathParam(group(groups, 2)))
	setHTMLURL(item, ctx.builder.PullRequest(pathWithNamespace, intField(item, "iid")))
}

func enrichRepoList(ctx enrichContext, root any, _ []string) {
	eachItem(root, func(item map[string]any) {
		setHTMLURL(item, ctx.builder.Repository(stringField(item, "path_with_namespace")))
	})
}

func enrichRepoTree(ctx enrichContext, root any, _ []string) {
	tree := mapField(singleObject(root), "tree")
	for _, key := range []string{"folders", "files"} {
		for _, entry := range sliceField(tree, key) {
			item := asMap(entry)
			setHTMLURL(item, ctx.builder.RepoWebPath(stringField(item, "path")))
		}
	}
}

func enrichReleaseList(ctx enrichContext, root any, groups []string) {
	fallback := pathParam(group(groups, 2))
	eachItem(root, func(item map[string]any) {
		pathWithNamespace := repoNamespace(item, fallback)
		setHTMLURL(item, ctx.builder.Release(pathWithNamespace, stringField(item, "tag_version")))
	})
}

func enrichProgramList(ctx enrichContext, root any, _ []string) {
	eachItem(root, func(item map[string]any) {
		setHTMLURL(item, ctx.builder.ProgramOverview(intField(item, "id")))
	})
}

func enrichProgramDetail(ctx enrichContext, root any, groups []string) {
	item := singleObject(root)
	programID := firstPositive(intField(item, "id"), atoiSafe(group(groups, 2)))
	setHTMLURL(item, ctx.builder.ProgramOverview(programID))
}

func enrichSprintList(ctx enrichContext, root any, groups []string) {
	programID := atoiSafe(group(groups, 2))
	eachItem(root, func(item map[string]any) {
		sprintProgramID := firstPositive(intField(item, "program_id"), programID)
		setHTMLURL(item, ctx.builder.SprintOverview(sprintProgramID, intField(item, "id")))
	})
}

func enrichSprintDetail(ctx enrichContext, root any, groups []string) {
	item := singleObject(root)
	setHTMLURL(item, ctx.builder.SprintOverview(
		firstPositive(intField(item, "program_id"), atoiSafe(group(groups, 2))),
		firstPositive(intField(item, "id"), atoiSafe(group(groups, 3))),
	))
}

func enrichVersionList(ctx enrichContext, root any, groups []string) {
	programID := atoiSafe(group(groups, 2))
	eachItem(root, func(item map[string]any) {
		versionProgramID := firstPositive(intField(item, "program_id"), programID)
		setHTMLURL(item, ctx.builder.Version(versionProgramID, intField(item, "id")))
	})
}

func enrichVersionDetail(ctx enrichContext, root any, groups []string) {
	item := singleObject(root)
	setHTMLURL(item, ctx.builder.Version(
		firstPositive(intField(item, "program_id"), atoiSafe(group(groups, 2))),
		firstPositive(intField(item, "id"), atoiSafe(group(groups, 3))),
	))
}

func enrichBaselineList(ctx enrichContext, root any, groups []string) {
	programID := atoiSafe(group(groups, 2))
	eachItem(root, func(item map[string]any) {
		baselineProgramID := firstPositive(intField(item, "program_id"), programID)
		setHTMLURL(item, ctx.builder.Baseline(baselineProgramID, intField(item, "id")))
	})
}

func enrichBaselineDetail(ctx enrichContext, root any, groups []string) {
	item := singleObject(root)
	setHTMLURL(item, ctx.builder.Baseline(
		firstPositive(intField(item, "program_id"), atoiSafe(group(groups, 2))),
		firstPositive(intField(item, "id"), atoiSafe(group(groups, 3))),
	))
}

func enrichMilestoneList(ctx enrichContext, root any, _ []string) {
	eachItem(root, func(item map[string]any) {
		setHTMLURL(item, milestoneURL(ctx, item, intField(item, "id")))
	})
}

func enrichMilestoneDetail(ctx enrichContext, root any, groups []string) {
	item := singleObject(root)
	setHTMLURL(item, milestoneURL(ctx, item, firstPositive(intField(item, "id"), atoiSafe(group(groups, 2)))))
}

func milestoneURL(ctx enrichContext, item map[string]any, milestoneID int) string {
	if pathWithNamespace := firstRepoPath(sliceField(item, "projects")); pathWithNamespace != "" {
		return ctx.builder.MilestoneRepo(pathWithNamespace, milestoneID)
	}
	if programID := intField(item, "program_id"); programID > 0 {
		return ctx.builder.MilestoneProgram(programID, milestoneID)
	}
	return ctx.builder.MilestoneEnterprise(milestoneID)
}

func enrichTestPlanList(ctx enrichContext, root any, _ []string) {
	eachItem(root, func(item map[string]any) {
		setHTMLURL(item, ctx.builder.TestPlan(intField(item, "program_id"), intField(item, "id")))
	})
}

func enrichTestPlanDetail(ctx enrichContext, root any, groups []string) {
	item := singleObject(root)
	setHTMLURL(item, ctx.builder.TestPlan(
		firstPositive(intField(item, "program_id"), atoiSafe(group(groups, 1))),
		firstPositive(intField(item, "id"), atoiSafe(group(groups, 2))),
	))
}

func enrichTestCaseList(ctx enrichContext, root any, _ []string) {
	eachItem(root, func(item map[string]any) {
		setHTMLURL(item, ctx.builder.TestCase(intField(item, "program_id"), testCaseIdent(item)))
	})
}

func enrichTestCaseDetail(ctx enrichContext, root any, groups []string) {
	item := singleObject(root)
	ident := testCaseIdent(item)
	if ident == "" {
		ident = group(groups, 2)
	}
	setHTMLURL(item, ctx.builder.TestCase(
		firstPositive(intField(item, "program_id"), atoiSafe(group(groups, 1))),
		ident,
	))
}

func enrichDocumentList(ctx enrichContext, root any, _ []string) {
	eachItem(root, func(item map[string]any) {
		setHTMLURL(item, documentURL(ctx, item))
	})
}

func enrichDocumentDetail(ctx enrichContext, root any, _ []string) {
	item := singleObject(root)
	setHTMLURL(item, documentURL(ctx, item))
}

func documentURL(ctx enrichContext, item map[string]any) string {
	docNodeID := firstPositive(intField(item, "document_node_id"), intField(item, "id"))
	fileID := firstPositive(intField(item, "document_id"), intField(item, "file_id"))
	programID := intField(mapField(item, "program"), "id")
	return ctx.builder.Document(programID, docNodeID, fileID)
}

func enrichGroupDetail(ctx enrichContext, root any, groups []string) {
	item := singleObject(root)
	groupPath := stringField(item, "path")
	if groupPath == "" {
		groupPath = group(groups, 2)
	}
	setHTMLURL(item, ctx.builder.Group(groupPath))
}

func enrichProgramSubPage(ctx enrichContext, root any, groups []string) {
	item := singleObject(root)
	setHTMLURL(item, ctx.builder.ProgramOverview(atoiSafe(group(groups, 2))))
}

func enrichIssueOverview(ctx enrichContext, root any, groups []string) {
	item := singleObject(root)
	ident := stringField(item, "ident")
	if ident == "" {
		ident = pathParam(group(groups, 2))
	}
	setHTMLURL(item, ctx.builder.Issue(ident))
}

func enrichIssueLinks(ctx enrichContext, root any, _ []string) {
	eachItem(root, func(item map[string]any) {
		setHTMLURL(item, ctx.builder.Issue(stringField(item, "ident")))
	})
}

func enrichKanbanList(ctx enrichContext, root any, _ []string) {
	eachItem(root, func(item map[string]any) {
		setHTMLURL(item, ctx.builder.Kanban(intField(item, "program_id")))
	})
}

func enrichTestPlanCases(ctx enrichContext, root any, groups []string) {
	planID := atoiSafe(group(groups, 2))
	eachItem(root, func(item map[string]any) {
		setHTMLURL(item, ctx.builder.TestPlanCase(
			intField(item, "program_id"),
			planID,
			testCaseIdent(item),
		))
	})
}

func enrichGroupList(ctx enrichContext, root any, _ []string) {
	eachItem(root, func(item map[string]any) {
		setHTMLURL(item, ctx.builder.Group(stringField(item, "path")))
	})
}

func eachItem(root any, fn func(map[string]any)) {
	switch value := root.(type) {
	case []any:
		for _, entry := range value {
			if item := asMap(entry); item != nil {
				fn(item)
			}
		}
	case map[string]any:
		for _, key := range []string{"data", "documents", "hits"} {
			if entries, ok := value[key].([]any); ok {
				for _, entry := range entries {
					if item := asMap(entry); item != nil {
						fn(item)
					}
				}
				return
			}
		}
		fn(value)
	}
}

func singleObject(root any) map[string]any {
	if item := asMap(root); item != nil {
		return item
	}
	if entries, ok := root.([]any); ok && len(entries) > 0 {
		if item := asMap(entries[0]); item != nil {
			return item
		}
	}
	return map[string]any{}
}

func asMap(value any) map[string]any {
	if item, ok := value.(map[string]any); ok {
		return item
	}
	return nil
}

func mapField(item map[string]any, key string) map[string]any {
	if item == nil {
		return nil
	}
	return asMap(item[key])
}

func sliceField(item map[string]any, key string) []any {
	if item == nil {
		return nil
	}
	entries, _ := item[key].([]any)
	return entries
}

func stringField(item map[string]any, key string) string {
	if item == nil {
		return ""
	}
	value, _ := item[key].(string)
	return value
}

func intField(item map[string]any, key string) int {
	if item == nil {
		return 0
	}
	switch value := item[key].(type) {
	case float64:
		return int(value)
	case int:
		return value
	case json.Number:
		parsed, _ := value.Int64()
		return int(parsed)
	}
	return 0
}

func setHTMLURL(item map[string]any, url string) {
	if item == nil || url == "" {
		return
	}
	item["html_url"] = url
}

func repoNamespace(item map[string]any, fallback string) string {
	if pathWithNamespace := stringField(mapField(item, "project"), "path_with_namespace"); pathWithNamespace != "" {
		return pathWithNamespace
	}
	return fallback
}

func firstRepoPath(projects []any) string {
	for _, entry := range projects {
		if pathWithNamespace := stringField(asMap(entry), "path_with_namespace"); pathWithNamespace != "" {
			return pathWithNamespace
		}
	}
	return ""
}

func testCaseIdent(item map[string]any) string {
	if ident := stringField(item, "ident"); ident != "" {
		return ident
	}
	if id := intField(item, "id"); id > 0 {
		return fmt.Sprintf("%d", id)
	}
	return ""
}

func pathParam(value string) string {
	if value == "" || IsAllDigits(value) {
		return ""
	}
	return value
}

func firstPositive(values ...int) int {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

func group(groups []string, index int) string {
	if index < 0 || index >= len(groups) {
		return ""
	}
	return groups[index]
}

func atoiSafe(value string) int {
	if value == "" || !IsAllDigits(value) {
		return 0
	}
	result := 0
	for _, char := range value {
		result = result*10 + int(char-'0')
	}
	return result
}

func mustRe(pattern string) *regexp.Regexp {
	return regexp.MustCompile(pattern)
}
