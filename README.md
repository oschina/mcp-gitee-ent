# Gitee Enterprise MCP Server

Gitee Enterprise MCP Server is a Model Context Protocol (MCP) server implementation for Gitee Enterprise. It provides a set of tools for interacting with Gitee Enterprise API, allowing AI assistants to manage enterprise repositories, issues, pull requests, and more.

## Features

- Interact with Gitee Enterprise repositories, issues, pull requests
- Support for enterprise-level operations and management
- Configurable API base URL to support different Gitee Enterprise instances

## Installation

### Prerequisites

- Go 1.23.0 or higher
- MCP token, [Go to get](https://gitee.com/profile/mcp_gitee_ent_access_tokens)

### Building from Source

1. Clone the repository:
   ```bash
   git clone https://gitee.com/oschina/mcp-gitee-ent.git
   cd mcp-gitee-ent
   ```

2. Build the project:
   ```bash
   make build
   ```
   Move ./bin/mcp-gitee-ent PATH env

### Use go install
   ```bash
   go install gitee.com/oschina/mcp-gitee-ent@latest
   ```

## Usage

Check mcp-gitee-ent version:

```bash
mcp-gitee-ent --version
```

### MCP Hosts Configuration
<div align="center">
  <a href="docs/install/claude.md" title="Claude"><img src="docs/install/logos/Claude.png" width=80 height=80></a>
  <a href="docs/install/cursor.md" title="Cursor"><img src="docs/install/logos/Cursor.png" width=80 height=80></a>
  <a href="docs/install/cline.md" title="Cline"><img src="docs/install/logos/Cline.png" width=80 height=80></a>
  <a href="docs/install/windsurf.md" title="Windsurf"><img src="docs/install/logos/Windsurf.png" width=80 height=80></a>
</div>

Config example:
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

### Command-line Options

- `-token`: Gitee access token
- `-api-base`: Gitee API base URL (default: https://api.gitee.com/api/enterprises)
- `-version`: Show version information
- `-transport`: Transport type (stdio or sse, default: stdio)
- `-sse-address`: The host and port to start the SSE server on (default: localhost:8000)

### Environment Variables

You can also configure the server using environment variables:

- `GITEE_ENT_MCP_ACCESS_TOKEN`: Gitee mcp access token
- `GITEE_ENT_API_BASE`: Gitee API base URL

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.

## Available Tools

The server provides various tools for interacting with Gitee Enterprise:

| Tool                                    | Category      | Description |
|----------------------------------------|---------------|-------------|
| **list_enterprises**     | Enterprise    | List user's enterprises |
| **list_enterprise_repositories** | Repository | List repositories in an enterprise |
| **create_enterprise_repository** | Repository | Create a repository in an enterprise |
| **create_enterprise_repo_release** | Repository | Create a release for repository |
| **list_enterprise_repo_releases** | Repository | List releases for repository |
| **list_enterprise_pulls** | Pull Request | List enterprise pull requests |
| **create_enterprise_repo_pull** | Pull Request | Create a pull request for repository |
| **merge_enterprise_pull** | Pull Request | Merge a pull request |
| **get_enterprise_pull_detail** | Pull Request | Get pull request detail |
| **update_enterprise_pull** | Pull Request | Update a pull request |
| **get_enterprise_pull_diff** | Pull Request | Get pull request diff |
| **comment_enterprise_pull** | Pull Request | Comment on a pull request |
| **list_enterprise_pull_comments** | Pull Request | List pull request comments |
| **create_enterprise_issue** | Issue | Create an issue |
| **update_enterprise_issue** | Issue | Update an issue |
| **get_enterprise_issue_detail** | Issue | Get issue detail |
| **list_enterprise_issues** | Issue | List issues |
| **comment_enterprise_issue** | Issue | Comment on an issue |
| **list_enterprise_issue_comments** | Issue | List issue comments |
| **get_user_info**        | User         | Get user info |
| **list_enterprise_members** | Member | List members of an enterprise |
| **list_enterprise_groups** | Group | List groups in an enterprise |
| **list_enterprise_labels** | Label | List labels of an enterprise |
| **list_programs**        | Program      | List programs of an enterprise |
| **list_scrum_sprints**   | Program        | List Scrum Sprints |
| **list_scrum_versions**  | Program        | List Scrum Versions |
| **list_issue_types**     | Issue Type   | List issue types |
| **list_issue_type_states** | Issue State | List issue states |


## Contribution

We welcome contributions from the open-source community! If you'd like to contribute to this project, please follow these guidelines:

1. Fork the repository.
2. Create a new branch for your feature or bug fix.
3. Make your changes and ensure the code is well-documented.
4. Submit a pull request with a clear description of your changes.

For more information, please refer to the [CONTRIBUTING](CONTRIBUTING.md) file.
