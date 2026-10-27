package webhooks

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ListWebhooks    = "list_webhooks"
	ListWebhookLogs = "list_webhook_logs"
)

var ListWebhooksTool = mcp.NewTool(ListWebhooks,
	mcp.WithDescription("List webhooks by scope, program, or repository."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("scope", mcp.Description("Webhook scope."), mcp.Enum("all", "enterprise", "program", "project")),
	mcp.WithNumber("program_id", mcp.Description("Optional project ID.")),
	mcp.WithNumber("project_id", mcp.Description("Optional repository ID.")),
	mcp.WithNumber("page", mcp.Description("Page number, starting at 1."), mcp.Min(1)),
	mcp.WithNumber("per_page", mcp.Description("Page size, max 100."), mcp.Min(1), mcp.Max(100)),
	mcp.WithReadOnlyHintAnnotation(true),
)

var ListWebhookLogsTool = mcp.NewTool(ListWebhookLogs,
	mcp.WithDescription("List request logs for one webhook."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("webhook_id", mcp.Description("Webhook ID."), mcp.Required()),
	mcp.WithNumber("page", mcp.Description("Page number, starting at 1."), mcp.Min(1)),
	mcp.WithNumber("per_page", mcp.Description("Page size, max 100."), mcp.Min(1), mcp.Max(100)),
	mcp.WithReadOnlyHintAnnotation(true),
)

func ListWebhooksHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, err := parseEnterprise(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	query := map[string]interface{}{}
	for _, key := range []string{"scope", "program_id", "project_id", "page", "per_page"} {
		if value, ok := arguments[key]; ok {
			query[key] = value
		}
	}
	opts = append(opts, utils.WithQuery(query))
	client := utils.NewGiteeClient("GET", fmt.Sprintf("/%d/hooks", enterpriseID), opts...)
	var data any
	result, err := client.HandleMCPResultContext(ctx, &data)
	if err != nil {
		return result, err
	}
	return utils.MCPJSONResult(redactWebhooks(data))
}

func ListWebhookLogsHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, err := parseEnterprise(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	webhookID, err := requiredInt(arguments, "webhook_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	payload := map[string]any{}
	for _, key := range []string{"page", "per_page"} {
		if value, ok := arguments[key]; ok && value != nil {
			payload[key] = value
		}
	}
	opts = append(opts, utils.WithPayload(payload))
	client := utils.NewGiteeClient(
		"POST",
		fmt.Sprintf("/%d/hooks/%d/logs", enterpriseID, webhookID),
		opts...,
	)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func parseEnterprise(request mcp.CallToolRequest) (map[string]any, int, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return nil, 0, err
	}
	enterpriseID, err := requiredInt(arguments, "enterprise_id")
	if err != nil {
		return nil, 0, err
	}
	return arguments, enterpriseID, nil
}

func requiredInt(arguments map[string]any, key string) (int, error) {
	value, ok := arguments[key]
	if !ok || value == nil {
		return 0, fmt.Errorf("Missing required parameter: %s", key)
	}
	return utils.SafelyConvertToInt(value)
}

var webhookTokenPattern = regexp.MustCompile(`(?i)(access_token|token|key)=([^&]+)`)

func redactWebhooks(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		for key, item := range typed {
			switch strings.ToLower(key) {
			case "password":
				typed[key] = "[REDACTED]"
			case "url", "title":
				if text, ok := item.(string); ok {
					typed[key] = webhookTokenPattern.ReplaceAllString(text, "$1=[REDACTED]")
					continue
				}
				typed[key] = redactWebhooks(item)
			default:
				typed[key] = redactWebhooks(item)
			}
		}
		return typed
	case []any:
		for index, item := range typed {
			typed[index] = redactWebhooks(item)
		}
		return typed
	default:
		return value
	}
}
