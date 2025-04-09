package user

import (
	"context"
	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	GetUserInfo = "get_user_info"
)

var GetUserInfoTool = mcp.NewTool(GetUserInfo, mcp.WithDescription("Get user info"))

func GetUserInfoHandleFunc(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	apiUrl := "/users"

	giteeClient := utils.NewGiteeClient("GET", apiUrl)

	data := types.UserInfo{}
	return giteeClient.HandleMCPResult(&data)
}
