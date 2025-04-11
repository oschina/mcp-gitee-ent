# Gitee Enterprise MCP Server

Gitee Enterprise MCP Server 是 Gitee Enterprise 的模型上下文协议（Model Context Protocol，MCP）服务器实现。它提供了一系列与 Gitee Enterprise API 交互的工具，使 AI 助手能够管理企业仓库、Issue、Pull Request 以及项目管理相关能力等。

## 功能特点

- 与 Gitee Enterprise 仓库、Issue、Pull Request 等资源进行交互
- 支持企业级操作和管理
- 可配置的 API BASE URL，支持不同的 Gitee Enterprise 实例

## 安装

### 前提条件

- Go 1.23.0 或更高版本
- 具有适当权限的 MCP 令牌，[前往获取](https://gitee.com/profile/mcp_gitee_ent_access_tokens)

### 从源代码构建

1. 克隆仓库：
   ```bash
   git clone https://gitee.com/oschina/mcp-gitee-ent.git
   cd mcp-gitee-ent
   ```

2. 构建项目：
   ```bash
   make build
   ```
   将 ./bin/mcp-gitee-ent 移动至系统环境变量
   
### 使用 go install 安装
   ```bash
   go install gitee.com/oschina/mcp-gitee-ent@latest
   ```

## 使用方法

检查 mcp-gitee-ent 版本：

```bash
mcp-gitee-ent --version
```

### MCP Hosts 配置

<div align="center">
  <a href="docs/install/claude.md" title="Claude"><img src="docs/install/logos/Claude.png" width=80 height=80></a>
  <a href="docs/install/cursor.md" title="Cursor"><img src="docs/install/logos/Cursor.png" width=80 height=80></a>
  <a href="docs/install/cline.md" title="Cline"><img src="docs/install/logos/Cline.png" width=80 height=80></a>
  <a href="docs/install/windsurf.md" title="Windsurf"><img src="docs/install/logos/Windsurf.png" width=80 height=80></a>
</div>

配置示例：
```json
{
  "mcpServers": {
    "gitee": {
      "command": "mcp-gitee-ent",
      "env": {
        "GITEE_ENT_API_BASE": "https://api.gitee.com/api/enterprises",
        "GITEE_ENT_MCP_ACCESS_TOKEN": "<your mcp ent access token>",
      }
    }
  }
}
```

### 命令行选项

- `-token`：访问令牌
- `-api-base`：Gitee ent API base URL（默认：https://api.gitee.com/api/enterprises）
- `-version`：显示版本信息
- `-transport`：传输类型（stdio 或 sse，默认：stdio）
- `-sse-address`：启动 SSE 服务器的主机和端口（默认：localhost:8000）

### 环境变量

您也可以使用环境变量配置服务器：

- `GITEE_ENT_MCP_ACCESS_TOKEN`：Gitee MCP ent 访问令牌
- `GITEE_ENT_API_BASE`：Gitee ent API base URL

## 许可证

本项目采用 MIT 许可证。有关更多详细信息，请参阅 [LICENSE](LICENSE) 文件。

## 可用工具

服务器提供了各种与 Gitee Enterprise 交互的工具：

| 工具                                    | 类别          | 描述               |
|----------------------------------------|---------------|------------------|
| **list_enterprises**     | 企业          | 列出用户的企业        |
| **list_enterprise_repositories** | 仓库    | 列出企业中的仓库       |
| **create_enterprise_repository** | 仓库    | 在企业中创建仓库       |
| **create_enterprise_repo_release** | 仓库  | 为仓库创建发行版       |
| **list_enterprise_repo_releases** | 仓库   | 列出仓库发行版        |
| **list_enterprise_pulls** | Pull Request | 列出企业拉取请求       |
| **create_enterprise_repo_pull** | Pull Request | 创建仓库拉取请求     |
| **merge_enterprise_pull** | Pull Request | 合并拉取请求         |
| **get_enterprise_pull_detail** | Pull Request | 获取拉取请求详情     |
| **update_enterprise_pull** | Pull Request | 更新拉取请求        |
| **get_enterprise_pull_diff** | Pull Request | 获取拉取请求差异     |
| **comment_enterprise_pull** | Pull Request | 评论拉取请求        |
| **list_enterprise_pull_comments** | Pull Request | 列出拉取请求评论   |
| **create_enterprise_issue** | Issue      | 创建 Issue         |
| **update_enterprise_issue** | Issue      | 更新 Issue         |
| **get_enterprise_issue_detail** | Issue   | 获取 Issue 详情     |
| **list_enterprise_issues** | Issue       | 列出 Issues        |
| **comment_enterprise_issue** | Issue      | 评论 Issue         |
| **list_enterprise_issue_comments** | Issue | 列出 Issue 评论    |
| **get_user_info**        | 用户          | 获取用户信息          |
| **list_enterprise_members** | 成员        | 列出企业成员          |
| **list_enterprise_groups** | 团队         | 列出企业团队          |
| **list_enterprise_labels** | 标签         | 列出企业标签          |
| **list_programs**        | 项目          | 列出企业项目          |
| **list_scrum_sprints**   | 项目        | 列出 Scrum 迭代     |
| **list_scrum_versions**  | 项目        | 列出 Scrum 版本     |
| **list_issue_types**     | 工作项类型      | 列出工作项类型        |
| **list_issue_type_states** | 工作项状态    | 列出工作项状态        |

## 贡献

我们欢迎开源社区的贡献！如果您想为这个项目做出贡献，请按照以下指南操作：

1. Fork 这个仓库。
2. 为您的功能或 bug 修复创建一个新分支。
3. 进行更改，并确保代码有良好的文档。
4. 提交一个 pull request，并附上清晰的更改描述。

更多信息，请参阅 [CONTRIBUTING](CONTRIBUTING.md) 文件。
