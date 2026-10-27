package catalog

import (
	"context"
	"strings"
	"testing"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

func testHandler(_ context.Context, _ mcp.CallToolRequest, _ ...utils.Option) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("ok"), nil
}

func testTool(name, description string, required ...string) mcp.Tool {
	options := []mcp.ToolOption{mcp.WithDescription(description)}
	for _, argument := range required {
		options = append(options, mcp.WithString(argument, mcp.Required()))
	}
	return mcp.NewTool(name, options...)
}

func TestRegistryRegisterAndListStableOrder(t *testing.T) {
	registry := NewRegistry()
	err := registry.Register(
		Spec{
			Tool:      testTool("zeta_tool", "Zeta description.", "enterprise_id"),
			Handler:   testHandler,
			Domain:    "issues",
			Risk:      RiskRead,
			ReadOnly:  true,
			Invokable: true,
		},
		Spec{
			Tool:      testTool("alpha_tool", "Alpha description."),
			Handler:   testHandler,
			Domain:    "pulls",
			Risk:      RiskWrite,
			Invokable: true,
		},
	)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	names := registry.Names()
	if len(names) != 2 || names[0] != "alpha_tool" || names[1] != "zeta_tool" {
		t.Fatalf("Names() = %v, want [alpha_tool zeta_tool]", names)
	}

	spec, ok := registry.Get("zeta_tool")
	if !ok {
		t.Fatal("Get(zeta_tool) not found")
	}
	if spec.Summary != "Zeta description" {
		t.Fatalf("Summary = %q, want derived first sentence", spec.Summary)
	}
	if spec.Verification != VerificationNone {
		t.Fatalf("Verification = %q, want none", spec.Verification)
	}
}

func TestRegistryRejectsDuplicate(t *testing.T) {
	registry := NewRegistry()
	spec := Spec{Tool: testTool("same_tool", "Same."), Handler: testHandler, Invokable: true}
	if err := registry.Register(spec); err != nil {
		t.Fatalf("first Register() error = %v", err)
	}
	if err := registry.Register(spec); err == nil {
		t.Fatal("second Register() error = nil, want duplicate error")
	}
}

func TestRegistryVerificationDisablesInvocation(t *testing.T) {
	registry := NewRegistry()
	err := registry.Register(Spec{
		Tool:         testTool("protected_tool", "Protected."),
		Handler:      testHandler,
		Verification: VerificationPassword,
		Invokable:    true,
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	spec, _ := registry.Get("protected_tool")
	if spec.Invokable {
		t.Fatal("Invokable = true, want false for password-protected tool")
	}
}

func TestRegistryDocumentsIncludeRequiredArgumentsAndAliases(t *testing.T) {
	registry := NewRegistry()
	err := registry.Register(Spec{
		Tool:      testTool("list_milestones", "List milestones.", "enterprise_id", "keyword"),
		Handler:   testHandler,
		Domain:    "milestones",
		Tags:      []string{"planning"},
		Aliases:   []string{"里程碑"},
		Risk:      RiskRead,
		ReadOnly:  true,
		Invokable: true,
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	documents := registry.Documents()
	if len(documents) != 1 {
		t.Fatalf("Documents() length = %d, want 1", len(documents))
	}
	document := documents[0]
	if document.Name != "list_milestones" || document.Domain != "milestones" {
		t.Fatalf("Document = %#v", document)
	}
	for _, expected := range []string{"list_milestones", "List milestones", "里程碑", "enterprise_id", "keyword"} {
		if !strings.Contains(document.SearchText, expected) {
			t.Errorf("SearchText %q does not contain %q", document.SearchText, expected)
		}
	}
}
