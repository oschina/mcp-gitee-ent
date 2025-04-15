stdio mode
```json
{
  "mcpServers": {
    "gitee": {
      "command": "mcp-gitee-ent",
      "env": {
        "GITEE_ENT_API_BASE": "https://api.gitee.com/enterprises",
        "GITEE_ENT_MCP_ACCESS_TOKEN": "<your mcp ent access token>"
      }
    }
  }
}
```

sse mode

start mcp server through sse
```bash
mcp-gitee-ent -token <your mcp ent access token> -transport sse
```
```json
{
  "mcpServers": {
    "gitee": {
      "url": "http://localhost:8000/sse",
    }
  }
}
```