// Package catalog defines the discoverable tool catalog used by the
// progressive-discovery meta tools (search_tools / inspect_tool / invoke_tool).
package catalog

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

// Risk classifies a tool by the blast radius of its side effects.
type Risk string

const (
	RiskRead        Risk = "read"
	RiskWrite       Risk = "write"
	RiskDestructive Risk = "destructive"
	RiskAdmin       Risk = "admin"
)

// Verification describes the secondary verification required by an upstream API.
type Verification string

const (
	VerificationNone     Verification = "none"
	VerificationPassword Verification = "password"
	VerificationCaptcha  Verification = "captcha"
)

// Spec describes one tool in the catalog.
type Spec struct {
	Tool    mcp.Tool
	Handler utils.OptionHandlerFunc
	Options []utils.Option

	Domain  string
	Tags    []string
	Aliases []string

	// Summary is a one-line description used by search results. When empty it
	// is derived from the tool description.
	Summary string
	// Notes are extra routing hints shown by inspect_tool.
	Notes []string
	// ExampleArguments is shown by inspect_tool and must satisfy the tool schema.
	ExampleArguments map[string]any

	Risk         Risk
	Verification Verification
	ReadOnly     bool
	Idempotent   bool

	// Visible controls whether the tool is directly registered on tools/list.
	Visible bool
	// Invokable controls whether invoke_tool may call the tool. It must be
	// false for tools excluded by toolsets or requiring secondary verification.
	Invokable bool
}

// Document is the searchable form of a Spec.
type Document struct {
	Name       string   `json:"name"`
	Summary    string   `json:"summary"`
	Domain     string   `json:"domain"`
	Tags       []string `json:"tags,omitempty"`
	Aliases    []string `json:"aliases,omitempty"`
	SearchText string   `json:"search_text"`
}

// Registry is a thread-safe, name-indexed collection of tool specs.
type Registry struct {
	mu    sync.RWMutex
	specs map[string]Spec
	order []string
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry {
	return &Registry{specs: make(map[string]Spec)}
}

// Register validates and adds specs. Duplicate or malformed specs are rejected.
func (r *Registry) Register(specs ...Spec) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, spec := range specs {
		normalized, err := normalizeSpec(spec)
		if err != nil {
			return err
		}
		if _, exists := r.specs[normalized.Tool.Name]; exists {
			return fmt.Errorf("duplicate tool name %q", normalized.Tool.Name)
		}
		r.specs[normalized.Tool.Name] = normalized
		r.order = append(r.order, normalized.Tool.Name)
	}

	sort.Strings(r.order)
	return nil
}

func normalizeSpec(spec Spec) (Spec, error) {
	name := strings.TrimSpace(spec.Tool.Name)
	if name == "" {
		return Spec{}, fmt.Errorf("tool name must not be empty")
	}
	if spec.Handler == nil {
		return Spec{}, fmt.Errorf("tool %q has no handler", name)
	}

	spec.Tool.Name = name
	if spec.Domain == "" {
		spec.Domain = "general"
	}
	if spec.Risk == "" {
		spec.Risk = RiskRead
	}
	if spec.Verification == "" {
		spec.Verification = VerificationNone
	}
	if spec.Summary == "" {
		spec.Summary = firstSentence(spec.Tool.Description)
	}
	if spec.Verification != VerificationNone {
		spec.Invokable = false
	}
	if spec.ReadOnly && spec.Risk != RiskRead {
		spec.ReadOnly = false
	}
	return spec, nil
}

// Get returns one spec by exact name.
func (r *Registry) Get(name string) (Spec, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	spec, ok := r.specs[name]
	return spec, ok
}

// List returns all specs in stable name order.
func (r *Registry) List() []Spec {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Spec, 0, len(r.order))
	for _, name := range r.order {
		result = append(result, r.specs[name])
	}
	return result
}

// Names returns all registered names in stable order.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]string(nil), r.order...)
}

// Documents returns the searchable representation of every registered tool.
func (r *Registry) Documents() []Document {
	specs := r.List()
	documents := make([]Document, 0, len(specs))
	for _, spec := range specs {
		required := append([]string(nil), spec.Tool.InputSchema.Required...)
		parts := []string{
			spec.Tool.Name,
			spec.Summary,
			spec.Tool.Description,
			spec.Domain,
		}
		parts = append(parts, spec.Tags...)
		parts = append(parts, spec.Aliases...)
		if len(required) > 0 {
			parts = append(parts, "required arguments: "+strings.Join(required, ", "))
		}

		documents = append(documents, Document{
			Name:       spec.Tool.Name,
			Summary:    spec.Summary,
			Domain:     spec.Domain,
			Tags:       append([]string(nil), spec.Tags...),
			Aliases:    append([]string(nil), spec.Aliases...),
			SearchText: strings.Join(parts, " "),
		})
	}
	return documents
}

// RequiredArguments returns the required argument names declared by the schema.
func (s Spec) RequiredArguments() []string {
	return append([]string(nil), s.Tool.InputSchema.Required...)
}

func firstSentence(description string) string {
	description = strings.TrimSpace(description)
	if description == "" {
		return ""
	}
	if idx := strings.IndexAny(description, ".\n"); idx >= 0 {
		sentence := strings.TrimSpace(description[:idx])
		if sentence != "" {
			return sentence
		}
	}
	return description
}
