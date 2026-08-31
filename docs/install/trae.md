npx
```json
{
  "mcpServers": {
    "gitee-ent": {
      "command": "npx",
      "args": [
        "-y",
        "@gitee/mcp-gitee-ent@latest"
      ],
      "env": {
         "GITEE_ENT_API_BASE": "https://api.gitee.com/enterprises",
         "GITEE_ENT_MCP_ACCESS_TOKEN": "<your mcp ent access token>",
         "GITEE_ENT_MCP_ALLOWED_FILE_ROOTS": "/absolute/path/to/allowed/files"
      }
    }
  }
}
```

stdio mode:
```json
{
  "mcpServers": {
    "gitee": {
      "command": "mcp-gitee-ent",
      "env": {
        "GITEE_ENT_API_BASE": "https://api.gitee.com/enterprises",
        "GITEE_ENT_MCP_ACCESS_TOKEN": "<your mcp ent access token>",
        "GITEE_ENT_MCP_ALLOWED_FILE_ROOTS": "/absolute/path/to/allowed/files"
      }
    }
  }
}
```

`GITEE_ENT_MCP_ALLOWED_FILE_ROOTS` controls which local directories document upload and download tools may access. Use the operating system path-list separator to configure multiple absolute directories. Set it to `/` on Unix/macOS to allow every path permitted by the operating system; on Windows, list each drive root, for example `C:\;D:\`. Avoid filesystem-wide access on shared or remote servers.
