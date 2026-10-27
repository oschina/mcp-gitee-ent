package test_plans

import (
	"context"
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ListTestPlans       = "list_test_plans"
	GetTestPlan         = "get_test_plan"
	CreateTestPlan      = "create_test_plan"
	UpdateTestPlan      = "update_test_plan"
	UpdateTestPlanState = "update_test_plan_state"
)

var ListTestPlansTool = mcp.NewTool(ListTestPlans,
	mcp.WithDescription("List test plans in a project. Filter by state, keyword, or linked milestone/sprint/version."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithString("states", mcp.Description("Comma-separated states: 0 not started, 1 in progress, 2 completed, 3 suspended.")),
	mcp.WithString("search", mcp.Description("Test plan title keyword.")),
	mcp.WithString("sort", mcp.Description("Sort field: created_at, updated_at, state, priority, ident.")),
	mcp.WithString("direction", mcp.Description("Sort direction."), mcp.Enum("asc", "desc")),
	mcp.WithString("ref_type", mcp.Description("Linked resource type."), mcp.Enum("sprint", "version", "milestone")),
	mcp.WithNumber("ref_id", mcp.Description("Linked resource ID.")),
	mcp.WithNumber("page", mcp.Description("Page number, starting at 1."), mcp.Min(1)),
	mcp.WithNumber("per_page", mcp.Description("Page size, max 100."), mcp.Min(1), mcp.Max(100)),
	mcp.WithReadOnlyHintAnnotation(true),
)

var GetTestPlanTool = mcp.NewTool(GetTestPlan,
	mcp.WithDescription("Get one test plan by ID."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithNumber("test_plan_id", mcp.Description("Test plan ID."), mcp.Required()),
	mcp.WithReadOnlyHintAnnotation(true),
)

var CreateTestPlanTool = mcp.NewTool(CreateTestPlan,
	mcp.WithDescription("Create a test plan. Requires title and assignee_id. A plan may optionally be linked to a milestone, sprint, or version."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithString("title", mcp.Description("Test plan title."), mcp.Required()),
	mcp.WithNumber("assignee_id", mcp.Description("Test plan assignee user ID."), mcp.Required()),
	mcp.WithString("ref_type", mcp.Description("Linked resource type."), mcp.Enum("milestone", "sprint", "version")),
	mcp.WithString("description", mcp.Description("Test plan description.")),
	mcp.WithString("start_date", mcp.Description("Start date in YYYY-MM-DD format.")),
	mcp.WithString("end_date", mcp.Description("End date in YYYY-MM-DD format.")),
	mcp.WithNumber("milestone_id", mcp.Description("Milestone ID when ref_type is milestone.")),
	mcp.WithNumber("scrum_sprint_id", mcp.Description("Sprint ID when ref_type is sprint.")),
	mcp.WithNumber("scrum_version_id", mcp.Description("Version ID when ref_type is version.")),
)

var UpdateTestPlanTool = mcp.NewTool(UpdateTestPlan,
	mcp.WithDescription("Update a test plan. The Gitee API requires title and assignee_id on update as well."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithNumber("test_plan_id", mcp.Description("Test plan ID."), mcp.Required()),
	mcp.WithString("title", mcp.Description("Test plan title."), mcp.Required()),
	mcp.WithNumber("assignee_id", mcp.Description("Test plan assignee user ID."), mcp.Required()),
	mcp.WithString("ref_type", mcp.Description("Linked resource type."), mcp.Enum("milestone", "sprint", "version")),
	mcp.WithString("description", mcp.Description("Test plan description.")),
	mcp.WithString("start_date", mcp.Description("Start date in YYYY-MM-DD format.")),
	mcp.WithString("end_date", mcp.Description("End date in YYYY-MM-DD format.")),
	mcp.WithNumber("milestone_id", mcp.Description("Milestone ID when ref_type is milestone.")),
	mcp.WithNumber("scrum_sprint_id", mcp.Description("Sprint ID when ref_type is sprint.")),
	mcp.WithNumber("scrum_version_id", mcp.Description("Version ID when ref_type is version.")),
)

var UpdateTestPlanStateTool = mcp.NewTool(UpdateTestPlanState,
	mcp.WithDescription("Change a test plan state: 0 not started, 1 in progress, 2 completed, 3 suspended."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithNumber("test_plan_id", mcp.Description("Test plan ID."), mcp.Required()),
	mcp.WithNumber("state", mcp.Description("New state: 0 not started, 1 in progress, 2 completed, 3 suspended."), mcp.Required(), mcp.Min(0), mcp.Max(3)),
)

func ListTestPlansHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, err := parseScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	query := map[string]interface{}{"program_id": programID}
	for _, key := range []string{"states", "search", "sort", "direction", "ref_type", "ref_id", "page", "per_page"} {
		if value, ok := arguments[key]; ok {
			query[key] = value
		}
	}
	opts = append(opts, utils.WithQuery(query))
	client := utils.NewGiteeClient("GET", fmt.Sprintf("/%d/test_plans", enterpriseID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func GetTestPlanHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, err := parseScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	testPlanID, err := requiredInt(arguments, "test_plan_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	opts = append(opts, utils.WithQuery(map[string]interface{}{"program_id": programID}))
	client := utils.NewGiteeClient("GET", fmt.Sprintf("/%d/test_plans/%d", enterpriseID, testPlanID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func CreateTestPlanHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, err := parseScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if _, err := requiredText(arguments, "title"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if _, err := requiredInt(arguments, "assignee_id"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	payload := map[string]any{"program_id": programID}
	for _, key := range []string{
		"title", "assignee_id", "ref_type", "description", "start_date", "end_date",
		"milestone_id", "scrum_sprint_id", "scrum_version_id",
	} {
		if value, ok := arguments[key]; ok && value != nil {
			payload[key] = value
		}
	}
	opts = append(opts, utils.WithPayload(payload))
	client := utils.NewGiteeClient("POST", fmt.Sprintf("/%d/test_plans", enterpriseID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func UpdateTestPlanHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, err := parseScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	testPlanID, err := requiredInt(arguments, "test_plan_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if _, err := requiredText(arguments, "title"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if _, err := requiredInt(arguments, "assignee_id"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	payload := map[string]any{"program_id": programID}
	for _, key := range []string{
		"title", "assignee_id", "ref_type", "description", "start_date", "end_date",
		"milestone_id", "scrum_sprint_id", "scrum_version_id",
	} {
		if value, ok := arguments[key]; ok && value != nil {
			payload[key] = value
		}
	}
	opts = append(opts, utils.WithPayload(payload))
	client := utils.NewGiteeClient("PUT", fmt.Sprintf("/%d/test_plans/%d", enterpriseID, testPlanID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func UpdateTestPlanStateHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, err := parseScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	testPlanID, err := requiredInt(arguments, "test_plan_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	state, err := requiredInt(arguments, "state")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if state < 0 || state > 3 {
		return mcp.NewToolResultError("state must be 0, 1, 2, or 3"), nil
	}
	payload := map[string]any{"program_id": programID, "state": state}
	opts = append(opts, utils.WithPayload(payload))
	client := utils.NewGiteeClient("PUT", fmt.Sprintf("/%d/test_plans/%d/state", enterpriseID, testPlanID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
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

func requiredText(arguments map[string]any, key string) (string, error) {
	value, ok := arguments[key].(string)
	if !ok || value == "" {
		return "", fmt.Errorf("Missing required parameter: %s", key)
	}
	return value, nil
}

const DeleteTestPlan = "delete_test_plan"

var DeleteTestPlanTool = mcp.NewTool(DeleteTestPlan,
	mcp.WithDescription("Delete one test plan. Its plan cases and execution records are cleaned up asynchronously."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithNumber("test_plan_id", mcp.Description("Test plan ID."), mcp.Required()),
	mcp.WithDestructiveHintAnnotation(true),
)

func DeleteTestPlanHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, err := parseScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	testPlanID, err := requiredInt(arguments, "test_plan_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	opts = append(opts, utils.WithQuery(map[string]interface{}{"program_id": programID}))
	client := utils.NewGiteeClient("DELETE", fmt.Sprintf("/%d/test_plans/%d", enterpriseID, testPlanID), opts...)
	return client.HandleMCPResultContext(ctx, nil)
}
