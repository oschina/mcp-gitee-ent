package toolspecs

import (
	"gitee.com/oschina/mcp-gitee-ent/catalog"
	"gitee.com/oschina/mcp-gitee-ent/operations/baselines"
	"gitee.com/oschina/mcp-gitee-ent/operations/check_runs"
	"gitee.com/oschina/mcp-gitee-ent/operations/pulls"
	"gitee.com/oschina/mcp-gitee-ent/operations/repository"
	"gitee.com/oschina/mcp-gitee-ent/operations/webhooks"
)

func repoSpecs() []catalog.Spec {
	return []catalog.Spec{
		visibleSpec(repository.ListRepositoriesTool, repository.ListRepositoriesHandleFunc, "repository", catalog.RiskRead),
		visibleSpec(repository.CreateRepositoryTool, repository.CreateRepositoryHandleFunc, "repository", catalog.RiskWrite),
		visibleSpec(repository.CreateReleaseTool, repository.CreateReleaseHandleFunc, "repository", catalog.RiskWrite),
		visibleSpec(repository.ListReleasesTool, repository.ListReleasesHandleFunc, "repository", catalog.RiskRead),
		visibleSpec(repository.GetRepoTreeTool, repository.GetRepoTreeHandleFunc, "repository", catalog.RiskRead),
		visibleSpec(repository.GetFileContentTool, repository.GetFileContentHandleFunc, "repository", catalog.RiskRead),
		visibleSpec(pulls.ListEntPullsTool, pulls.ListEntPullsHandleFunc, "pulls", catalog.RiskRead),
		visibleSpec(pulls.CreateEntPullTool, pulls.CreateEntPullHandleFunc, "pulls", catalog.RiskWrite),
		visibleSpec(pulls.GetEntPullDetailTool, pulls.GetPullDetailHandleFunc, "pulls", catalog.RiskRead),
		visibleSpec(pulls.GetPullDiffTool, pulls.GetPullDiffHandleFunc, "pulls", catalog.RiskRead),
		visibleSpec(pulls.CommentPullTool, pulls.CommentPullHandleFunc, "pulls", catalog.RiskWrite),
		visibleSpec(pulls.ListPullCommentsTool, pulls.ListPullCommentsHandleFunc, "pulls", catalog.RiskRead),
		visibleSpec(pulls.MergePullTool, pulls.MergePullHandleFunc, "pulls", catalog.RiskWrite),
		visibleSpec(pulls.UpdatePullTool, pulls.UpdateEntPullHandleFunc, "pulls", catalog.RiskWrite),
		hiddenSpec(check_runs.ListPRCheckRunsTool, check_runs.ListPRCheckRunsHandleFunc, "check_runs", catalog.RiskRead,
			[]string{"PR检查", "PR门禁", "检查状态", "check runs", "pull request checks"},
			"List check runs for pull requests.", map[string]any{"enterprise_id": 123, "pull_request_ids": "456,789"}),
		hiddenSpec(check_runs.ListCommitCheckRunsTool, check_runs.ListCommitCheckRunsHandleFunc, "check_runs", catalog.RiskRead,
			[]string{"提交检查", "Commit检查", "commit check runs"},
			"List check runs for commits.", map[string]any{"enterprise_id": 123, "project_id": 456, "commit_ids": "abc123"}),
		hiddenSpec(check_runs.ListProjectCheckRunsTool, check_runs.ListProjectCheckRunsHandleFunc, "check_runs", catalog.RiskRead,
			[]string{"门禁列表", "仓库门禁", "gate checks", "project check runs"},
			"List gate/check runs for a repository.", map[string]any{"enterprise_id": 123, "project_id": 456, "pull_request_id": 789}),
		hiddenSpec(check_runs.GetCheckRunAnnotationsTool, check_runs.GetCheckRunAnnotationsHandleFunc, "check_runs", catalog.RiskRead,
			[]string{"门禁详情", "检查详情", "check run annotations", "失败原因"},
			"Get check run annotations and failure details.", map[string]any{"enterprise_id": 123, "project_id": 456, "check_run_id": 789}),
		hiddenSpec(webhooks.ListWebhooksTool, webhooks.ListWebhooksHandleFunc, "webhooks", catalog.RiskRead,
			[]string{"Webhook列表", "hooks", "list webhooks"},
			"List webhooks.", map[string]any{"enterprise_id": 123, "scope": "all"}),
		hiddenSpec(webhooks.ListWebhookLogsTool, webhooks.ListWebhookLogsHandleFunc, "webhooks", catalog.RiskRead,
			[]string{"Webhook日志", "回调日志", "webhook logs"},
			"List request logs for one webhook.", map[string]any{"enterprise_id": 123, "webhook_id": 456}),
		hiddenSpec(baselines.ListBaselinesTool, baselines.ListBaselinesHandleFunc, "baselines", catalog.RiskRead,
			[]string{"基线列表", "baselines", "list baselines"},
			"List project baselines.", map[string]any{"enterprise_id": 123, "program_id": 456}),
		hiddenSpec(baselines.GetBaselineTool, baselines.GetBaselineHandleFunc, "baselines", catalog.RiskRead,
			[]string{"基线详情", "baseline detail", "get baseline"},
			"Get one project baseline.", map[string]any{"enterprise_id": 123, "program_id": 456, "baseline_id": 789}),
		hiddenSpec(repository.CreateBranchTool, repository.CreateBranchHandleFunc, "repository", catalog.RiskWrite,
			[]string{"创建分支", "新建分支", "仓库分支", "branch", "branches"},
			"Create a repository branch.",
			map[string]any{"enterprise_id": 123, "repository_id": 456, "branch_name": "feat/x", "refs": "master"}),
		hiddenSpec(repository.CreateRepositoryFileTool, repository.CreateRepositoryFileHandleFunc, "repository", catalog.RiskWrite,
			[]string{"创建文件", "新建文件", "新建仓库文件", "仓库文件", "file"},
			"Create a file in a repository branch.",
			map[string]any{"enterprise_id": 123, "repository_id": 456, "ref": "master", "file_path": "docs/a.md", "content": "# hi", "commit_title": "docs: add a.md"}),
		hiddenSpec(repository.UpdateRepositoryFileTool, repository.UpdateRepositoryFileHandleFunc, "repository", catalog.RiskWrite,
			[]string{"更新文件", "修改文件", "更新仓库文件", "仓库文件", "update file"},
			"Update an existing repository file.",
			map[string]any{"enterprise_id": 123, "repository_id": 456, "ref": "master", "previous_path": "docs/a.md", "file_name": "docs/a.md", "content": "# hi2", "commit_title": "docs: update a.md"}),
		hiddenSpec(repository.DeleteRepositoryFileTool, repository.DeleteRepositoryFileHandleFunc, "repository", catalog.RiskDestructive,
			[]string{"删除文件", "移除文件", "删除仓库文件", "仓库文件", "delete file"},
			"Delete a file from a repository branch.",
			map[string]any{"enterprise_id": 123, "repository_id": 456, "ref": "master/docs/a.md"}),
	}
}
