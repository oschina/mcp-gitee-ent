package toolspecs

import "testing"

func TestAllSpecs(t *testing.T) {
	specs := All()
	if len(specs) != 140 {
		t.Fatalf("specs = %d, want 140", len(specs))
	}
	visible := 0
	for _, spec := range specs {
		if spec.Visible {
			visible++
		}
	}
	if visible != 40 {
		t.Fatalf("visible = %d, want 40", visible)
	}
}

func TestBuildRespectsAllowlist(t *testing.T) {
	registry, err := Build(func(name string) bool { return name == "list_enterprise_issues" })
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	spec, ok := registry.Get("list_enterprise_issues")
	if !ok || !spec.Visible || !spec.Invokable {
		t.Fatalf("list_enterprise_issues = %#v", spec)
	}
	hidden, _ := registry.Get("create_test_report")
	if hidden.Visible || hidden.Invokable {
		t.Fatalf("create_test_report should be hidden/not invokable: %#v", hidden)
	}
}
