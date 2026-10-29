package test_plans

import (
	"context"
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const ExecuteTestPlanCases = "execute_test_plan_cases"

var ExecuteTestPlanCasesTool = mcp.NewTool(ExecuteTestPlanCases,
	mcp.WithDescription("Record execution results for one or more test cases in a test plan (asynchronous batch). "+
		"test_plan_case_ids are the ids returned by list_test_plan_cases, not the repository test case ids; "+
		"provide them unless is_all is true with custom_conditions. At least one of execute_state or assignee_id is required. "+
		"execute_state: 0 untested, 1 passed, 2 blocked, 3 failed, 4 skipped. "+
		"The API queues a background job, so this call returns a submission message; confirm results with "+
		"get_test_plan or list_test_plan_cases afterwards."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithNumber("test_plan_id", mcp.Description("Test plan ID."), mcp.Required()),
	mcp.WithArray("test_plan_case_ids",
		mcp.Description("Test plan case IDs from list_test_plan_cases. Required unless is_all is true."),
		mcp.Items(map[string]any{"type": "number"})),
	mcp.WithBoolean("is_all", mcp.Description("Apply to every case matching custom_conditions instead of an explicit id list.")),
	mcp.WithString("custom_conditions", mcp.Description("JSON filter expression used together with is_all.")),
	mcp.WithNumber("execute_state",
		mcp.Description("Execution result to record: 0 untested, 1 passed, 2 blocked, 3 failed, 4 skipped."),
		mcp.Min(0), mcp.Max(4)),
	mcp.WithNumber("assignee_id", mcp.Description("Reassign the owner of the targeted cases.")),
	mcp.WithString("body", mcp.Description("Execution note text.")),
)

func ExecuteTestPlanCasesHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, testPlanID, err := parsePlanScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	isAll, _ := arguments["is_all"].(bool)
	payload := map[string]any{"program_id": programID}
	if isAll {
		payload["is_all"] = true
		if conditions, ok := arguments["custom_conditions"].(string); ok && conditions != "" {
			payload["custom_conditions"] = conditions
		}
	} else {
		caseIDs, err := intArray(arguments["test_plan_case_ids"])
		if err != nil || len(caseIDs) == 0 {
			return mcp.NewToolResultError("test_plan_case_ids must contain at least one test plan case ID unless is_all is true"), nil
		}
		payload["test_plan_case_ids"] = caseIDs
	}

	for _, key := range []string{"execute_state", "assignee_id", "body"} {
		if value, ok := arguments[key]; ok && value != nil {
			payload[key] = value
		}
	}
	if _, hasState := payload["execute_state"]; !hasState {
		if _, hasAssignee := payload["assignee_id"]; !hasAssignee {
			return mcp.NewToolResultError("At least one of execute_state or assignee_id is required"), nil
		}
	}

	opts = append(opts, utils.WithPayload(payload))
	client := utils.NewGiteeClient(
		"POST",
		fmt.Sprintf("/%d/test_plans/%d/batch_execute", enterpriseID, testPlanID),
		opts...,
	)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}
