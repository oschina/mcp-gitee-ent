package main

import (
	"context"
	"flag"
	"fmt"
	"gitee.com/oschina/mcp-gitee-ent/operations/enterprises"
	"gitee.com/oschina/mcp-gitee-ent/operations/groups"
	"gitee.com/oschina/mcp-gitee-ent/operations/issue_states"
	"gitee.com/oschina/mcp-gitee-ent/operations/issue_types"
	"gitee.com/oschina/mcp-gitee-ent/operations/issues"
	"gitee.com/oschina/mcp-gitee-ent/operations/labels"
	"gitee.com/oschina/mcp-gitee-ent/operations/members"
	"gitee.com/oschina/mcp-gitee-ent/operations/programs"
	"gitee.com/oschina/mcp-gitee-ent/operations/pulls"
	"gitee.com/oschina/mcp-gitee-ent/operations/repository"
	"gitee.com/oschina/mcp-gitee-ent/operations/scrum_sprints"
	"gitee.com/oschina/mcp-gitee-ent/operations/scrum_versions"
	"gitee.com/oschina/mcp-gitee-ent/operations/user"
	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"log"
	"os"
	"strings"
)

var (
	Version              = utils.Version
	disabledToolsetsFlag string
	enabledToolsetsFlag  string
)


// wrapOptionHandler creates a standard ToolHandlerFunc from an OptionHandlerFunc,
// allowing predefined options to be passed during registration.
func wrapOptionHandler(handler utils.OptionHandlerFunc, opts ...utils.Option) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Call the original handler, passing the captured options.
		return handler(ctx, request, opts...)
	}
}

func newMCPServer() *server.MCPServer {
	return server.NewMCPServer(
		"mcp-gitee-ent",
		Version,
		server.WithToolCapabilities(false),
		server.WithLogging(),
	)
}

func addTool(s *server.MCPServer, tool mcp.Tool, handleFunc utils.OptionHandlerFunc) {
	enabledToolsets := getEnabledToolsets()
	if len(enabledToolsets) == 0 {
		s.AddTool(tool, wrapOptionHandler(handleFunc))
		return
	}

	for i := range enabledToolsets {
		enabledToolsets[i] = strings.TrimSpace(enabledToolsets[i])
	}

	for _, keepTool := range enabledToolsets {
		if tool.Name == keepTool {
			s.AddTool(tool, wrapOptionHandler(handleFunc))
			return
		}
	}
}

func disableTools(s *server.MCPServer) {
	if enabledToolsetsFlag != "" {
		enabledToolsetsFlag = os.Getenv("ENABLED_TOOLSETS")
	}

	if enabledToolsetsFlag != "" {
		return
	}

	if disabledTools := getDisabledToolsets(); len(disabledTools) > 0 {
		s.DeleteTools(disabledTools...)
	}
}

func addTools(s *server.MCPServer) {
	// Issues
	addTool(s, issues.ListIssuesTool, issues.ListIssuesHandleFunc)
	addTool(s, issues.CreateIssueTool, issues.CreateIssueHandleFunc)
	addTool(s, issues.GetIssueDetailTool, issues.GetIssueDetailHandleFunc)
	addTool(s, issues.UpdateIssueTool, issues.UpdateIssueHandleFunc)
	addTool(s, issues.CommentIssueTool, issues.CommentIssueHandleFunc)
	addTool(s, issues.ListIssueCommentsTool, issues.ListIssueCommentsHandleFunc)

	// Repositories
	addTool(s, repository.ListRepositoriesTool, repository.ListRepositoriesHandleFunc)
	addTool(s, repository.CreateRepositoryTool, repository.CreateRepositoryHandleFunc)
	addTool(s, repository.CreateReleaseTool, repository.CreateReleaseHandleFunc)
	addTool(s, repository.ListReleasesTool, repository.ListReleasesHandleFunc)

	// Pulls
	addTool(s, pulls.ListEntPullsTool, pulls.ListEntPullsHandleFunc)
	addTool(s, pulls.CreateEntPullTool, pulls.CreateEntPullHandleFunc)
	addTool(s, pulls.GetEntPullDetailTool, pulls.GetPullDetailHandleFunc)
	addTool(s, pulls.GetPullDiffTool, pulls.GetPullDiffHandleFunc)
	addTool(s, pulls.CommentPullTool, pulls.CommentPullHandleFunc)
	addTool(s, pulls.ListPullCommentsTool, pulls.ListPullCommentsHandleFunc)
	addTool(s, pulls.MergePullTool, pulls.MergePullHandleFunc)
	addTool(s, pulls.UpdatePullTool, pulls.UpdatePullHandleFunc)

	// Enterprises
	addTool(s, enterprises.ListEnterprisesTool, enterprises.ListEnterprisesHandleFunc)

	// Labels
	addTool(s, labels.ListEnterpriseLabelsTool, labels.ListEnterpriseLabelsHandleFunc)

	// Issue Types
	addTool(s, issue_types.ListIssueTypesTool, issue_types.ListIssueTypesHandleFunc)

	// Issue States
	addTool(s, issue_states.ListIssueTypeStatesTool, issue_states.ListIssueTypeStatesHandleFunc)

	// Users
	addTool(s, user.GetUserInfoTool, user.GetUserInfoHandleFunc)

	// Programs
	addTool(s, programs.ListProgramsTool, programs.ListProgramsHandleFunc)

	// Scrum Sprints
	addTool(s, scrum_sprints.CreateScrumSprintTool, scrum_sprints.CreateScrumSprintHandleFunc)
	addTool(s, scrum_sprints.ListScrumSprintsTool, scrum_sprints.ListScrumSprintsHandleFunc)

	// Scrum Versions
	addTool(s, scrum_versions.ListScrumVersionsTool, scrum_versions.ListScrumVersionsHandleFunc)

	// Members
	addTool(s, members.ListEntMembersTool, members.ListEntMembersHandleFunc)

	// Groups
	addTool(s, groups.ListEntGroupsTool, groups.ListEntGroupsHandleFunc)
}

func getDisabledToolsets() []string {
	if disabledToolsetsFlag == "" {
		disabledToolsetsFlag = os.Getenv("DISABLED_TOOLSETS")
	}

	if disabledToolsetsFlag == "" {
		return nil
	}

	tools := strings.Split(disabledToolsetsFlag, ",")
	for i := range tools {
		tools[i] = strings.TrimSpace(tools[i])
	}

	return tools
}

func getEnabledToolsets() []string {
	if enabledToolsetsFlag == "" {
		enabledToolsetsFlag = os.Getenv("ENABLED_TOOLSETS")
	}
	if enabledToolsetsFlag == "" {
		return nil
	}
	tools := strings.Split(enabledToolsetsFlag, ",")
	for i := range tools {
		tools[i] = strings.TrimSpace(tools[i])
	}
	return tools
}

func run(transport, addr string) error {
	s := newMCPServer()

	addTools(s)
	disableTools(s)

	switch transport {
	case "stdio":
		if err := server.ServeStdio(s); err != nil {
			if err == context.Canceled {
				return nil
			}
			return err
		}
	case "sse":
		srv := server.NewSSEServer(s, server.WithBaseURL(addr))
		log.Printf("SSE server listening on %s", addr)
		if err := srv.Start(addr); err != nil {
			if err == context.Canceled {
				return nil
			}
			return fmt.Errorf("server error: %v", err)
		}
	default:
		return fmt.Errorf(
			"invalid transport type: %s. Must be 'stdio' or 'sse'",
			transport,
		)
	}
	return nil
}

func main() {
	var (
		accessToken string
		apiBase     string
		showVersion bool
		transport   string
		addr        string
	)

	flag.StringVar(&accessToken, "token", "", "Gitee Ent MCP access token")
	flag.StringVar(&apiBase, "api-base", "", "Gitee Ent API base URL (default: https://api.gitee.com/enterprises)")
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.StringVar(&transport, "transport", "stdio", "Transport type (stdio or sse)")
	flag.StringVar(&addr, "sse-address", "localhost:8000", "The host and port to start the sse server on")
	flag.StringVar(&disabledToolsetsFlag, "disabled-toolsets", "", "Comma-separated list of tools to disable")
	flag.StringVar(&enabledToolsetsFlag, "enabled-toolsets", "", "Comma-separated list of tools to enable (if specified, only these tools will be available)")
	flag.Parse()

	if showVersion {
		fmt.Printf("Gitee MCP Ent Server\n")
		fmt.Printf("Version: %s\n", Version)
		os.Exit(0)
	}

	if accessToken != "" {
		utils.SetGiteeAccessToken(accessToken)
	}

	if apiBase != "" {
		utils.SetApiBase(apiBase)
	}

	if err := run(transport, addr); err != nil {
		panic(err)
	}
}
