package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gitee.com/oschina/mcp-gitee-ent/operations/discovery"
	"gitee.com/oschina/mcp-gitee-ent/utils"
)

type agentEvalTask struct {
	Query     string
	Tool      string
	Arguments map[string]any
}

func TestAgentToolWorkflowEval(t *testing.T) {
	resetToolsetFlags(t)

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer apiServer.Close()

	previousAPIBase := utils.GetApiBase()
	previousToken := utils.GetGiteeAccessToken()
	utils.SetApiBase(apiServer.URL)
	utils.SetGiteeAccessToken("test-token")
	t.Cleanup(func() {
		utils.SetApiBase(previousAPIBase)
		utils.SetGiteeAccessToken(previousToken)
	})

	mcpClient := initializeInProcessClient(t)
	tasks := []agentEvalTask{
		{"里程碑列表", "list_milestones", map[string]any{"enterprise_id": 7}},
		{"工作项研发关联", "get_issue_development_links", map[string]any{"enterprise_id": 7, "issue_id": "I1", "type": "pull_requests"}},
		{"测试用例", "list_test_cases", map[string]any{"enterprise_id": 7, "program_id": 9}},
		{"创建测试用例", "create_test_case", map[string]any{"enterprise_id": 7, "program_id": 9, "title": "case", "module_id": 3, "case_steps": []any{map[string]any{"description": "step 1"}}}},
		{"测试计划", "list_test_plans", map[string]any{"enterprise_id": 7, "program_id": 9}},
		{"创建测试报告", "create_test_report", map[string]any{"enterprise_id": 7, "program_id": 9, "title": "report", "test_plan_ids": []any{1}, "components": []any{map[string]any{"ident": "basic-info", "title": "基础信息"}}}},
		{"仓库门禁", "list_project_check_runs", map[string]any{"enterprise_id": 7, "repository_id": 9}},
		{"项目概览", "get_program_overview", map[string]any{"enterprise_id": 7, "program_id": 9}},
		{"团队成员", "list_group_members", map[string]any{"enterprise_id": 7, "group_id": "team-a"}},
		{"文档搜索", "search_documents", map[string]any{"enterprise_id": 7, "search": "test"}},
		{"工作项附件", "list_issue_attachments", map[string]any{"enterprise_id": 7, "issue_id": 1}},
		{"工作项权限", "list_issue_authorizations", map[string]any{"enterprise_id": 7, "issue_id": 1}},
		{"删除测试计划", "delete_test_plan", map[string]any{"enterprise_id": 7, "program_id": 9, "test_plan_id": 1}},
		{"修改测试报告", "update_test_report", map[string]any{"enterprise_id": 7, "program_id": 9, "test_report_id": 1, "title": "updated"}},
		{"最近文档", "list_recent_documents", map[string]any{"enterprise_id": 7}},
		{"团队仓库", "list_group_projects", map[string]any{"enterprise_id": 7, "group_id": 9}},
		{"测试报告列表", "list_test_reports", map[string]any{"enterprise_id": 7, "program_id": 9}},
		{"工时", "list_workloads", map[string]any{"enterprise_id": 7, "start_date": "2026-10-01", "end_date": "2026-10-31"}},
		{"成员详情", "get_member", map[string]any{"enterprise_id": 7, "user_id": 1}},
		{"门禁详情", "get_check_run_annotations", map[string]any{"enterprise_id": 7, "repository_id": 9, "head_sha": "abc"}},
		{"星标工作项", "set_issue_starred", map[string]any{"enterprise_id": 7, "issue_id": 1, "starred": true}},
		{"关联文档", "link_issue_document", map[string]any{"enterprise_id": 7, "issue_id": 1, "doc_node_id": 2}},
		{"恢复工作项", "restore_issues", map[string]any{"enterprise_id": 7, "issue_ids": []any{1}}},
	}

	searchHits, inspectOK, invokeOK := 0, 0, 0
	for _, task := range tasks {
		searchResult := callTool(t, mcpClient, discovery.SearchTools, map[string]any{"query": task.Query})
		if !searchResult.IsError && strings.Contains(textContent(t, searchResult), task.Tool) {
			searchHits++
		} else {
			t.Errorf("search %q did not return %s: %s", task.Query, task.Tool, textContent(t, searchResult))
			continue
		}

		inspectResult := callTool(t, mcpClient, discovery.InspectTool, map[string]any{"name": task.Tool})
		inspectText := textContent(t, inspectResult)
		if inspectResult.IsError {
			t.Errorf("inspect %s failed: %s", task.Tool, inspectText)
		} else {
			var schema map[string]any
			if err := json.Unmarshal([]byte(inspectText), &schema); err == nil && schema["name"] == task.Tool {
				inspectOK++
			} else {
				t.Errorf("inspect %s returned invalid schema", task.Tool)
			}
		}

		invokeResult := callTool(t, mcpClient, discovery.InvokeTool, map[string]any{
			"name":      task.Tool,
			"arguments": task.Arguments,
		})
		if invokeResult.IsError {
			t.Errorf("invoke %s failed: %s", task.Tool, textContent(t, invokeResult))
		} else {
			invokeOK++
		}
	}

	total := len(tasks)
	t.Logf("agent eval: search=%d/%d inspect=%d/%d invoke=%d/%d round_trips=%d", searchHits, total, inspectOK, total, invokeOK, total, total*3)
	if searchHits != total || inspectOK != total || invokeOK != total {
		t.Fatalf("agent workflow eval failed")
	}
}
