package utils

import (
	"context"
	"regexp"

	"github.com/mark3labs/mcp-go/mcp"
)

// OptionHandlerFunc defines the signature for handlers that accept utils.Option.
type OptionHandlerFunc func(context.Context, mcp.CallToolRequest, ...Option) (*mcp.CallToolResult, error)

func IsAllDigits(s string) bool {
	pattern := regexp.MustCompile(`^[0-9]+$`)
	return pattern.MatchString(s)
}
