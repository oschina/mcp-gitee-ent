package test_plans

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ListTestPlanCases        = "list_test_plan_cases"
	GetTestPlanCase          = "get_test_plan_case"
	AddTestCasesToPlan       = "add_test_cases_to_plan"
	CreateTestCaseInPlan     = "create_test_case_in_plan"
	UpdateTestPlanCase       = "update_test_plan_case"
	LinkTestPlanCaseIssues   = "link_test_plan_case_issues"
	UnlinkTestPlanCaseIssues = "unlink_test_plan_case_issues"
	ListTestPlanCaseIssues   = "list_test_plan_case_issues"
)

var ListTestPlanCasesTool = mcp.NewTool(ListTestPlanCases,
	mcp.WithDescription("List test cases in a test plan. Filter by execution result, executor, module, state, or linked_issue_id."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithNumber("test_plan_id", mcp.Description("Test plan ID."), mcp.Required()),
	mcp.WithNumber("module_id", mcp.Description("Optional module ID.")),
	mcp.WithString("case_types", mcp.Description("Comma-separated case type IDs: 1-9.")),
	mcp.WithString("maintainer_ids", mcp.Description("Comma-separated maintainer user IDs.")),
	mcp.WithString("priorities", mcp.Description("Comma-separated priorities: 0-4.")),
	mcp.WithString("states", mcp.Description("Comma-separated case states: 0-4.")),
	mcp.WithString("creator_ids", mcp.Description("Comma-separated creator user IDs.")),
	mcp.WithString("results", mcp.Description("Comma-separated execution results: 0 untested, 1 passed, 2 blocked, 3 failed, 4 skipped.")),
	mcp.WithString("executor_ids", mcp.Description("Comma-separated executor user IDs.")),
	mcp.WithNumber("issue_id", mcp.Description("Only return cases not yet linked to this issue.")),
	mcp.WithString("search", mcp.Description("Search keyword.")),
	mcp.WithNumber("page", mcp.Description("Page number, starting at 1."), mcp.Min(1)),
	mcp.WithNumber("per_page", mcp.Description("Page size, max 100."), mcp.Min(1), mcp.Max(100)),
	mcp.WithReadOnlyHintAnnotation(true),
)

var GetTestPlanCaseTool = mcp.NewTool(GetTestPlanCase,
	mcp.WithDescription("Get one test case in a test plan, including execution notes and history."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithNumber("test_plan_id", mcp.Description("Test plan ID."), mcp.Required()),
	mcp.WithNumber("test_case_id", mcp.Description("Test case ID."), mcp.Required()),
	mcp.WithReadOnlyHintAnnotation(true),
)

var AddTestCasesToPlanTool = mcp.NewTool(AddTestCasesToPlan,
	mcp.WithDescription("Add existing test cases to a test plan. Provide case_ids for automatic version resolution, "+
		"or version_ids in caseId-versionId,... format when versions are already known. "+
		"Only cases with an approved version can be added; a freshly created case has none until it passes review, "+
		"so use create_test_case_in_plan for new cases. Adding is asynchronous: the call returns a queued acknowledgement, "+
		"verify with list_test_plan_cases."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithNumber("test_plan_id", mcp.Description("Test plan ID."), mcp.Required()),
	mcp.WithArray("case_ids", mcp.Description("Test case IDs to add. The server resolves the latest approved version for each case."), mcp.Items(map[string]any{"type": "number"})),
	mcp.WithString("version_ids", mcp.Description("Advanced override: comma-separated caseId-versionId pairs.")),
	mcp.WithArray("module_ids", mcp.Description("Optional module IDs; use with version_ids."), mcp.Items(map[string]any{"type": "number"})),
	mcp.WithArray("exclude_case_ids", mcp.Description("Case IDs to exclude when adding by module."), mcp.Items(map[string]any{"type": "number"})),
	mcp.WithBoolean("all", mcp.Description("Add all cases when used with version_ids.")),
)

var CreateTestCaseInPlanTool = mcp.NewTool(CreateTestCaseInPlan,
	mcp.WithDescription("Create a new test case directly inside a test plan. Requires title, module_id, and case_steps. "+
		"Every step needs a non-empty description; expected_result is optional. Step order follows the array order."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithNumber("test_plan_id", mcp.Description("Test plan ID."), mcp.Required()),
	mcp.WithString("title", mcp.Description("Test case title."), mcp.Required()),
	mcp.WithNumber("module_id", mcp.Description("Module ID."), mcp.Required()),
	mcp.WithString("precondition", mcp.Description("Precondition.")),
	mcp.WithNumber("case_type", mcp.Description("Case type: 1-9."), mcp.Min(1), mcp.Max(9)),
	mcp.WithNumber("priority", mcp.Description("Priority: 0-4."), mcp.Min(0), mcp.Max(4)),
	mcp.WithNumber("maintainer_id", mcp.Description("Maintainer user ID.")),
	mcp.WithString("remark", mcp.Description("Remark.")),
	mcp.WithString("attachment_ids", mcp.Description("Comma-separated attachment IDs.")),
	mcp.WithArray("case_steps",
		mcp.Description("Required case steps, at least one. Each item needs description; expected_result is optional."),
		mcp.Required(),
		mcp.MinItems(1),
		mcp.Items(map[string]any{
			"type": "object",
			"properties": map[string]any{
				"description":     map[string]any{"type": "string"},
				"expected_result": map[string]any{"type": "string"},
			},
			"required": []string{"description"},
		}),
	),
)

var UpdateTestPlanCaseTool = mcp.NewTool(UpdateTestPlanCase,
	mcp.WithDescription("Update the assignee of one test case in a test plan."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithNumber("test_plan_id", mcp.Description("Test plan ID."), mcp.Required()),
	mcp.WithNumber("test_case_id", mcp.Description("Test case ID."), mcp.Required()),
	mcp.WithNumber("assignee_id", mcp.Description("New assignee user ID."), mcp.Required()),
)

var LinkTestPlanCaseIssuesTool = mcp.NewTool(LinkTestPlanCaseIssues,
	mcp.WithDescription("Link work items to a test case in a test plan."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithNumber("test_plan_id", mcp.Description("Test plan ID."), mcp.Required()),
	mcp.WithNumber("test_case_id", mcp.Description("Test case ID."), mcp.Required()),
	mcp.WithArray("issue_ids", mcp.Description("Work item IDs to link."), mcp.Required(), mcp.Items(map[string]any{"type": "number"})),
)

var UnlinkTestPlanCaseIssuesTool = mcp.NewTool(UnlinkTestPlanCaseIssues,
	mcp.WithDescription("Unlink work items from a test case in a test plan."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithNumber("test_plan_id", mcp.Description("Test plan ID."), mcp.Required()),
	mcp.WithNumber("test_case_id", mcp.Description("Test case ID."), mcp.Required()),
	mcp.WithArray("issue_ids", mcp.Description("Work item IDs to unlink."), mcp.Required(), mcp.Items(map[string]any{"type": "number"})),
)

var ListTestPlanCaseIssuesTool = mcp.NewTool(ListTestPlanCaseIssues,
	mcp.WithDescription("List work items linked to a test case in a test plan."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithNumber("test_plan_id", mcp.Description("Test plan ID."), mcp.Required()),
	mcp.WithNumber("test_case_id", mcp.Description("Test case ID."), mcp.Required()),
	mcp.WithBoolean("linked", mcp.Description("Return linked issues when true, unlinked issues when false.")),
	mcp.WithString("search", mcp.Description("Search keyword.")),
	mcp.WithString("only_related_me", mcp.Description("Whether to list only issues related to the authorized user: 0 no, 1 yes."), mcp.Enum("0", "1")),
	mcp.WithString("assignee_id", mcp.Description("Assignee user ID.")),
	mcp.WithString("issue_type_id", mcp.Description("Issue type ID.")),
	mcp.WithString("author_id", mcp.Description("Creator user ID.")),
	mcp.WithString("collaborator_ids", mcp.Description("Comma-separated collaborator user IDs.")),
	mcp.WithNumber("page", mcp.Description("Page number, starting at 1."), mcp.Min(1)),
	mcp.WithNumber("per_page", mcp.Description("Page size, max 100."), mcp.Min(1), mcp.Max(100)),
	mcp.WithReadOnlyHintAnnotation(true),
)

func ListTestPlanCasesHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, testPlanID, err := parsePlanScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	query := map[string]interface{}{"program_id": programID}
	for _, key := range []string{
		"module_id", "case_types", "maintainer_ids", "priorities", "states",
		"creator_ids", "results", "executor_ids", "issue_id", "search", "page", "per_page",
	} {
		if value, ok := arguments[key]; ok {
			query[key] = value
		}
	}
	opts = append(opts, utils.WithQuery(query))
	client := utils.NewGiteeClient("GET", fmt.Sprintf("/%d/test_plans/%d/test_cases", enterpriseID, testPlanID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func GetTestPlanCaseHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, testPlanID, err := parsePlanScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	testCaseID, err := requiredInt(arguments, "test_case_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	opts = append(opts, utils.WithQuery(map[string]interface{}{"program_id": programID}))
	client := utils.NewGiteeClient(
		"GET",
		fmt.Sprintf("/%d/test_plans/%d/test_cases/%d", enterpriseID, testPlanID, testCaseID),
		opts...,
	)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func AddTestCasesToPlanHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, testPlanID, err := parsePlanScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	payload := map[string]any{"program_id": programID}

	versionIDs, _ := arguments["version_ids"].(string)
	if strings.TrimSpace(versionIDs) == "" {
		caseIDs, err := intArray(arguments["case_ids"])
		if err != nil || len(caseIDs) == 0 {
			return mcp.NewToolResultError("Provide case_ids for automatic version resolution, or version_ids in caseId-versionId,... format."), nil
		}
		resolved, err := resolveVersionIDs(ctx, enterpriseID, programID, caseIDs)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		versionIDs = resolved
	}
	payload["version_ids"] = versionIDs

	for _, key := range []string{"module_ids", "exclude_case_ids"} {
		if value, ok := arguments[key]; ok && value != nil {
			payload[key] = value
		}
	}
	if value, ok := arguments["case_ids"]; ok && value != nil {
		payload["case_ids"] = value
	}
	if value, ok := arguments["all"]; ok && value != nil {
		payload["all"] = value
	}
	opts = append(opts, utils.WithPayload(payload))
	client := utils.NewGiteeClient(
		"POST",
		fmt.Sprintf("/%d/test_plans/%d/test_cases/add_test_cases", enterpriseID, testPlanID),
		opts...,
	)
	if result, err := client.HandleMCPResultContext(ctx, nil); result != nil && result.IsError {
		return result, err
	}
	return utils.MCPActionResult(map[string]any{
		"test_plan_id": testPlanID,
		"version_ids":  versionIDs,
		"queued":       true,
	})
}

func CreateTestCaseInPlanHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, testPlanID, err := parsePlanScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if _, err := requiredText(arguments, "title"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if _, err := requiredInt(arguments, "module_id"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	caseSteps, err := utils.NormalizeCaseSteps(arguments["case_steps"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	payload := map[string]any{"program_id": programID, "case_steps": caseSteps}
	for _, key := range []string{
		"title", "precondition", "case_type", "priority", "maintainer_id",
		"remark", "attachment_ids", "module_id",
	} {
		if value, ok := arguments[key]; ok && value != nil {
			payload[key] = value
		}
	}
	opts = append(opts, utils.WithPayload(payload))
	client := utils.NewGiteeClient(
		"POST",
		fmt.Sprintf("/%d/test_plans/%d/test_cases", enterpriseID, testPlanID),
		opts...,
	)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func UpdateTestPlanCaseHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, testPlanID, err := parsePlanScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	testCaseID, err := requiredInt(arguments, "test_case_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	assigneeID, err := requiredInt(arguments, "assignee_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	payload := map[string]any{"program_id": programID, "assignee_id": assigneeID}
	opts = append(opts, utils.WithPayload(payload))
	client := utils.NewGiteeClient(
		"PUT",
		fmt.Sprintf("/%d/test_plans/%d/test_cases/%d", enterpriseID, testPlanID, testCaseID),
		opts...,
	)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func LinkTestPlanCaseIssuesHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	return changePlanCaseIssues(ctx, request, opts, "POST")
}

func UnlinkTestPlanCaseIssuesHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	return changePlanCaseIssues(ctx, request, opts, "DELETE")
}

func changePlanCaseIssues(ctx context.Context, request mcp.CallToolRequest, opts []utils.Option, method string) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, testPlanID, err := parsePlanScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	testCaseID, err := requiredInt(arguments, "test_case_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	issueIDs, err := intArray(arguments["issue_ids"])
	if err != nil || len(issueIDs) == 0 {
		return mcp.NewToolResultError("issue_ids must contain at least one work item ID"), nil
	}
	payload := map[string]any{"program_id": programID, "issue_ids": issueIDs}
	opts = append(opts, utils.WithPayload(payload))
	client := utils.NewGiteeClient(
		method,
		fmt.Sprintf("/%d/test_plans/%d/test_cases/%d/link_issues", enterpriseID, testPlanID, testCaseID),
		opts...,
	)
	if result, err := client.HandleMCPResultContext(ctx, nil); result != nil && result.IsError {
		return result, err
	}
	return utils.MCPActionResult(map[string]any{
		"test_plan_id": testPlanID,
		"test_case_id": testCaseID,
		"issue_ids":    issueIDs,
		"linked":       method == "POST",
	})
}

func ListTestPlanCaseIssuesHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, testPlanID, err := parsePlanScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	testCaseID, err := requiredInt(arguments, "test_case_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	query := map[string]interface{}{"program_id": programID}
	for _, key := range []string{
		"linked", "search", "only_related_me", "assignee_id",
		"issue_type_id", "author_id", "collaborator_ids", "page", "per_page",
	} {
		if value, ok := arguments[key]; ok {
			query[key] = value
		}
	}
	opts = append(opts, utils.WithQuery(query))
	client := utils.NewGiteeClient(
		"GET",
		fmt.Sprintf("/%d/test_plans/%d/test_cases/%d/issues", enterpriseID, testPlanID, testCaseID),
		opts...,
	)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func parsePlanScope(request mcp.CallToolRequest) (map[string]any, int, int, int, error) {
	arguments, enterpriseID, programID, err := parseScope(request)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	testPlanID, err := requiredInt(arguments, "test_plan_id")
	if err != nil {
		return nil, 0, 0, 0, err
	}
	return arguments, enterpriseID, programID, testPlanID, nil
}

type versionItem struct {
	ID    int `json:"id"`
	State int `json:"state"`
}

type versionListResponse struct {
	Data []versionItem `json:"data"`
}

func resolveVersionIDs(ctx context.Context, enterpriseID, programID int, caseIDs []int) (string, error) {
	pairs := make([]string, 0, len(caseIDs))
	for _, caseID := range caseIDs {
		versionID, err := latestVersionID(ctx, enterpriseID, programID, caseID)
		if err != nil {
			return "", err
		}
		pairs = append(pairs, fmt.Sprintf("%d-%d", caseID, versionID))
	}
	return strings.Join(pairs, ","), nil
}

func latestVersionID(ctx context.Context, enterpriseID, programID, caseID int) (int, error) {
	client := utils.NewGiteeClient(
		"GET",
		fmt.Sprintf("/%d/test_cases/%d/test_case_versions", enterpriseID, caseID),
		utils.WithQuery(map[string]interface{}{"program_id": programID, "page": 1, "per_page": 100}),
	)
	if _, err := client.DoContext(ctx); err != nil {
		return 0, fmt.Errorf("failed to read versions for test case %d: %w", caseID, err)
	}
	defer client.Response.Body.Close()
	body, err := client.GetRespBody()
	if err != nil {
		return 0, fmt.Errorf("failed to read versions for test case %d: %w", caseID, err)
	}
	var response versionListResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return 0, fmt.Errorf("failed to parse versions for test case %d: %w", caseID, err)
	}
	bestApproved := 0
	for _, version := range response.Data {
		if version.State == 2 && version.ID > bestApproved {
			bestApproved = version.ID
		}
	}
	if bestApproved > 0 {
		return bestApproved, nil
	}
	return 0, fmt.Errorf("test case %d has no approved version, so it cannot be added by case_ids; use create_test_case_in_plan to create it directly inside the plan", caseID)
}

func intArray(value any) ([]int, error) {
	items, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("expected an array of numeric IDs")
	}
	result := make([]int, 0, len(items))
	for _, item := range items {
		id, err := utils.SafelyConvertToInt(item)
		if err != nil {
			return nil, err
		}
		result = append(result, id)
	}
	return result, nil
}
