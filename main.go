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
	"github.com/mark3labs/mcp-go/server"
	"log"
	"os"
)

var (
	Version = utils.Version
)

func newMCPServer() *server.MCPServer {
	return server.NewMCPServer(
		"mcp-gitee-ent",
		Version,
		server.WithToolCapabilities(false),
		server.WithLogging(),
	)
}

func addTools(s *server.MCPServer) {
	//Issues
	s.AddTool(issues.ListIssuesTool, issues.ListIssuesHandleFunc)
	s.AddTool(issues.CreateIssueTool, issues.CreateIssueHandleFunc)
	s.AddTool(issues.GetIssueDetailTool, issues.GetIssueDetailHandleFunc)
	s.AddTool(issues.UpdateIssueTool, issues.UpdateIssueHandleFunc)
	s.AddTool(issues.CommentIssueTool, issues.CommentIssueHandleFunc)
	s.AddTool(issues.ListIssueCommentsTool, issues.ListIssueCommentsHandleFunc)

	// Repositories
	s.AddTool(repository.ListRepositoriesTool, repository.ListRepositoriesHandleFunc)
	s.AddTool(repository.CreateRepositoryTool, repository.CreateRepositoryHandleFunc)
	s.AddTool(repository.CreateReleaseTool, repository.CreateReleaseHandleFunc)
	s.AddTool(repository.ListReleasesTool, repository.ListReleasesHandleFunc)

	// Pulls
	s.AddTool(pulls.ListEntPullsTool, pulls.ListEntPullsHandleFunc)
	s.AddTool(pulls.CreateEntPullTool, pulls.CreateEntPullHandleFunc)
	s.AddTool(pulls.GetEntPullDetailTool, pulls.GetPullDetailHandleFunc)
	s.AddTool(pulls.GetPullDiffTool, pulls.GetPullDiffHandleFunc)
	s.AddTool(pulls.CommentPullTool, pulls.CommentPullHandleFunc)
	s.AddTool(pulls.ListPullCommentsTool, pulls.ListPullCommentsHandleFunc)
	s.AddTool(pulls.MergePullTool, pulls.MergePullHandleFunc)
	s.AddTool(pulls.UpdatePullTool, pulls.UpdatePullHandleFunc)

	// Enterprises
	s.AddTool(enterprises.ListEnterprisesTool, enterprises.ListEnterprisesHandleFunc)

	// Labels
	s.AddTool(labels.ListEnterpriseLabelsTool, labels.ListEnterpriseLabelsHandleFunc)

	// IssueTypes
	s.AddTool(issue_types.ListIssueTypesTool, issue_types.ListIssueTypesHandleFunc)

	// IssueStates
	s.AddTool(issue_states.ListIssueTypeStatesTool, issue_states.ListIssueTypeStatesHandleFunc)

	//Users
	s.AddTool(user.GetUserInfoTool, user.GetUserInfoHandleFunc)

	// Programs
	s.AddTool(programs.ListProgramsTool, programs.ListProgramsHandleFunc)

	// ScrumSprints
	s.AddTool(scrum_sprints.CreateScrumSprintTool, scrum_sprints.CreateScrumSprintHandleFunc)
	s.AddTool(scrum_sprints.ListScrumSprintsTool, scrum_sprints.ListScrumSprintsHandleFunc)

	// ScrumVersions
	s.AddTool(scrum_versions.ListScrumVersionsTool, scrum_versions.ListScrumVersionsHandleFunc)

	// Members
	s.AddTool(members.ListEntMembersTool, members.ListEntMembersHandleFunc)

	// Groups
	s.AddTool(groups.ListEntGroupsTool, groups.ListEntGroupsHandleFunc)

}

func run(transport, addr string) error {
	s := newMCPServer()
	addTools(s)
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
	accessToken := flag.String("token", "", "Gitee Ent MCP access token")
	apiBase := flag.String("api-base", "", "Gitee Ent API base URL (default: https://api.gitee.com/enterprises)")
	showVersion := flag.Bool("version", false, "Show version information")
	transport := flag.String("transport", "stdio", "Transport type (stdio or sse)")
	addr := flag.String("sse-address", "localhost:8000", "The host and port to start the sse server on")
	flag.Parse()

	if *showVersion {
		fmt.Printf("Gitee MCP Ent Server\n")
		fmt.Printf("Version: %s\n", Version)
		os.Exit(0)
	}

	if *accessToken != "" {
		utils.SetGiteeAccessToken(*accessToken)
	}

	if *apiBase != "" {
		utils.SetApiBase(*apiBase)
	}

	if err := run(*transport, *addr); err != nil {
		panic(err)
	}
}
