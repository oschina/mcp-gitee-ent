package enterprises

import (
	"context"
	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ListEnterprises = "list_enterprises"
)

var ListEnterprisesTool = mcp.NewTool(ListEnterprises,
	mcp.WithDescription("List user's enterprises"),
)

func ListEnterprisesHandleFunc(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	apiUrl := "/list"
	giteeClient := utils.NewGiteeClient("GET", apiUrl, utils.WithPayload(request.Params.Arguments))

	data := types.PagedResponse[types.BasicEnterprise]{}

	return giteeClient.HandleMCPResult(&data)
}
