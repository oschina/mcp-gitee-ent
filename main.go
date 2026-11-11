package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"gitee.com/oschina/mcp-gitee-ent/catalog"
	"gitee.com/oschina/mcp-gitee-ent/operations/discovery"
	"gitee.com/oschina/mcp-gitee-ent/toolspecs"
	"gitee.com/oschina/mcp-gitee-ent/utils"
	"gitee.com/oschina/mcp-gitee-ent/utils/urlbuilder"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
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
		if timeout := utils.ToolTimeout(); timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
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

func addTools(s *server.MCPServer) {
	registry, err := buildToolCatalog()
	if err != nil {
		panic(fmt.Errorf("build tool catalog: %w", err))
	}

	for _, spec := range registry.List() {
		if !spec.Visible {
			continue
		}
		s.AddTool(spec.Tool, wrapOptionHandler(spec.Handler, spec.Options...))
	}

	addDiscoveryTools(s, registry)
}

// buildToolCatalog builds the full catalog and applies the existing
// ENABLED_TOOLSETS / DISABLED_TOOLSETS semantics to visibility and invocation.
func buildToolCatalog() (*catalog.Registry, error) {
	return toolspecs.Build(toolAllowed)
}

// toolAllowed mirrors the historical registration behavior: an enabled list is
// an exclusive allowlist; otherwise a disabled list is a denylist.
func toolAllowed(name string) bool {
	if enabled := getEnabledToolsets(); len(enabled) > 0 {
		for _, candidate := range enabled {
			if strings.TrimSpace(candidate) == name {
				return true
			}
		}
		return false
	}

	for _, candidate := range getDisabledToolsets() {
		if strings.TrimSpace(candidate) == name {
			return false
		}
	}
	return true
}

// infrastructureToolAllowed keeps discovery tools available even when an
// enabled-tools allowlist is configured, unless they are explicitly disabled.
// An explicit enabled entry wins over a matching disabled entry.
func infrastructureToolAllowed(name string) bool {
	for _, candidate := range getEnabledToolsets() {
		if strings.TrimSpace(candidate) == name {
			return true
		}
	}
	for _, candidate := range getDisabledToolsets() {
		if strings.TrimSpace(candidate) == name {
			return false
		}
	}
	return true
}

func addDiscoveryTools(s *server.MCPServer, registry *catalog.Registry) {
	searchHandler := discovery.NewSearchToolsHandler(registry, nil)
	inspectHandler := discovery.NewInspectToolHandler(registry)
	invokeHandler := discovery.NewInvokeToolHandler(registry)

	if infrastructureToolAllowed(discovery.SearchTools) {
		s.AddTool(discovery.SearchToolsTool, wrapOptionHandler(searchHandler.Handle))
	}
	if infrastructureToolAllowed(discovery.InspectTool) {
		s.AddTool(discovery.InspectToolTool, wrapOptionHandler(inspectHandler.Handle))
	}
	if infrastructureToolAllowed(discovery.InvokeTool) {
		s.AddTool(discovery.InvokeToolTool, wrapOptionHandler(invokeHandler.Handle))
	}
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
		webBase     string
		showVersion bool
		transport   string
		addr        string
	)

	flag.StringVar(&accessToken, "token", "", "Gitee Ent MCP access token")
	flag.StringVar(&apiBase, "api-base", "", "Gitee Ent API base URL (default: https://api.gitee.com/enterprises)")
	flag.StringVar(&webBase, "web-base", "", "Gitee Ent web origin used for html_url (default: https://e.gitee.com)")
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

	if webBase != "" {
		urlbuilder.SetWebBase(webBase)
	}

	utils.EnableEnterprisePathAutoResolve(true)
	go utils.WarmEnterprisePaths()

	if err := run(transport, addr); err != nil {
		panic(err)
	}
}
