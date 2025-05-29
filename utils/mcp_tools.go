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
