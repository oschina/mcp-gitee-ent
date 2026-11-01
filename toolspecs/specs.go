package toolspecs

import (
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/catalog"
	"gitee.com/oschina/mcp-gitee-ent/operations/baselines"
	"gitee.com/oschina/mcp-gitee-ent/operations/check_runs"
	"gitee.com/oschina/mcp-gitee-ent/operations/common_views"
	"gitee.com/oschina/mcp-gitee-ent/operations/documents"
	"gitee.com/oschina/mcp-gitee-ent/operations/enterprises"
	"gitee.com/oschina/mcp-gitee-ent/operations/groups"
	"gitee.com/oschina/mcp-gitee-ent/operations/issue_links"
	"gitee.com/oschina/mcp-gitee-ent/operations/issue_recycle_bin"
	"gitee.com/oschina/mcp-gitee-ent/operations/issue_states"
	"gitee.com/oschina/mcp-gitee-ent/operations/issue_types"
	"gitee.com/oschina/mcp-gitee-ent/operations/issues"
	"gitee.com/oschina/mcp-gitee-ent/operations/kanbans"
	"gitee.com/oschina/mcp-gitee-ent/operations/labels"
	"gitee.com/oschina/mcp-gitee-ent/operations/members"
	"gitee.com/oschina/mcp-gitee-ent/operations/milestones"
	"gitee.com/oschina/mcp-gitee-ent/operations/programs"
	"gitee.com/oschina/mcp-gitee-ent/operations/pulls"
	"gitee.com/oschina/mcp-gitee-ent/operations/repository"
	"gitee.com/oschina/mcp-gitee-ent/operations/scrum_sprints"
	"gitee.com/oschina/mcp-gitee-ent/operations/scrum_versions"
	"gitee.com/oschina/mcp-gitee-ent/operations/test_cases"
	"gitee.com/oschina/mcp-gitee-ent/operations/test_modules"
	"gitee.com/oschina/mcp-gitee-ent/operations/test_plans"
	"gitee.com/oschina/mcp-gitee-ent/operations/test_reports"
	"gitee.com/oschina/mcp-gitee-ent/operations/user"
	"gitee.com/oschina/mcp-gitee-ent/operations/webhooks"
	"gitee.com/oschina/mcp-gitee-ent/operations/workloads"
	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

// domainAliases are indexed by the lexical searcher so that business language
// can find tools even when the tool name uses different terminology.
var domainAliases = map[string][]string{
	"documents":      {"doc", "docs", "wiki", "文档", "知识库"},
	"issues":         {"issue", "task", "bug", "任务", "需求", "缺陷", "工作项"},
	"repository":     {"repo", "repository", "仓库", "代码库", "release", "发布"},
	"pulls":          {"pr", "mr", "pull request", "merge request", "合并请求", "代码评审"},
	"enterprises":    {"enterprise", "企业"},
	"labels":         {"label", "标签"},
	"issue_types":    {"issue type", "工作项类型", "任务类型"},
	"issue_states":   {"issue state", "工作项状态", "任务状态", "状态流转"},
	"user":           {"user", "profile", "用户", "个人信息"},
	"programs":       {"program", "项目"},
	"scrum_sprints":  {"sprint", "迭代"},
	"scrum_versions": {"version", "版本"},
	"members":        {"member", "成员", "用户"},
	"groups":         {"group", "team", "团队", "分组"},
}

func existingToolSpec(tool mcp.Tool, handler utils.OptionHandlerFunc, domain string, risk catalog.Risk) catalog.Spec {
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

func existingToolSpecs() []catalog.Spec {
	return []catalog.Spec{
		// Documents
		existingToolSpec(documents.CreateDocumentTool, documents.CreateDocumentHandleFunc, "documents", catalog.RiskWrite),
		existingToolSpec(documents.CreateDocumentNodeTool, documents.CreateDocumentNodeHandleFunc, "documents", catalog.RiskWrite),
		existingToolSpec(documents.RenameDocumentFolderTool, documents.RenameDocumentFolderHandleFunc, "documents", catalog.RiskWrite),
		existingToolSpec(documents.ListDocumentsTool, documents.ListDocumentsHandleFunc, "documents", catalog.RiskRead),
		existingToolSpec(documents.ListDocumentPagesTool, documents.ListDocumentPagesHandleFunc, "documents", catalog.RiskRead),
		existingToolSpec(documents.GetDocumentPageTool, documents.GetDocumentPageHandleFunc, "documents", catalog.RiskRead),
		existingToolSpec(documents.UploadDocumentTool, documents.UploadDocumentHandleFunc, "documents", catalog.RiskWrite),
		existingToolSpec(documents.DownloadDocumentTool, documents.DownloadDocumentHandleFunc, "documents", catalog.RiskRead),
		existingToolSpec(documents.UpdateDocumentPageTool, documents.UpdateDocumentPageHandleFunc, "documents", catalog.RiskWrite),

		// Issues
		existingToolSpec(issues.ListIssuesTool, issues.ListIssuesHandleFunc, "issues", catalog.RiskRead),
		existingToolSpec(issues.CreateIssueTool, issues.CreateIssueHandleFunc, "issues", catalog.RiskWrite),
		existingToolSpec(issues.GetIssueDetailTool, issues.GetIssueDetailHandleFunc, "issues", catalog.RiskRead),
		existingToolSpec(issues.UpdateIssueTool, issues.UpdateIssueHandleFunc, "issues", catalog.RiskWrite),
		existingToolSpec(issues.CommentIssueTool, issues.CommentIssueHandleFunc, "issues", catalog.RiskWrite),
		existingToolSpec(issues.ListIssueCommentsTool, issues.ListIssueCommentsHandleFunc, "issues", catalog.RiskRead),

		// Repositories
		existingToolSpec(repository.ListRepositoriesTool, repository.ListRepositoriesHandleFunc, "repository", catalog.RiskRead),
		existingToolSpec(repository.CreateRepositoryTool, repository.CreateRepositoryHandleFunc, "repository", catalog.RiskWrite),
		existingToolSpec(repository.CreateReleaseTool, repository.CreateReleaseHandleFunc, "repository", catalog.RiskWrite),
		existingToolSpec(repository.ListReleasesTool, repository.ListReleasesHandleFunc, "repository", catalog.RiskRead),
		existingToolSpec(repository.GetRepoTreeTool, repository.GetRepoTreeHandleFunc, "repository", catalog.RiskRead),
		existingToolSpec(repository.GetFileContentTool, repository.GetFileContentHandleFunc, "repository", catalog.RiskRead),

		// Pull requests
		existingToolSpec(pulls.ListEntPullsTool, pulls.ListEntPullsHandleFunc, "pulls", catalog.RiskRead),
		existingToolSpec(pulls.CreateEntPullTool, pulls.CreateEntPullHandleFunc, "pulls", catalog.RiskWrite),
		existingToolSpec(pulls.GetEntPullDetailTool, pulls.GetPullDetailHandleFunc, "pulls", catalog.RiskRead),
		existingToolSpec(pulls.GetPullDiffTool, pulls.GetPullDiffHandleFunc, "pulls", catalog.RiskRead),
		existingToolSpec(pulls.CommentPullTool, pulls.CommentPullHandleFunc, "pulls", catalog.RiskWrite),
		existingToolSpec(pulls.ListPullCommentsTool, pulls.ListPullCommentsHandleFunc, "pulls", catalog.RiskRead),
		existingToolSpec(pulls.MergePullTool, pulls.MergePullHandleFunc, "pulls", catalog.RiskWrite),
		existingToolSpec(pulls.UpdatePullTool, pulls.UpdateEntPullHandleFunc, "pulls", catalog.RiskWrite),

		// Enterprise metadata
		existingToolSpec(enterprises.ListEnterprisesTool, enterprises.ListEnterprisesHandleFunc, "enterprises", catalog.RiskRead),
		existingToolSpec(labels.ListEnterpriseLabelsTool, labels.ListEnterpriseLabelsHandleFunc, "labels", catalog.RiskRead),
		existingToolSpec(issue_types.ListIssueTypesTool, issue_types.ListIssueTypesHandleFunc, "issue_types", catalog.RiskRead),
		existingToolSpec(issue_states.ListIssueTypeStatesTool, issue_states.ListIssueTypeStatesHandleFunc, "issue_states", catalog.RiskRead),
		existingToolSpec(user.GetUserInfoTool, user.GetUserInfoHandleFunc, "user", catalog.RiskRead),
		existingToolSpec(programs.ListProgramsTool, programs.ListProgramsHandleFunc, "programs", catalog.RiskRead),

		// Scrum
		existingToolSpec(scrum_sprints.CreateScrumSprintTool, scrum_sprints.CreateScrumSprintHandleFunc, "scrum_sprints", catalog.RiskWrite),
		existingToolSpec(scrum_sprints.ListScrumSprintsTool, scrum_sprints.ListScrumSprintsHandleFunc, "scrum_sprints", catalog.RiskRead),
		existingToolSpec(scrum_versions.ListScrumVersionsTool, scrum_versions.ListScrumVersionsHandleFunc, "scrum_versions", catalog.RiskRead),

		// Organization
		existingToolSpec(members.ListEntMembersTool, members.ListEntMembersHandleFunc, "members", catalog.RiskRead),
		existingToolSpec(groups.ListEntGroupsTool, groups.ListEntGroupsHandleFunc, "groups", catalog.RiskRead),
	}
}

func pilotToolSpecs() []catalog.Spec {
	return []catalog.Spec{
		{
			Tool:             issue_links.GetIssueOverviewTool,
			Handler:          issue_links.GetIssueOverviewHandleFunc,
			Domain:           "issue_links",
			Tags:             []string{"issues", "relations"},
			Aliases:          []string{"工作项概览", "需求概览", "issue overview"},
			Summary:          "Get an issue overview for planning and navigation.",
			ExampleArguments: map[string]any{"enterprise_id": 123, "issue_id": "I123"},
			Risk:             catalog.RiskRead,
			ReadOnly:         true,
			Invokable:        true,
		},
		{
			Tool:             issue_links.ListIssueChildrenTool,
			Handler:          issue_links.ListIssueChildrenHandleFunc,
			Domain:           "issue_links",
			Tags:             []string{"issues", "hierarchy"},
			Aliases:          []string{"子工作项", "子任务", "有哪些子任务", "需求下面还有哪些任务", "children"},
			Summary:          "List child work items of an issue.",
			ExampleArguments: map[string]any{"enterprise_id": 123, "issue_id": "I123", "page": 1, "per_page": 20},
			Risk:             catalog.RiskRead,
			ReadOnly:         true,
			Invokable:        true,
		},
		{
			Tool:             issue_links.ListIssueAncestorsTool,
			Handler:          issue_links.ListIssueAncestorsHandleFunc,
			Domain:           "issue_links",
			Tags:             []string{"issues", "hierarchy"},
			Aliases:          []string{"祖先工作项", "父需求", "ancestors"},
			Summary:          "List ancestor work items of an issue.",
			ExampleArguments: map[string]any{"enterprise_id": 123, "issue_id": "I123", "scope": "part"},
			Risk:             catalog.RiskRead,
			ReadOnly:         true,
			Invokable:        true,
		},
		{
			Tool:             milestones.ListMilestonesTool,
			Handler:          milestones.ListMilestonesHandleFunc,
			Domain:           "milestones",
			Tags:             []string{"planning"},
			Aliases:          []string{"里程碑", "milestone", "milestones"},
			Summary:          "List milestones or search them by keyword.",
			ExampleArguments: map[string]any{"enterprise_id": 123, "search": "Q4", "page": 1, "per_page": 20},
			Risk:             catalog.RiskRead,
			ReadOnly:         true,
			Invokable:        true,
		},
		{
			Tool:             milestones.GetMilestoneTool,
			Handler:          milestones.GetMilestoneHandleFunc,
			Domain:           "milestones",
			Tags:             []string{"planning"},
			Aliases:          []string{"里程碑详情", "milestone detail"},
			Summary:          "Get one milestone by ID.",
			ExampleArguments: map[string]any{"enterprise_id": 123, "milestone_id": 456},
			Risk:             catalog.RiskRead,
			ReadOnly:         true,
			Invokable:        true,
		},
	}
}

func pilotToolSpecsPhase2() []catalog.Spec {
	return []catalog.Spec{
		{
			Tool:             issue_links.GetIssueDevelopmentLinksTool,
			Handler:          issue_links.GetIssueDevelopmentLinksHandleFunc,
			Domain:           "issue_links",
			Tags:             []string{"issues", "relations", "development"},
			Aliases:          []string{"研发关联", "关联代码", "关联PR", "关联的PR", "工作项关联的PR", "这个需求关联了哪些代码", "需求关联代码", "data links", "linked pull requests"},
			Summary:          "Get one category of development links for an issue.",
			ExampleArguments: map[string]any{"enterprise_id": 123, "issue_id": "I123", "type": "pull_requests"},
			Risk:             catalog.RiskRead,
			ReadOnly:         true,
			Invokable:        true,
		},
		{
			Tool:             issue_links.ListIssueOperationLogsTool,
			Handler:          issue_links.ListIssueOperationLogsHandleFunc,
			Domain:           "issue_links",
			Tags:             []string{"issues", "history"},
			Aliases:          []string{"操作日志", "变更历史", "修改记录", "工作项操作日志", "工作项变更历史", "activity log", "operation history"},
			Summary:          "List operation history for an issue.",
			ExampleArguments: map[string]any{"enterprise_id": 123, "issue_id": "I123", "limit": 20},
			Risk:             catalog.RiskRead,
			ReadOnly:         true,
			Invokable:        true,
		},
		{
			Tool:             issue_links.ListIssueNextStatesTool,
			Handler:          issue_links.ListIssueNextStatesHandleFunc,
			Domain:           "issue_links",
			Tags:             []string{"issues", "workflow"},
			Aliases:          []string{"下一状态", "状态流转", "next states", "available transitions"},
			Summary:          "List states to which this issue can currently transition.",
			ExampleArguments: map[string]any{"enterprise_id": 123, "issue_id": "I123"},
			Risk:             catalog.RiskRead,
			ReadOnly:         true,
			Invokable:        true,
		},
		{
			Tool:             members.SearchMembersTool,
			Handler:          members.SearchMembersHandleFunc,
			Domain:           "members",
			Tags:             []string{"organization"},
			Aliases:          []string{"搜索成员", "找人", "find member", "member search"},
			Summary:          "Search enterprise members.",
			ExampleArguments: map[string]any{"enterprise_id": 123, "search": "alice", "page": 1, "per_page": 20},
			Risk:             catalog.RiskRead,
			ReadOnly:         true,
			Invokable:        true,
		},
		{
			Tool:             members.GetMemberTool,
			Handler:          members.GetMemberHandleFunc,
			Domain:           "members",
			Tags:             []string{"organization"},
			Aliases:          []string{"成员详情", "用户详情", "member detail"},
			Summary:          "Get one enterprise member by user ID or username.",
			ExampleArguments: map[string]any{"enterprise_id": 123, "user_id": "alice"},
			Risk:             catalog.RiskRead,
			ReadOnly:         true,
			Invokable:        true,
		},
		{
			Tool:             members.GetMemberByEmailTool,
			Handler:          members.GetMemberByEmailHandleFunc,
			Domain:           "members",
			Tags:             []string{"organization"},
			Aliases:          []string{"邮箱找人", "按邮箱查询成员", "find member by email"},
			Summary:          "Find an enterprise member by exact email address.",
			ExampleArguments: map[string]any{"enterprise_id": 123, "email": "alice@example.com"},
			Risk:             catalog.RiskRead,
			ReadOnly:         true,
			Invokable:        true,
		},
		{
			Tool:             members.ListMemberMembershipsTool,
			Handler:          members.ListMemberMembershipsHandleFunc,
			Domain:           "members",
			Tags:             []string{"organization", "relations"},
			Aliases:          []string{"成员归属", "参与项目", "参与仓库", "参与团队", "参与的项目", "参与的团队", "成员参与的项目", "成员参与了哪些团队", "参与了哪些团队", "参与了哪些项目", "memberships"},
			Summary:          "List one membership collection for a member: repositories, programs, or groups.",
			ExampleArguments: map[string]any{"enterprise_id": 123, "user_id": "alice", "type": "programs"},
			Risk:             catalog.RiskRead,
			ReadOnly:         true,
			Invokable:        true,
		},
		{
			Tool:             members.GetGroupTool,
			Handler:          members.GetGroupHandleFunc,
			Domain:           "groups",
			Tags:             []string{"organization"},
			Aliases:          []string{"团队详情", "group detail", "team detail"},
			Summary:          "Get enterprise team details by group ID or path.",
			ExampleArguments: map[string]any{"enterprise_id": 123, "group_id": 456},
			Risk:             catalog.RiskRead,
			ReadOnly:         true,
			Invokable:        true,
		},
		{
			Tool:             workloads.ListWorkloadsTool,
			Handler:          workloads.ListWorkloadsHandleFunc,
			Domain:           "workloads",
			Tags:             []string{"reporting"},
			Aliases:          []string{"工时", "工作量", "worklog", "time entries"},
			Summary:          "List workload entries within an explicit date range.",
			ExampleArguments: map[string]any{"enterprise_id": 123, "start_date": "2026-10-01", "end_date": "2026-10-31"},
			Risk:             catalog.RiskRead,
			ReadOnly:         true,
			Invokable:        true,
		},
	}
}

func pilotWriteSpecs() []catalog.Spec {
	return []catalog.Spec{
		{
			Tool:             milestones.CreateMilestoneTool,
			Handler:          milestones.CreateMilestoneHandleFunc,
			Domain:           "milestones",
			Tags:             []string{"planning", "write"},
			Aliases:          []string{"创建里程碑", "新建里程碑", "new milestone"},
			Summary:          "Create a milestone.",
			ExampleArguments: map[string]any{"enterprise_id": 123, "title": "2026 Q4", "due_date": "2026-12-31"},
			Risk:             catalog.RiskWrite,
			Invokable:        true,
		},
		{
			Tool:             milestones.UpdateMilestoneTool,
			Handler:          milestones.UpdateMilestoneHandleFunc,
			Domain:           "milestones",
			Tags:             []string{"planning", "write"},
			Aliases:          []string{"编辑里程碑", "修改里程碑", "update milestone"},
			Summary:          "Update mutable fields of one milestone.",
			ExampleArguments: map[string]any{"enterprise_id": 123, "milestone_id": 456, "due_date": "2026-12-31"},
			Risk:             catalog.RiskWrite,
			Idempotent:       true,
			Invokable:        true,
		},
		{
			Tool:             labels.CreateLabelTool,
			Handler:          labels.CreateLabelHandleFunc,
			Domain:           "labels",
			Tags:             []string{"write"},
			Aliases:          []string{"创建标签", "新增标签", "new label"},
			Summary:          "Create an enterprise label.",
			ExampleArguments: map[string]any{"enterprise_id": 123, "name": "backend", "color": "#1E90FF"},
			Risk:             catalog.RiskWrite,
			Invokable:        true,
		},
		{
			Tool:             labels.UpdateLabelTool,
			Handler:          labels.UpdateLabelHandleFunc,
			Domain:           "labels",
			Tags:             []string{"write"},
			Aliases:          []string{"编辑标签", "修改标签", "update label"},
			Summary:          "Update an existing label.",
			ExampleArguments: map[string]any{"enterprise_id": 123, "label_id": 456, "name": "backend", "color": "#1E90FF"},
			Risk:             catalog.RiskWrite,
			Idempotent:       true,
			Invokable:        true,
		},
		{
			Tool:             labels.DeleteLabelTool,
			Handler:          labels.DeleteLabelHandleFunc,
			Domain:           "labels",
			Tags:             []string{"write", "destructive"},
			Aliases:          []string{"删除标签", "移除标签", "delete label"},
			Summary:          "Delete an enterprise label that no work item uses.",
			ExampleArguments: map[string]any{"enterprise_id": 123, "label_id": 456},
			Risk:             catalog.RiskDestructive,
			Invokable:        true,
		},
		{
			Tool:             milestones.DeleteMilestoneTool,
			Handler:          milestones.DeleteMilestoneHandleFunc,
			Domain:           "milestones",
			Tags:             []string{"planning", "write", "destructive"},
			Aliases:          []string{"删除里程碑", "移除里程碑", "delete milestone"},
			Summary:          "Delete one milestone.",
			ExampleArguments: map[string]any{"enterprise_id": 123, "milestone_id": 456},
			Risk:             catalog.RiskDestructive,
			Invokable:        true,
		},
		{
			Tool:             issue_links.UpdateIssueNoteTool,
			Handler:          issue_links.UpdateIssueNoteHandleFunc,
			Domain:           "issue_links",
			Tags:             []string{"issues", "write"},
			Aliases:          []string{"修改评论", "编辑评论", "update comment"},
			Summary:          "Edit one issue comment.",
			ExampleArguments: map[string]any{"enterprise_id": 123, "issue_id": "I123", "note_id": 456, "body": "Updated comment"},
			Risk:             catalog.RiskWrite,
			Idempotent:       true,
			Invokable:        true,
		},
		{
			Tool:             issue_links.DeleteIssueNoteTool,
			Handler:          issue_links.DeleteIssueNoteHandleFunc,
			Domain:           "issue_links",
			Tags:             []string{"issues", "destructive"},
			Aliases:          []string{"删除评论", "移除评论", "delete comment"},
			Summary:          "Delete one issue comment.",
			ExampleArguments: map[string]any{"enterprise_id": 123, "issue_id": "I123", "note_id": 456},
			Risk:             catalog.RiskDestructive,
			Idempotent:       true,
			Invokable:        true,
		},
		{
			Tool:             issue_links.LinkIssuePullRequestTool,
			Handler:          issue_links.LinkIssuePullRequestHandleFunc,
			Domain:           "issue_links",
			Tags:             []string{"issues", "relations", "write"},
			Aliases:          []string{"关联PR", "关联合并请求", "link pull request"},
			Summary:          "Associate an existing pull request with an issue.",
			ExampleArguments: map[string]any{"enterprise_id": 123, "issue_id": "I123", "pull_request_id": 789},
			Risk:             catalog.RiskWrite,
			Invokable:        true,
		},
		{
			Tool:             issue_links.UnlinkIssuePullRequestTool,
			Handler:          issue_links.UnlinkIssuePullRequestHandleFunc,
			Domain:           "issue_links",
			Tags:             []string{"issues", "relations", "write"},
			Aliases:          []string{"取消关联PR", "解除PR关联", "unlink pull request"},
			Summary:          "Remove the association between an issue and a pull request.",
			ExampleArguments: map[string]any{"enterprise_id": 123, "issue_id": "I123", "pull_request_id": 789},
			Risk:             catalog.RiskWrite,
			Idempotent:       true,
			Invokable:        true,
		},
	}
}

func testSpec(tool mcp.Tool, handler utils.OptionHandlerFunc, domain string, risk catalog.Risk, aliases []string, summary string, example map[string]any) catalog.Spec {
	return catalog.Spec{
		Tool:             tool,
		Handler:          handler,
		Domain:           domain,
		Tags:             []string{"testing", domain},
		Aliases:          aliases,
		Summary:          summary,
		ExampleArguments: example,
		Risk:             risk,
		ReadOnly:         risk == catalog.RiskRead,
		Invokable:        true,
	}
}

func testToolSpecs() []catalog.Spec {
	return []catalog.Spec{
		testSpec(test_modules.ListTestCaseModulesTool, test_modules.ListTestCaseModulesHandleFunc, "test_modules", catalog.RiskRead,
			[]string{"用例模块", "功能模块", "测试模块", "test module", "test case module"},
			"List test case modules under a project.", map[string]any{"enterprise_id": 123, "program_id": 456}),
		testSpec(test_modules.GetTestCaseModuleTreeTool, test_modules.GetTestCaseModuleTreeHandleFunc, "test_modules", catalog.RiskRead,
			[]string{"模块树", "用例目录", "module tree", "test module tree"},
			"Get the test case module tree.", map[string]any{"enterprise_id": 123, "program_id": 456}),
		testSpec(test_modules.CreateTestCaseModuleTool, test_modules.CreateTestCaseModuleHandleFunc, "test_modules", catalog.RiskWrite,
			[]string{"创建用例模块", "新建功能模块", "create test module"},
			"Create a test case module.", map[string]any{"enterprise_id": 123, "program_id": 456, "name": "Login"}),
		testSpec(test_modules.UpdateTestCaseModuleTool, test_modules.UpdateTestCaseModuleHandleFunc, "test_modules", catalog.RiskWrite,
			[]string{"修改用例模块", "重命名功能模块", "update test module"},
			"Rename a test case module.", map[string]any{"enterprise_id": 123, "program_id": 456, "case_module_id": 789, "name": "Auth"}),
		testSpec(test_modules.MoveTestCaseModuleTool, test_modules.MoveTestCaseModuleHandleFunc, "test_modules", catalog.RiskWrite,
			[]string{"移动用例模块", "move test module"},
			"Move a test case module.", map[string]any{"enterprise_id": 123, "program_id": 456, "case_module_id": 789, "direction": "inside", "target_id": 790}),

		testSpec(test_cases.ListTestCasesTool, test_cases.ListTestCasesHandleFunc, "test_cases", catalog.RiskRead,
			[]string{"测试用例", "用例列表", "test case", "test cases"},
			"List test cases in a project.", map[string]any{"enterprise_id": 123, "program_id": 456, "search": "login"}),
		testSpec(test_cases.GetTestCaseTool, test_cases.GetTestCaseHandleFunc, "test_cases", catalog.RiskRead,
			[]string{"用例详情", "测试用例详情", "test case detail"},
			"Get one test case by ID.", map[string]any{"enterprise_id": 123, "program_id": 456, "test_case_id": 789}),
		testSpec(test_cases.CreateTestCaseTool, test_cases.CreateTestCaseHandleFunc, "test_cases", catalog.RiskWrite,
			[]string{"创建测试用例", "新建用例", "create test case"},
			"Create a test case with required steps.",
			map[string]any{"enterprise_id": 123, "program_id": 456, "title": "Login succeeds", "module_id": 789,
				"case_steps": []any{map[string]any{"description": "Open the login page", "expected_result": "Login form is shown"}}}),
		testSpec(test_cases.UpdateTestCaseTool, test_cases.UpdateTestCaseHandleFunc, "test_cases", catalog.RiskWrite,
			[]string{"修改测试用例", "编辑用例", "update test case"},
			"Update a test case.", map[string]any{"enterprise_id": 123, "program_id": 456, "test_case_id": 789, "priority": 1}),
		testSpec(test_cases.ArchiveTestCaseTool, test_cases.ArchiveTestCaseHandleFunc, "test_cases", catalog.RiskWrite,
			[]string{"归档测试用例", "archive test case"},
			"Archive a test case.", map[string]any{"enterprise_id": 123, "program_id": 456, "test_case_id": 789}),
		testSpec(test_cases.RestoreTestCaseTool, test_cases.RestoreTestCaseHandleFunc, "test_cases", catalog.RiskWrite,
			[]string{"恢复测试用例", "restore test case"},
			"Restore an archived test case.", map[string]any{"enterprise_id": 123, "program_id": 456, "test_case_id": 789}),

		testSpec(test_plans.ListTestPlansTool, test_plans.ListTestPlansHandleFunc, "test_plans", catalog.RiskRead,
			[]string{"测试计划", "计划列表", "test plan", "test plans"},
			"List test plans in a project.", map[string]any{"enterprise_id": 123, "program_id": 456}),
		testSpec(test_plans.GetTestPlanTool, test_plans.GetTestPlanHandleFunc, "test_plans", catalog.RiskRead,
			[]string{"测试计划详情", "test plan detail"},
			"Get one test plan by ID.", map[string]any{"enterprise_id": 123, "program_id": 456, "test_plan_id": 789}),
		testSpec(test_plans.CreateTestPlanTool, test_plans.CreateTestPlanHandleFunc, "test_plans", catalog.RiskWrite,
			[]string{"创建测试计划", "新建测试计划", "create test plan"},
			"Create a test plan.", map[string]any{"enterprise_id": 123, "program_id": 456, "title": "Release test", "assignee_id": 1001}),
		testSpec(test_plans.UpdateTestPlanTool, test_plans.UpdateTestPlanHandleFunc, "test_plans", catalog.RiskWrite,
			[]string{"修改测试计划", "编辑测试计划", "update test plan"},
			"Update a test plan.", map[string]any{"enterprise_id": 123, "program_id": 456, "test_plan_id": 789, "title": "Release test", "assignee_id": 1001}),
		testSpec(test_plans.UpdateTestPlanStateTool, test_plans.UpdateTestPlanStateHandleFunc, "test_plans", catalog.RiskWrite,
			[]string{"测试计划状态", "变更计划状态", "test plan state"},
			"Change a test plan state.", map[string]any{"enterprise_id": 123, "program_id": 456, "test_plan_id": 789, "state": 1}),
		testSpec(test_plans.ListTestPlanCasesTool, test_plans.ListTestPlanCasesHandleFunc, "test_plans", catalog.RiskRead,
			[]string{"计划内用例", "计划用例", "执行用例", "计划里的用例", "test plan cases"},
			"List test cases in a test plan.", map[string]any{"enterprise_id": 123, "program_id": 456, "test_plan_id": 789}),
		testSpec(test_plans.GetTestPlanCaseTool, test_plans.GetTestPlanCaseHandleFunc, "test_plans", catalog.RiskRead,
			[]string{"计划用例详情", "test plan case detail"},
			"Get one test case in a test plan.", map[string]any{"enterprise_id": 123, "program_id": 456, "test_plan_id": 789, "test_case_id": 101}),
		testSpec(test_plans.AddTestCasesToPlanTool, test_plans.AddTestCasesToPlanHandleFunc, "test_plans", catalog.RiskWrite,
			[]string{"添加用例到计划", "添加用例到测试计划", "计划导入用例", "add cases to plan"},
			"Add existing test cases to a test plan.", map[string]any{"enterprise_id": 123, "program_id": 456, "test_plan_id": 789, "case_ids": []any{101, 102}}),
		testSpec(test_plans.CreateTestCaseInPlanTool, test_plans.CreateTestCaseInPlanHandleFunc, "test_plans", catalog.RiskWrite,
			[]string{"在计划中创建用例", "新建计划用例", "create case in plan"},
			"Create a new test case inside a test plan with required steps.",
			map[string]any{"enterprise_id": 123, "program_id": 456, "test_plan_id": 789, "title": "Plan case", "module_id": 11,
				"case_steps": []any{map[string]any{"description": "Open the login page", "expected_result": "Login form is shown"}}}),
		testSpec(test_plans.UpdateTestPlanCaseTool, test_plans.UpdateTestPlanCaseHandleFunc, "test_plans", catalog.RiskWrite,
			[]string{"修改计划用例", "设置用例负责人", "update plan case"},
			"Update a test case assignee in a test plan.", map[string]any{"enterprise_id": 123, "program_id": 456, "test_plan_id": 789, "test_case_id": 101, "assignee_id": 1001}),
		testSpec(test_plans.LinkTestPlanCaseIssuesTool, test_plans.LinkTestPlanCaseIssuesHandleFunc, "test_plans", catalog.RiskWrite,
			[]string{"用例关联工作项", "关联缺陷", "link issues"},
			"Link work items to a test case in a test plan.", map[string]any{"enterprise_id": 123, "program_id": 456, "test_plan_id": 789, "test_case_id": 101, "issue_ids": []any{1, 2}}),
		testSpec(test_plans.UnlinkTestPlanCaseIssuesTool, test_plans.UnlinkTestPlanCaseIssuesHandleFunc, "test_plans", catalog.RiskWrite,
			[]string{"取消用例关联工作项", "取消关联缺陷", "unlink issues"},
			"Unlink work items from a test case in a test plan.", map[string]any{"enterprise_id": 123, "program_id": 456, "test_plan_id": 789, "test_case_id": 101, "issue_ids": []any{1}}),
		testSpec(test_plans.ListTestPlanCaseIssuesTool, test_plans.ListTestPlanCaseIssuesHandleFunc, "test_plans", catalog.RiskRead,
			[]string{"用例关联工作项列表", "关联缺陷列表", "plan case issues"},
			"List work items linked to a test case in a test plan.", map[string]any{"enterprise_id": 123, "program_id": 456, "test_plan_id": 789, "test_case_id": 101}),
		testSpec(test_plans.ExecuteTestPlanCasesTool, test_plans.ExecuteTestPlanCasesHandleFunc, "test_plans", catalog.RiskWrite,
			[]string{"记录执行结果", "批量执行用例", "执行用例", "execute test cases", "batch execute"},
			"Record execution results for test plan cases in one asynchronous batch.",
			map[string]any{"enterprise_id": 123, "program_id": 456, "test_plan_id": 789, "test_plan_case_ids": []any{2001}, "execute_state": 1}),
		testSpec(test_reports.CreateTestReportTool, test_reports.CreateTestReportHandleFunc, "test_reports", catalog.RiskWrite,
			[]string{"测试报告", "创建测试报告", "新建测试报告", "test report", "create test report"},
			"Create a test report from existing test plans.", map[string]any{"enterprise_id": 123, "program_id": 456, "title": "Release report", "test_plan_ids": []any{789}, "components": []any{map[string]any{"ident": "basic-info", "title": "Basic info"}, map[string]any{"ident": "case-statistics", "title": "Case statistics"}}}),
		testSpec(test_reports.ListTestReportsTool, test_reports.ListTestReportsHandleFunc, "test_reports", catalog.RiskRead,
			[]string{"测试报告列表", "报告列表", "test report list", "list test reports"},
			"List test reports in a project.", map[string]any{"enterprise_id": 123, "program_id": 456}),
		testSpec(test_reports.GetTestReportTool, test_reports.GetTestReportHandleFunc, "test_reports", catalog.RiskRead,
			[]string{"测试报告详情", "报告详情", "test report detail", "get test report"},
			"Get one test report with computed components.", map[string]any{"enterprise_id": 123, "program_id": 456, "test_report_id": 789}),
		testSpec(test_reports.UpdateTestReportTool, test_reports.UpdateTestReportHandleFunc, "test_reports", catalog.RiskWrite,
			[]string{"修改测试报告", "编辑测试报告", "update test report"},
			"Update a test report.", map[string]any{"enterprise_id": 123, "program_id": 456, "test_report_id": 789, "title": "Release report v2"}),
		testSpec(test_reports.DeleteTestReportTool, test_reports.DeleteTestReportHandleFunc, "test_reports", catalog.RiskDestructive,
			[]string{"删除测试报告", "移除测试报告", "delete test report"},
			"Delete one test report.", map[string]any{"enterprise_id": 123, "program_id": 456, "test_report_id": 789}),
		testSpec(test_cases.DeleteTestCaseTool, test_cases.DeleteTestCaseHandleFunc, "test_cases", catalog.RiskDestructive,
			[]string{"删除测试用例", "移除测试用例", "delete test case"},
			"Delete one test case permanently.", map[string]any{"enterprise_id": 123, "program_id": 456, "test_case_id": 789}),
		testSpec(test_plans.DeleteTestPlanTool, test_plans.DeleteTestPlanHandleFunc, "test_plans", catalog.RiskDestructive,
			[]string{"删除测试计划", "移除测试计划", "delete test plan"},
			"Delete one test plan.", map[string]any{"enterprise_id": 123, "program_id": 456, "test_plan_id": 789}),
		testSpec(test_modules.DeleteTestCaseModuleTool, test_modules.DeleteTestCaseModuleHandleFunc, "test_modules", catalog.RiskDestructive,
			[]string{"删除测试模块", "删除功能模块", "delete test module"},
			"Delete a test case module and all descendants.", map[string]any{"enterprise_id": 123, "program_id": 456, "case_module_id": 789}),
	}
}

func contextSpec(tool mcp.Tool, handler utils.OptionHandlerFunc, domain string, aliases []string, summary string, example map[string]any) catalog.Spec {
	return catalog.Spec{
		Tool:             tool,
		Handler:          handler,
		Domain:           domain,
		Tags:             []string{domain},
		Aliases:          aliases,
		Summary:          summary,
		ExampleArguments: example,
		Risk:             catalog.RiskRead,
		ReadOnly:         true,
		Invokable:        true,
	}
}

func p0ToolSpecs() []catalog.Spec {
	return []catalog.Spec{
		contextSpec(check_runs.ListPRCheckRunsTool, check_runs.ListPRCheckRunsHandleFunc, "check_runs",
			[]string{"PR检查", "PR门禁", "检查状态", "check runs", "pull request checks"},
			"List check runs for pull requests.", map[string]any{"enterprise_id": 123, "pull_request_ids": "456,789"}),
		contextSpec(check_runs.ListCommitCheckRunsTool, check_runs.ListCommitCheckRunsHandleFunc, "check_runs",
			[]string{"提交检查", "Commit检查", "commit check runs"},
			"List check runs for commits.", map[string]any{"enterprise_id": 123, "project_id": 456, "commit_ids": "abc123"}),
		contextSpec(check_runs.ListProjectCheckRunsTool, check_runs.ListProjectCheckRunsHandleFunc, "check_runs",
			[]string{"门禁列表", "仓库门禁", "gate checks", "project check runs"},
			"List gate/check runs for a repository.", map[string]any{"enterprise_id": 123, "project_id": 456, "pull_request_id": 789}),
		contextSpec(check_runs.GetCheckRunAnnotationsTool, check_runs.GetCheckRunAnnotationsHandleFunc, "check_runs",
			[]string{"门禁详情", "检查详情", "check run annotations", "失败原因"},
			"Get check run annotations and failure details.", map[string]any{"enterprise_id": 123, "project_id": 456, "check_run_id": 789}),

		contextSpec(programs.GetProgramTool, programs.GetProgramHandleFunc, "programs",
			[]string{"项目详情", "项目信息", "program detail", "project detail"},
			"Get one project by ID.", map[string]any{"enterprise_id": 123, "program_id": 456}),
		contextSpec(programs.GetProgramOverviewTool, programs.GetProgramOverviewHandleFunc, "programs",
			[]string{"项目概览", "项目进展", "program overview"},
			"Get a project overview with progress statistics.", map[string]any{"enterprise_id": 123, "program_id": 456}),
		contextSpec(programs.GetProgramComponentsTool, programs.GetProgramComponentsHandleFunc, "programs",
			[]string{"项目组件", "启用的模块", "program components"},
			"List project components.", map[string]any{"enterprise_id": 123, "program_id": 456}),

		contextSpec(groups.ListGroupMembersTool, groups.ListGroupMembersHandleFunc, "groups",
			[]string{"团队成员", "团队成员列表", "group members", "team members"},
			"List members of a team.", map[string]any{"enterprise_id": 123, "group_id": 456}),
		contextSpec(groups.ListGroupProjectsTool, groups.ListGroupProjectsHandleFunc, "groups",
			[]string{"团队仓库", "团队关联仓库", "group projects"},
			"List repositories associated with a team.", map[string]any{"enterprise_id": 123, "group_id": 456}),
		contextSpec(groups.ListGroupProgramsTool, groups.ListGroupProgramsHandleFunc, "groups",
			[]string{"团队项目", "团队关联项目", "group programs"},
			"List projects associated with a team.", map[string]any{"enterprise_id": 123, "group_id": 456}),
		contextSpec(groups.GetGroupStatisticsTool, groups.GetGroupStatisticsHandleFunc, "groups",
			[]string{"团队统计", "团队仓库统计", "group statistics"},
			"Get repository statistics for a team.", map[string]any{"enterprise_id": 123, "group_id": 456}),

		contextSpec(documents.SearchDocumentsTool, documents.SearchDocumentsHandleFunc, "documents",
			[]string{"文档搜索", "知识库搜索", "search documents", "wiki search"},
			"Search documents and knowledge-base nodes.", map[string]any{"enterprise_id": 123, "search": "onboarding"}),
		contextSpec(documents.ListRecentDocumentsTool, documents.ListRecentDocumentsHandleFunc, "documents",
			[]string{"最近文档", "最近编辑文档", "recent documents"},
			"List recently edited documents.", map[string]any{"enterprise_id": 123, "program_id": 456}),
		contextSpec(documents.ListSharedDocumentsTool, documents.ListSharedDocumentsHandleFunc, "documents",
			[]string{"共享文档", "分享给我的文档", "shared documents"},
			"List documents shared with the current user.", map[string]any{"enterprise_id": 123}),
		contextSpec(documents.GetDocumentNodeTool, documents.GetDocumentNodeHandleFunc, "documents",
			[]string{"文档节点", "文档详情", "document node", "doc node detail"},
			"Get one document node by ID.", map[string]any{"enterprise_id": 123, "doc_node_id": 456}),
		contextSpec(documents.ListDocumentVersionsTool, documents.ListDocumentVersionsHandleFunc, "documents",
			[]string{"文档版本", "历史版本", "document versions"},
			"List document attachment versions.", map[string]any{"enterprise_id": 123, "doc_node_id": 456}),

		contextSpec(issue_links.ListIssueAttachmentsTool, issue_links.ListIssueAttachmentsHandleFunc, "issue_links",
			[]string{"工作项附件", "任务附件", "issue attachments"},
			"List attachments of an issue.", map[string]any{"enterprise_id": 123, "issue_id": "I123"}),
		contextSpec(issue_links.ListIssueRelatedIssuesTool, issue_links.ListIssueRelatedIssuesHandleFunc, "issue_links",
			[]string{"关联任务", "关联工作项", "related issues", "linked issues"},
			"List work items related to an issue.", map[string]any{"enterprise_id": 123, "issue_id": "I123"}),
		contextSpec(issue_links.ListIssueAuthorizationsTool, issue_links.ListIssueAuthorizationsHandleFunc, "issue_links",
			[]string{"工作项权限", "任务权限", "issue permissions", "issue auths"},
			"Get current user's issue permissions.", map[string]any{"enterprise_id": 123, "issue_id": "I123"}),
	}
}

func p1ToolSpecs() []catalog.Spec {
	return []catalog.Spec{
		contextSpec(scrum_sprints.GetScrumSprintTool, scrum_sprints.GetScrumSprintHandleFunc, "scrum_sprints",
			[]string{"迭代详情", "sprint detail", "get sprint"},
			"Get one Scrum sprint by ID.", map[string]any{"enterprise_id": 123, "program_id": 456, "scrum_sprint_id": 789}),
		contextSpec(scrum_versions.GetScrumVersionTool, scrum_versions.GetScrumVersionHandleFunc, "scrum_versions",
			[]string{"版本详情", "version detail", "get version"},
			"Get one Scrum version by ID.", map[string]any{"enterprise_id": 123, "program_id": 456, "scrum_version_id": 789}),
		contextSpec(scrum_versions.ListScrumStagesTool, scrum_versions.ListScrumStagesHandleFunc, "scrum_versions",
			[]string{"阶段列表", "迭代阶段", "scrum stages"},
			"List stages of a Scrum version.", map[string]any{"enterprise_id": 123, "program_id": 456, "scrum_version_id": 789}),
		contextSpec(webhooks.ListWebhooksTool, webhooks.ListWebhooksHandleFunc, "webhooks",
			[]string{"Webhook列表", "hooks", "list webhooks"},
			"List webhooks.", map[string]any{"enterprise_id": 123, "scope": "all"}),
		contextSpec(webhooks.ListWebhookLogsTool, webhooks.ListWebhookLogsHandleFunc, "webhooks",
			[]string{"Webhook日志", "回调日志", "webhook logs"},
			"List request logs for one webhook.", map[string]any{"enterprise_id": 123, "webhook_id": 456}),
		contextSpec(baselines.ListBaselinesTool, baselines.ListBaselinesHandleFunc, "baselines",
			[]string{"基线列表", "baselines", "list baselines"},
			"List project baselines.", map[string]any{"enterprise_id": 123, "program_id": 456}),
		contextSpec(baselines.GetBaselineTool, baselines.GetBaselineHandleFunc, "baselines",
			[]string{"基线详情", "baseline detail", "get baseline"},
			"Get one project baseline.", map[string]any{"enterprise_id": 123, "program_id": 456, "baseline_id": 789}),
		contextSpec(common_views.ListCommonViewsTool, common_views.ListCommonViewsHandleFunc, "views",
			[]string{"公共视图", "视图列表", "common views", "saved views"},
			"List saved common views.", map[string]any{"enterprise_id": 123, "type": "program_view"}),
		contextSpec(kanbans.ListKanbansTool, kanbans.ListKanbansHandleFunc, "kanbans",
			[]string{"看板列表", "kanbans", "boards"},
			"List kanban boards.", map[string]any{"enterprise_id": 123}),
		contextSpec(kanbans.ListKanbanColumnsTool, kanbans.ListKanbanColumnsHandleFunc, "kanbans",
			[]string{"看板栏", "看板列", "kanban columns"},
			"List columns of a kanban board.", map[string]any{"enterprise_id": 123, "program_id": 456, "kanban_id": 789}),
	}
}

func p2ToolSpecs() []catalog.Spec {
	return []catalog.Spec{
		testSpec(documents.DeleteDocumentTool, documents.DeleteDocumentHandleFunc, "documents", catalog.RiskDestructive,
			[]string{"删除文档", "移除文档", "delete document"},
			"Delete an enterprise knowledge-base document or folder.",
			map[string]any{"enterprise_id": 123, "doc_node_id": 456}),
		testSpec(documents.DeleteDocumentPageTool, documents.DeleteDocumentPageHandleFunc, "documents", catalog.RiskDestructive,
			[]string{"删除文档页面", "删除文件夹", "delete document page"},
			"Delete one page or folder inside a document.",
			map[string]any{"enterprise_id": 123, "document_id": 456, "page_id": 789}),
		testSpec(scrum_sprints.DeleteScrumSprintTool, scrum_sprints.DeleteScrumSprintHandleFunc, "scrum_sprints", catalog.RiskDestructive,
			[]string{"删除迭代", "移除迭代", "delete sprint"},
			"Delete one Scrum sprint.",
			map[string]any{"enterprise_id": 123, "program_id": 456, "scrum_sprint_id": 789}),
		testSpec(issue_links.SetIssueStarredTool, issue_links.SetIssueStarredHandleFunc, "issue_links", catalog.RiskWrite,
			[]string{"星标工作项", "取消星标", "star issue", "unstar issue"},
			"Star or unstar an issue.", map[string]any{"enterprise_id": 123, "issue_id": "I123", "starred": true}),
		testSpec(issue_links.AddIssueReactionTool, issue_links.AddIssueReactionHandleFunc, "issue_links", catalog.RiskWrite,
			[]string{"工作项表态", "表情回应", "issue reaction"},
			"Add an emoji reaction to an issue.", map[string]any{"enterprise_id": 123, "issue_id": "I123", "text": "rocket"}),
		testSpec(issue_links.RemoveIssueReactionTool, issue_links.RemoveIssueReactionHandleFunc, "issue_links", catalog.RiskDestructive,
			[]string{"删除表态", "移除表情", "remove reaction"},
			"Remove one issue reaction.", map[string]any{"enterprise_id": 123, "reaction_id": 456}),
		testSpec(issue_links.AddIssueRelationTool, issue_links.AddIssueRelationHandleFunc, "issue_links", catalog.RiskWrite,
			[]string{"关联任务", "关联需求", "添加关联关系", "link issues", "add relation"},
			"Create a relation between two issues.", map[string]any{"enterprise_id": 123, "issue_id": "I123", "link_issue_id": 456, "ref_type": "normal"}),
		testSpec(issue_links.UpdateIssueRelationTool, issue_links.UpdateIssueRelationHandleFunc, "issue_links", catalog.RiskWrite,
			[]string{"修改关联关系", "update relation"},
			"Update an issue relation.", map[string]any{"enterprise_id": 123, "issue_id": "I123", "link_issue_id": 456, "ref_type": "finish_to_start"}),
		testSpec(issue_links.RemoveIssueRelationTool, issue_links.RemoveIssueRelationHandleFunc, "issue_links", catalog.RiskDestructive,
			[]string{"移除关联关系", "取消关联任务", "remove relation"},
			"Remove a relation between two issues.", map[string]any{"enterprise_id": 123, "issue_id": "I123", "link_issue_id": 456}),
		testSpec(issue_links.LinkIssueDocumentTool, issue_links.LinkIssueDocumentHandleFunc, "issue_links", catalog.RiskWrite,
			[]string{"工作项关联文档", "关联文档", "文档关联", "关联到工作项", "文档挂到任务", "关联知识库文档", "link document", "link doc node", "document_node_id"},
			"Link an existing knowledge-base document or attachment (document node) to an issue; doc_node_id is document_node_id from list_enterprise_documents, not document_id or page_id.",
			map[string]any{"enterprise_id": 123, "issue_id": "I123", "doc_node_id": 456}),
		testSpec(issue_links.UnlinkIssueDocumentTool, issue_links.UnlinkIssueDocumentHandleFunc, "issue_links", catalog.RiskWrite,
			[]string{"取消关联文档", "移除关联文档", "文档解除关联", "解除文档关联", "取消文档关联", "unlink document", "unlink doc node"},
			"Detach a document or attachment (document node) from an issue; read current links with get_issue_development_links(type='docs') first.",
			map[string]any{"enterprise_id": 123, "issue_id": "I123", "doc_node_id": 456}),
		testSpec(issue_links.LinkIssueTestCasesTool, issue_links.LinkIssueTestCasesHandleFunc, "issue_links", catalog.RiskWrite,
			[]string{"工作项关联测试用例", "关联用例", "link test cases"},
			"Link test plan cases to an issue.", map[string]any{"enterprise_id": 123, "issue_id": "I123", "test_plan_case_ids": []any{1, 2}}),
		testSpec(issue_links.UnlinkIssueTestCasesTool, issue_links.UnlinkIssueTestCasesHandleFunc, "issue_links", catalog.RiskWrite,
			[]string{"取消关联测试用例", "unlink test cases"},
			"Unlink test plan cases from an issue.", map[string]any{"enterprise_id": 123, "issue_id": "I123", "test_plan_case_ids": []any{1}}),
		testSpec(issue_recycle_bin.RestoreIssuesTool, issue_recycle_bin.RestoreIssuesHandleFunc, "issue_recycle_bin", catalog.RiskWrite,
			[]string{"恢复工作项", "回收站恢复", "restore issues"},
			"Restore deleted issues from the recycle bin.", map[string]any{"enterprise_id": 123, "issue_ids": []any{1, 2}}),
	}
}

func All() []catalog.Spec {
	specs := existingToolSpecs()
	specs = append(specs, pilotToolSpecs()...)
	specs = append(specs, pilotToolSpecsPhase2()...)
	specs = append(specs, pilotWriteSpecs()...)
	specs = append(specs, testToolSpecs()...)
	specs = append(specs, p0ToolSpecs()...)
	specs = append(specs, p1ToolSpecs()...)
	specs = append(specs, p2ToolSpecs()...)
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
