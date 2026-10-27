package milestones

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	CreateMilestone = "create_milestone"
	UpdateMilestone = "update_milestone"
)

var CreateMilestoneTool = mcp.NewTool(CreateMilestone,
	mcp.WithDescription("Create a milestone. Use list_milestones first when you need to check for an existing milestone with the same name. "+
		"Requires title and due_date."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("title", mcp.Description("Milestone title."), mcp.Required()),
	mcp.WithString("due_date", mcp.Description("Due date in YYYY-MM-DD format."), mcp.Required()),
	mcp.WithString("start_date", mcp.Description("Start date in YYYY-MM-DD format.")),
	mcp.WithString("description", mcp.Description("Milestone description.")),
	mcp.WithNumber("program_id", mcp.Description("Optional program ID.")),
	mcp.WithString("project_ids", mcp.Description("Comma-separated repository IDs.")),
	mcp.WithNumber("assignee_id", mcp.Description("Optional assignee user ID.")),
)

var UpdateMilestoneTool = mcp.NewTool(UpdateMilestone,
	mcp.WithDescription("Update mutable fields of one milestone. Omitted fields remain unchanged; at least one mutable field is required."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("milestone_id", mcp.Description("Milestone ID."), mcp.Required()),
	mcp.WithString("title", mcp.Description("Milestone title.")),
	mcp.WithString("due_date", mcp.Description("Due date in YYYY-MM-DD format.")),
	mcp.WithString("start_date", mcp.Description("Start date in YYYY-MM-DD format.")),
	mcp.WithString("state_event", mcp.Description("Close or reopen the milestone."), mcp.Enum("activate", "close")),
	mcp.WithString("description", mcp.Description("Milestone description.")),
	mcp.WithNumber("program_id", mcp.Description("Optional program ID.")),
	mcp.WithString("project_ids", mcp.Description("Comma-separated repository IDs.")),
	mcp.WithNumber("assignee_id", mcp.Description("Optional assignee user ID.")),
	mcp.WithNumber("top", mcp.Description("Pin the milestone: 0 unpinned, 1 pinned."), mcp.Min(0), mcp.Max(1)),
)

var milestoneMutableFields = []string{
	"title", "due_date", "start_date", "state_event",
	"description", "program_id", "project_ids", "assignee_id", "top",
}

func CreateMilestoneHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if checkResult, err := utils.CheckRequired(arguments, "title", "due_date"); err != nil {
		return checkResult, err
	}
	enterpriseID, err := utils.SafelyConvertToInt(arguments["enterprise_id"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	title, err := nonEmptyString(arguments, "title")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if err := validateMilestoneDates(arguments, true); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	payload := map[string]any{"title": title}
	for _, key := range []string{"due_date", "start_date", "description", "program_id", "project_ids", "assignee_id"} {
		if value, ok := arguments[key]; ok && value != nil {
			payload[key] = value
		}
	}
	opts = append(opts, utils.WithPayload(payload))

	client := utils.NewGiteeClient("POST", fmt.Sprintf("/%d/milestones", enterpriseID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func UpdateMilestoneHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if checkResult, err := utils.CheckRequired(arguments, "milestone_id"); err != nil {
		return checkResult, err
	}
	enterpriseID, err := utils.SafelyConvertToInt(arguments["enterprise_id"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	milestoneID, err := utils.SafelyConvertToInt(arguments["milestone_id"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	payload := make(map[string]any)
	for _, key := range milestoneMutableFields {
		if value, ok := arguments[key]; ok && value != nil {
			payload[key] = value
		}
	}
	if len(payload) == 0 {
		return mcp.NewToolResultError("At least one mutable field is required: " + strings.Join(milestoneMutableFields, ", ") + "."), nil
	}
	if _, ok := payload["title"]; ok {
		title, err := nonEmptyString(payload, "title")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		payload["title"] = title
	}
	if err := validateMilestoneDates(payload, false); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	opts = append(opts, utils.WithPayload(payload))

	client := utils.NewGiteeClient(
		"PUT",
		fmt.Sprintf("/%d/milestones/%d", enterpriseID, milestoneID),
		opts...,
	)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func nonEmptyString(arguments map[string]any, key string) (string, error) {
	value, ok := arguments[key].(string)
	if !ok || strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("Parameter %s must be a non-empty string", key)
	}
	return strings.TrimSpace(value), nil
}

func validateMilestoneDates(arguments map[string]any, dueRequired bool) error {
	if dueRequired {
		if _, ok := arguments["due_date"]; !ok {
			return fmt.Errorf("Missing required parameter: due_date")
		}
	}
	for _, key := range []string{"due_date", "start_date"} {
		value, ok := arguments[key]
		if !ok || value == nil {
			continue
		}
		text, ok := value.(string)
		if !ok {
			return fmt.Errorf("Parameter %s must be a YYYY-MM-DD string", key)
		}
		if _, err := time.Parse("2006-01-02", strings.TrimSpace(text)); err != nil {
			return fmt.Errorf("Parameter %s must use YYYY-MM-DD format: %w", key, err)
		}
	}
	return nil
}
