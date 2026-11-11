package toolspecs

import (
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/catalog"
	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

// domainAliases are indexed by the lexical searcher so that business language
// can find tools even when the tool name uses different terminology.
var domainAliases = map[string][]string{
	"documents":      {"doc", "docs", "wiki", "文档", "知识库"},
	"issues":         {"issue", "task", "bug", "任务", "需求", "缺陷", "工作项"},
	"repository":     {"repo", "repository", "仓库", "代码库", "release", "发布", "分支", "branch", "branches", "文件", "file", "文件内容"},
	"pulls":          {"pr", "mr", "pull request", "merge request", "合并请求", "代码评审"},
	"enterprises":    {"enterprise", "企业"},
	"labels":         {"label", "标签"},
	"issue_types":    {"issue type", "工作项类型", "任务类型", "自定义字段", "custom field", "custom fields"},
	"issue_states":   {"issue state", "工作项状态", "任务状态", "状态流转"},
	"user":           {"user", "profile", "用户", "个人信息"},
	"programs":       {"program", "项目"},
	"scrum_sprints":  {"sprint", "迭代"},
	"scrum_versions": {"version", "版本"},
	"members":        {"member", "成员", "用户"},
	"groups":         {"group", "team", "团队", "分组"},
	"week_reports":   {"week report", "weekly report", "周报", "周报详情"},
}

// visibleSpec registers a tool that is exposed directly on tools/list.
func visibleSpec(tool mcp.Tool, handler utils.OptionHandlerFunc, domain string, risk catalog.Risk) catalog.Spec {
	return catalog.Spec{
		Tool:         tool,
		Handler:      handler,
		Domain:       domain,
		Tags:         []string{domain},
		Aliases:      append([]string(nil), domainAliases[domain]...),
		Risk:         risk,
		Verification: catalog.VerificationNone,
		ReadOnly:     risk == catalog.RiskRead,
		Visible:      true,
	}
}

// hiddenSpec registers a tool that is discoverable only via search_tools/invoke_tool.
func hiddenSpec(tool mcp.Tool, handler utils.OptionHandlerFunc, domain string, risk catalog.Risk, aliases []string, summary string, example map[string]any) catalog.Spec {
	return catalog.Spec{
		Tool:             tool,
		Handler:          handler,
		Domain:           domain,
		Tags:             []string{domain},
		Aliases:          aliases,
		Summary:          summary,
		ExampleArguments: example,
		Risk:             risk,
		ReadOnly:         risk == catalog.RiskRead,
		Invokable:        true,
	}
}

func All() []catalog.Spec {
	specs := issueSpecs()
	specs = append(specs, repoSpecs()...)
	specs = append(specs, docSpecs()...)
	specs = append(specs, orgSpecs()...)
	specs = append(specs, testingSpecs()...)
	return specs
}

func Build(toolAllowed func(string) bool) (*catalog.Registry, error) {
	specs := All()
	for i := range specs {
		canonical, err := canonicalizeSpec(specs[i])
		if err != nil {
			return nil, fmt.Errorf("canonicalize %s: %w", specs[i].Tool.Name, err)
		}
		specs[i] = canonical
		allowed := true
		if toolAllowed != nil {
			allowed = toolAllowed(specs[i].Tool.Name)
		}
		specs[i].Invokable = allowed
		if specs[i].Visible {
			specs[i].Visible = allowed
		}
	}
	registry := catalog.NewRegistry()
	if err := registry.Register(specs...); err != nil {
		return nil, err
	}
	return registry, nil
}
