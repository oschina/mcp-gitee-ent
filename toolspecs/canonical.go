package toolspecs

import (
	"context"
	"fmt"
	"maps"
	"strings"

	"gitee.com/oschina/mcp-gitee-ent/catalog"
	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

type paramRename struct {
	Legacy      string
	Canonical   string
	Description string
	NewType     string
	NewItems    map[string]any
	ToLegacy    func(any) (any, error)
}

func canonicalizeSpec(spec catalog.Spec) (catalog.Spec, error) {
	rules := canonicalParamRules[spec.Tool.Name]
	tool := spec.Tool
	tool.InputSchema.Properties = cloneMap(spec.Tool.InputSchema.Properties)
	tool.InputSchema.Required = append([]string(nil), spec.Tool.InputSchema.Required...)

	for _, rule := range rules {
		property, ok := tool.InputSchema.Properties[rule.Legacy]
		if !ok {
			return catalog.Spec{}, fmt.Errorf("%s: legacy parameter %q is missing", spec.Tool.Name, rule.Legacy)
		}
		if rule.Canonical != rule.Legacy {
			if _, exists := tool.InputSchema.Properties[rule.Canonical]; exists {
				return catalog.Spec{}, fmt.Errorf("%s: canonical parameter %q already exists", spec.Tool.Name, rule.Canonical)
			}
			delete(tool.InputSchema.Properties, rule.Legacy)
		}
		updated, ok := cloneValue(property).(map[string]any)
		if !ok {
			return catalog.Spec{}, fmt.Errorf("%s: parameter %q is not an object schema", spec.Tool.Name, rule.Legacy)
		}
		if rule.Description != "" {
			updated["description"] = rule.Description
		}
		if rule.NewType != "" {
			updated["type"] = rule.NewType
		}
		if rule.NewItems != nil {
			updated["items"] = cloneValue(rule.NewItems)
		}
		if rule.Legacy == "project_id" && rule.Description == "" {
			updated["description"] = repositoryIDDescription(stringValue(updated["description"]))
		}
		if rule.Legacy == "project_ids" && rule.Description == "" {
			updated["description"] = repositoryIDsDescription(stringValue(updated["description"]))
		}
		tool.InputSchema.Properties[rule.Canonical] = updated
		for i, name := range tool.InputSchema.Required {
			if name == rule.Legacy {
				tool.InputSchema.Required[i] = rule.Canonical
			}
		}
	}

	for name, property := range tool.InputSchema.Properties {
		updated, ok := property.(map[string]any)
		if !ok {
			continue
		}
		switch name {
		case "program_id", "program_ids":
			updated["description"] = programDescription(name, stringValue(updated["description"]))
		case "pull_request_iid":
			updated["description"] = "Pull request IID: the display number shown in Gitee (for example !42), not the internal pull_request_id."
		case "pull_request_id":
			updated["description"] = pullRequestIDDescription(stringValue(updated["description"]))
		case "pull_request_ids":
			updated["description"] = "Comma-separated internal pull request IDs (not display IIDs)."
		}
	}

	spec.Tool = tool
	spec.ExampleArguments = canonicalExamples(spec.ExampleArguments, rules)
	if len(rules) > 0 {
		spec.Handler = canonicalHandler(spec.Handler, rules, tool.InputSchema.Required)
	}
	return spec, nil
}

func canonicalHandler(handler utils.OptionHandlerFunc, rules []paramRename, required []string) utils.OptionHandlerFunc {
	byCanonical := make(map[string]paramRename, len(rules))
	for _, rule := range rules {
		byCanonical[rule.Canonical] = rule
	}
	return func(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
		arguments, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			if request.Params.Arguments != nil {
				return mcp.NewToolResultError("arguments must be an object"), nil
			}
			arguments = map[string]any{}
		} else {
			copied := make(map[string]any, len(arguments))
			maps.Copy(copied, arguments)
			arguments = copied
		}

		for _, rule := range rules {
			if rule.Legacy == rule.Canonical {
				continue
			}
			if _, exists := arguments[rule.Legacy]; exists {
				return mcp.NewToolResultError(fmt.Sprintf(
					"Parameter %q was renamed to %q; use %q.", rule.Legacy, rule.Canonical, rule.Canonical)), nil
			}
		}
		for _, name := range required {
			if _, exists := arguments[name]; !exists {
				return mcp.NewToolResultError(fmt.Sprintf("Missing required parameter: %s.", name)), nil
			}
		}
		for _, rule := range rules {
			value, exists := arguments[rule.Canonical]
			if !exists {
				continue
			}
			if rule.ToLegacy != nil {
				converted, err := rule.ToLegacy(value)
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				value = converted
			}
			arguments[rule.Legacy] = value
			if rule.Legacy != rule.Canonical {
				delete(arguments, rule.Canonical)
			}
		}
		request.Params.Arguments = arguments
		return handler(ctx, request, opts...)
	}
}

func canonicalExamples(example map[string]any, rules []paramRename) map[string]any {
	if example == nil {
		return nil
	}
	out := cloneMap(example)
	for _, rule := range rules {
		value, exists := out[rule.Canonical]
		if !exists && rule.Legacy != rule.Canonical {
			value, exists = out[rule.Legacy]
			if !exists {
				continue
			}
			delete(out, rule.Legacy)
		}
		if !exists {
			continue
		}
		if rule.Legacy == "test_plan_ids" && rule.Legacy == rule.Canonical {
			if text, ok := value.(string); ok {
				parts := strings.Split(text, ",")
				items := make([]any, 0, len(parts))
				for _, part := range parts {
					if id, err := utils.SafelyConvertToInt(strings.TrimSpace(part)); err == nil {
						items = append(items, id)
					}
				}
				value = items
			}
		}
		out[rule.Canonical] = value
	}
	return out
}

func repositoryIDDescription(description string) string {
	if strings.Contains(description, "PathWithNamespace") {
		return "Repository ID or repository's PathWithNamespace."
	}
	if strings.Contains(strings.ToLower(description), "optional") {
		return "Optional repository ID."
	}
	return "Repository ID."
}

func repositoryIDsDescription(description string) string {
	if strings.Contains(strings.ToLower(description), "comma") {
		return "Comma-separated repository IDs."
	}
	return "Optional repository IDs to restrict the search to."
}

func programDescription(name, description string) string {
	if strings.Contains(description, "Gitee 项目") {
		return description
	}
	if name == "program_ids" {
		return strings.TrimSpace(strings.ReplaceAll(description, "program IDs", "program IDs (Gitee 项目 IDs)"))
	}
	description = strings.ReplaceAll(description, "Project ID", "Program ID (Gitee 项目 ID)")
	description = strings.ReplaceAll(description, "project ID", "program ID (Gitee 项目 ID)")
	if strings.Contains(description, "Gitee 项目") {
		return description
	}
	if strings.Contains(description, "Program ID") {
		return strings.Replace(description, "Program ID", "Program ID (Gitee 项目 ID)", 1)
	}
	if strings.Contains(description, "program ID") {
		return strings.Replace(description, "program ID", "program ID (Gitee 项目 ID)", 1)
	}
	return description
}

func pullRequestIDDescription(description string) string {
	if strings.Contains(strings.ToLower(description), "optional") {
		return "Optional internal pull request ID (database ID), not the display IID."
	}
	return "Internal pull request ID (database ID), not the display IID."
}

func stringValue(value any) string {
	text, _ := value.(string)
	return text
}

func cloneMap(input map[string]any) map[string]any {
	if input == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(input))
	for key, value := range input {
		out[key] = cloneValue(value)
	}
	return out
}

func cloneValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneMap(typed)
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = cloneValue(item)
		}
		return out
	default:
		return value
	}
}
