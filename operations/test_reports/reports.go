package test_reports

import (
	"context"
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const CreateTestReport = "create_test_report"

var reportComponentIdents = []string{
	"basic-info",
	"related-info",
	"summary",
	"overview",
	"case-statistics",
	"bug-statistics",
	"users-bug-statistics",
	"bug-list",
	"issue-list",
	"pr-list",
	"execute-process",
	"case-list",
	"issue-statistics",
	"users-issue-statistics",
}

var CreateTestReportTool = mcp.NewTool(CreateTestReport,
	mcp.WithDescription("Create a test report from one or more existing test plans. "+
		"Requires title, test_plan_ids, and components: the Gitee API rejects a report with no components. "+
		"Each component needs ident and title; start with basic-info and case-statistics."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithString("title", mcp.Description("Test report title."), mcp.Required()),
	mcp.WithArray("test_plan_ids",
		mcp.Description("Test plan IDs to include in the report."),
		mcp.Required(),
		mcp.Items(map[string]any{"type": "number"}),
	),
	mcp.WithBoolean("state", mcp.Description("Whether the test report is marked as passed.")),
	mcp.WithArray("components",
		mcp.Description("Report components. Required by the Gitee API; at least one is needed. Allowed ident values: basic-info, related-info, summary, overview, case-statistics, bug-statistics, users-bug-statistics, bug-list, issue-list, pr-list, execute-process, case-list, issue-statistics, users-issue-statistics."),
		mcp.Required(),
		mcp.Items(map[string]any{
			"type": "object",
			"properties": map[string]any{
				"ident":    map[string]any{"type": "string", "enum": reportComponentIdents},
				"title":    map[string]any{"type": "string"},
				"query":    map[string]any{"type": "object"},
				"metadata": map[string]any{"type": "object"},
			},
			"required": []string{"ident", "title"},
		}),
	),
)

func CreateTestReportHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if checkResult, err := utils.CheckRequired(arguments, "title", "test_plan_ids", "components"); err != nil {
		return checkResult, err
	}
	enterpriseID, err := utils.SafelyConvertToInt(arguments["enterprise_id"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	programID, err := utils.SafelyConvertToInt(arguments["program_id"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	title, ok := arguments["title"].(string)
	if !ok || title == "" {
		return mcp.NewToolResultError("Parameter title must be a non-empty string"), nil
	}
	testPlanIDs, err := intArray(arguments["test_plan_ids"])
	if err != nil || len(testPlanIDs) == 0 {
		return mcp.NewToolResultError("test_plan_ids must contain at least one test plan ID"), nil
	}
	components, ok := arguments["components"].([]any)
	if !ok || len(components) == 0 {
		return mcp.NewToolResultError("components must contain at least one component such as {\"ident\":\"basic-info\",\"title\":\"Basic info\"}"), nil
	}

	payload := map[string]any{
		"program_id":    programID,
		"title":         title,
		"test_plan_ids": testPlanIDs,
		"components":    components,
	}
	if value, ok := arguments["state"]; ok && value != nil {
		payload["state"] = value
	}
	opts = append(opts, utils.WithPayload(payload))

	client := utils.NewGiteeClient(
		"POST",
		fmt.Sprintf("/%d/test_reports", enterpriseID),
		opts...,
	)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
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

const (
	ListTestReports  = "list_test_reports"
	GetTestReport    = "get_test_report"
	UpdateTestReport = "update_test_report"
	DeleteTestReport = "delete_test_report"
)

var ListTestReportsTool = mcp.NewTool(ListTestReports,
	mcp.WithDescription("List test reports in a project. Filter by keyword, creator, linked test plans, or date range."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithString("search", mcp.Description("Test report title keyword.")),
	mcp.WithString("creator_ids", mcp.Description("Comma-separated creator user IDs.")),
	mcp.WithString("test_plan_ids", mcp.Description("Comma-separated test plan IDs.")),
	mcp.WithString("start_date", mcp.Description("Created on or after this date, YYYY-MM-DD.")),
	mcp.WithString("end_date", mcp.Description("Created on or before this date, YYYY-MM-DD.")),
	mcp.WithNumber("page", mcp.Description("Page number, starting at 1."), mcp.Min(1)),
	mcp.WithNumber("per_page", mcp.Description("Page size, max 100."), mcp.Min(1), mcp.Max(100)),
	mcp.WithReadOnlyHintAnnotation(true),
)

var GetTestReportTool = mcp.NewTool(GetTestReport,
	mcp.WithDescription("Get one test report with its computed components. The Gitee API uses POST for this read operation."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithNumber("test_report_id", mcp.Description("Test report ID."), mcp.Required()),
	mcp.WithArray("components",
		mcp.Description("Optional components to compute. Each item requires id, ident, and title."),
		mcp.Items(map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id":    map[string]any{"type": "number"},
				"ident": map[string]any{"type": "string", "enum": reportComponentIdents},
				"title": map[string]any{"type": "string"},
				"query": map[string]any{"type": "object"},
			},
			"required": []string{"id", "ident", "title"},
		}),
	),
	mcp.WithReadOnlyHintAnnotation(true),
)

var UpdateTestReportTool = mcp.NewTool(UpdateTestReport,
	mcp.WithDescription("Update a test report title, pass state, components, or recompute its statistics."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithNumber("test_report_id", mcp.Description("Test report ID."), mcp.Required()),
	mcp.WithString("title", mcp.Description("New report title.")),
	mcp.WithBoolean("state", mcp.Description("Whether the test report is marked as passed.")),
	mcp.WithBoolean("reload", mcp.Description("Whether to recompute report statistics.")),
	mcp.WithArray("components",
		mcp.Description("Existing report components to update. Each item requires id; title and metadata are optional."),
		mcp.Items(map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id":       map[string]any{"type": "number"},
				"title":    map[string]any{"type": "string"},
				"metadata": map[string]any{"type": "object"},
			},
			"required": []string{"id"},
		}),
	),
)

var DeleteTestReportTool = mcp.NewTool(DeleteTestReport,
	mcp.WithDescription("Delete one test report. This does not delete its linked test plans."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithNumber("test_report_id", mcp.Description("Test report ID."), mcp.Required()),
	mcp.WithDestructiveHintAnnotation(true),
)

func ListTestReportsHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, err := parseScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	query := map[string]interface{}{"program_id": programID}
	for _, key := range []string{"search", "creator_ids", "test_plan_ids", "start_date", "end_date", "page", "per_page"} {
		if value, ok := arguments[key]; ok {
			query[key] = value
		}
	}
	opts = append(opts, utils.WithQuery(query))
	client := utils.NewGiteeClient("GET", fmt.Sprintf("/%d/test_reports", enterpriseID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func GetTestReportHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, err := parseScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	reportID, err := requiredInt(arguments, "test_report_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	payload := map[string]any{"program_id": programID}
	if value, ok := arguments["components"]; ok && value != nil {
		payload["components"] = value
	}
	opts = append(opts, utils.WithPayload(payload))
	client := utils.NewGiteeClient("POST", fmt.Sprintf("/%d/test_reports/%d", enterpriseID, reportID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func UpdateTestReportHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, err := parseScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	reportID, err := requiredInt(arguments, "test_report_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	payload := map[string]any{"program_id": programID}
	for _, key := range []string{"title", "state", "reload", "components"} {
		if value, ok := arguments[key]; ok && value != nil {
			payload[key] = value
		}
	}
	if len(payload) == 1 {
		return mcp.NewToolResultError("At least one of title, state, reload, or components is required"), nil
	}
	opts = append(opts, utils.WithPayload(payload))
	client := utils.NewGiteeClient("PUT", fmt.Sprintf("/%d/test_reports/%d", enterpriseID, reportID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func DeleteTestReportHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, err := parseScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	reportID, err := requiredInt(arguments, "test_report_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	opts = append(opts, utils.WithQuery(map[string]interface{}{"program_id": programID}))
	client := utils.NewGiteeClient("DELETE", fmt.Sprintf("/%d/test_reports/%d", enterpriseID, reportID), opts...)
	if result, err := client.HandleMCPResultContext(ctx, nil); result != nil && result.IsError {
		return result, err
	}
	return utils.MCPActionResult(map[string]any{"test_report_id": reportID, "deleted": true})
}

func parseScope(request mcp.CallToolRequest) (map[string]any, int, int, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return nil, 0, 0, err
	}
	enterpriseID, err := requiredInt(arguments, "enterprise_id")
	if err != nil {
		return nil, 0, 0, err
	}
	programID, err := requiredInt(arguments, "program_id")
	if err != nil {
		return nil, 0, 0, err
	}
	return arguments, enterpriseID, programID, nil
}

func requiredInt(arguments map[string]any, key string) (int, error) {
	value, ok := arguments[key]
	if !ok || value == nil {
		return 0, fmt.Errorf("Missing required parameter: %s", key)
	}
	return utils.SafelyConvertToInt(value)
}
