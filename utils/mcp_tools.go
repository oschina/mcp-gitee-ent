package utils

import "github.com/mark3labs/mcp-go/mcp"

// ConvertArgumentsToMap 将 any 类型的参数安全转换为 map[string]interface{}
func ConvertArgumentsToMap(arguments any) (map[string]interface{}, error) {
	if arguments == nil {
		return make(map[string]interface{}), nil
	}

	if argMap, ok := arguments.(map[string]interface{}); ok {
		return argMap, nil
	}

	return nil, NewParamError("arguments", "arguments must be a map[string]interface{}")
}

func CombineOptions(options ...[]mcp.ToolOption) []mcp.ToolOption {
	var result []mcp.ToolOption
	for _, option := range options {
		result = append(result, option...)
	}
	return result
}

func ConvertToHash(params map[string]interface{}, key string, properties ...string) map[string]interface{} {
	keyStr := key
	nestedParams := make(map[string]interface{})

	for _, prop := range properties {
		propStr := prop
		paramKey := keyStr + "_" + propStr
		if val, ok := params[paramKey]; ok {
			nestedParams[propStr] = val
			delete(params, paramKey)
		}
	}

	if len(nestedParams) > 0 {
		params[keyStr] = nestedParams
	}

	return params
}

func CheckRequired(params map[string]interface{}, required ...string) (*mcp.CallToolResult, error) {
	required = append(required, "enterprise_id")
	for _, key := range required {
		if _, ok := params[key]; !ok {
			return mcp.NewToolResultError("Missing required parameter: " + key),
				NewParamError(key, "parameter is required")
		}
	}
	return nil, nil
}

// MCPActionResult reports a completed mutation that has no meaningful upstream
// payload, echoing the identifiers the caller passed so the result is verifiable.
func MCPActionResult(fields map[string]any) (*mcp.CallToolResult, error) {
	result := make(map[string]any, len(fields)+1)
	result["ok"] = true
	for key, value := range fields {
		result[key] = value
	}
	return MCPJSONResult(result)
}

// MCPListResult wraps a bare list in the {data,total_count} envelope used by
// most list tools so callers see one predictable shape.
func MCPListResult(data any) (*mcp.CallToolResult, error) {
	count := 0
	if list, ok := data.([]any); ok {
		count = len(list)
	}
	return MCPJSONResult(map[string]any{"data": data, "total_count": count})
}
