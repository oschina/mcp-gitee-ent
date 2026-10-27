package test_plans

import (
	"context"
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	CreateTestPlanCaseExecution = "create_test_plan_case_execution"
	UpdateTestPlanCaseExecution = "update_test_plan_case_execution"
)

var CreateTestPlanCaseExecutionTool = mcp.NewTool(CreateTestPlanCaseExecution,
	mcp.WithDescription("Record one execution result for a test case in a test plan. "+
		"execute_state: 0 untested, 1 passed, 2 blocked, 3 failed, 4 skipped."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithNumber("test_plan_id", mcp.Description("Test plan ID."), mcp.Required()),
	mcp.WithNumber("test_case_id", mcp.Description("Test case ID."), mcp.Required()),
	mcp.WithNumber("execute_state", mcp.Description("Execution result: 0 untested, 1 passed, 2 blocked, 3 failed, 4 skipped."), mcp.Required(), mcp.Min(0), mcp.Max(4)),
	mcp.WithString("body", mcp.Description("Execution note text.")),
	mcp.WithString("body_json", mcp.Description("Rich-text execution note as JSON.")),
)

var UpdateTestPlanCaseExecutionTool = mcp.NewTool(UpdateTestPlanCaseExecution,
	mcp.WithDescription("Update an existing test case execution record or its note."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithNumber("test_plan_id", mcp.Description("Test plan ID."), mcp.Required()),
	mcp.WithNumber("test_case_id", mcp.Description("Test case ID."), mcp.Required()),
	mcp.WithNumber("execution_id", mcp.Description("Execution record ID."), mcp.Required()),
	mcp.WithNumber("execute_state", mcp.Description("Execution result: 0 untested, 1 passed, 2 blocked, 3 failed, 4 skipped."), mcp.Min(0), mcp.Max(4)),
	mcp.WithString("body", mcp.Description("Updated execution note text.")),
	mcp.WithString("body_json", mcp.Description("Updated rich-text execution note as JSON.")),
)

func CreateTestPlanCaseExecutionHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, testPlanID, err := parsePlanScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	testCaseID, err := requiredInt(arguments, "test_case_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	executeState, err := requiredInt(arguments, "execute_state")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if executeState < 0 || executeState > 4 {
		return mcp.NewToolResultError("execute_state must be 0, 1, 2, 3, or 4"), nil
	}
	payload := map[string]any{"program_id": programID, "execute_state": executeState}
	for _, key := range []string{"body", "body_json"} {
		if value, ok := arguments[key]; ok && value != nil {
			payload[key] = value
		}
	}
	opts = append(opts, utils.WithPayload(payload))
	client := utils.NewGiteeClient(
		"POST",
		fmt.Sprintf("/%d/test_plans/%d/test_cases/%d/execution", enterpriseID, testPlanID, testCaseID),
		opts...,
	)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func UpdateTestPlanCaseExecutionHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, testPlanID, err := parsePlanScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	testCaseID, err := requiredInt(arguments, "test_case_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	executionID, err := requiredInt(arguments, "execution_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	payload := map[string]any{"program_id": programID}
	for _, key := range []string{"execute_state", "body", "body_json"} {
		if value, ok := arguments[key]; ok && value != nil {
			payload[key] = value
		}
	}
	if len(payload) == 1 {
		return mcp.NewToolResultError("At least one of execute_state, body, or body_json is required"), nil
	}
	opts = append(opts, utils.WithPayload(payload))
	client := utils.NewGiteeClient(
		"PUT",
		fmt.Sprintf("/%d/test_plans/%d/test_cases/%d/execution/%d", enterpriseID, testPlanID, testCaseID, executionID),
		opts...,
	)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}
