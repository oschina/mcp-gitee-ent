package urlbuilder

import "testing"

func newTestBuilder() Builder {
	return New("oschina").WithBase("https://e.gitee.com")
}

func TestBuilderURLs(t *testing.T) {
	b := newTestBuilder()

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"issue", b.Issue("I5DFIB"), "https://e.gitee.com/oschina/issues/list?issue=I5DFIB"},
		{"issue empty", b.Issue(""), ""},
		{"pull request", b.PullRequest("oschina/gitee", 24300), "https://e.gitee.com/oschina/repos/oschina/gitee/pulls/24300"},
		{"pull request subgroup", b.PullRequest("oschina/sub/gitee", 1), "https://e.gitee.com/oschina/repos/oschina/sub/gitee/-/pulls/1"},
		{"repository", b.Repository("oschina/gitee"), "https://e.gitee.com/oschina/repos/oschina/gitee"},
		{"repository subgroup", b.Repository("oschina/sub/gitee"), "https://e.gitee.com/oschina/repos/oschina/sub/gitee/-"},
		{"repo web path", b.RepoWebPath("/oschina/gitee/blob/master/README.md"), "https://e.gitee.com/oschina/repos/oschina/gitee/blob/master/README.md"},
		{"repo web path subgroup", b.RepoWebPath("/oschina/sub/gitee/tree/master/docs"), "https://e.gitee.com/oschina/repos/oschina/sub/gitee/-/tree/master/docs"},
		{"repo web path root", b.RepoWebPath("/oschina/gitee"), "https://e.gitee.com/oschina/repos/oschina/gitee"},
		{"program overview", b.ProgramOverview(776601), "https://e.gitee.com/oschina/projects/776601/overview"},
		{"sprint overview", b.SprintOverview(1, 2), "https://e.gitee.com/oschina/projects/1/sprints/2/overview"},
		{"version", b.Version(1, 3), "https://e.gitee.com/oschina/projects/1/versions/3"},
		{"baseline", b.Baseline(1, 4), "https://e.gitee.com/oschina/projects/1/baseline/4/issues"},
		{"test case", b.TestCase(1, "CP6N5"), "https://e.gitee.com/oschina/projects/1/tests/test_cases?test_case=CP6N5"},
		{"test plan", b.TestPlan(1, 5), "https://e.gitee.com/oschina/projects/1/tests/test_plans/5"},
		{"test report", b.TestReport(1, 6), "https://e.gitee.com/oschina/projects/1/tests/test_reports/6"},
		{"milestone enterprise", b.MilestoneEnterprise(9), "https://e.gitee.com/oschina/milestones/9"},
		{"milestone program", b.MilestoneProgram(1, 9), "https://e.gitee.com/oschina/projects/1/milestones/9"},
		{"milestone repo", b.MilestoneRepo("oschina/gitee", 9), "https://e.gitee.com/oschina/repos/oschina/gitee/milestones/9"},
		{"document enterprise", b.Document(0, 4168359, 8512977), "https://e.gitee.com/oschina/docs/4168359/file/8512977"},
		{"document program", b.Document(776601, 1, 2), "https://e.gitee.com/oschina/projects/776601/docs/1/file/2"},
		{"group", b.Group("sub/team"), "https://e.gitee.com/oschina/members/team/sub/team"},
		{"kanban enterprise", b.Kanban(0), "https://e.gitee.com/oschina/issues/kanban"},
		{"kanban program", b.Kanban(1), "https://e.gitee.com/oschina/projects/1/issues/kanban"},
		{"release", b.Release("oschina/gitee", "v1.0.0"), "https://e.gitee.com/oschina/repos/oschina/gitee/releases/v1.0.0"},
		{"commit", b.Commit("oschina/gitee", "abc123"), "https://e.gitee.com/oschina/repos/oschina/gitee/commit/abc123"},
		{"wiki", b.Wiki("oschina/gitee"), "https://e.gitee.com/oschina/repos/oschina/gitee/wiki"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("got %q, want %q", tc.got, tc.want)
			}
		})
	}
}

func TestSubGroupSuffix(t *testing.T) {
	tests := map[string]string{
		"oschina/gitee":     "",
		"oschina/sub/gitee": "/-",
		"a/b/c/d":           "/-",
		"":                  "",
	}
	for path, want := range tests {
		if got := SubGroupSuffix(path); got != want {
			t.Fatalf("SubGroupSuffix(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestSplitRepoWebPath(t *testing.T) {
	tests := []struct {
		relPath  string
		wantRepo string
		wantRest string
	}{
		{"/oschina/gitee/blob/master/README.md", "oschina/gitee", "blob/master/README.md"},
		{"/oschina/sub/gitee/tree/master/docs", "oschina/sub/gitee", "tree/master/docs"},
		{"/oschina/gitee", "oschina/gitee", ""},
		{"/oschina/gitee/pulls/24300", "oschina/gitee", "pulls/24300"},
		{"/gitee/blob/master/README.md", "oschina/gitee", "blob/master/README.md"},
	}
	for _, tc := range tests {
		repo, rest := SplitRepoWebPath("oschina", tc.relPath)
		if repo != tc.wantRepo || rest != tc.wantRest {
			t.Fatalf("SplitRepoWebPath(%q) = (%q, %q), want (%q, %q)", tc.relPath, repo, rest, tc.wantRepo, tc.wantRest)
		}
	}
}

func TestConvertCommunityURL(t *testing.T) {
	b := newTestBuilder()

	tests := []struct {
		name   string
		input  string
		want   string
		wantOK bool
	}{
		{"pull", "https://gitee.com/oschina/gitee/pulls/24300", "https://e.gitee.com/oschina/repos/oschina/gitee/pulls/24300", true},
		{"issue", "https://gitee.com/oschina/gitee/issues/I5DFIB", "https://e.gitee.com/oschina/issues/list?issue=I5DFIB", true},
		{"blob", "https://gitee.com/oschina/gitee/blob/master/README.md", "https://e.gitee.com/oschina/repos/oschina/gitee/blob/master/README.md", true},
		{"repo", "https://gitee.com/oschina/gitee", "https://e.gitee.com/oschina/repos/oschina/gitee", true},
		{"other enterprise", "https://gitee.com/other/gitee/pulls/1", "", false},
		{"other host", "https://example.com/oschina/gitee/pulls/1", "", false},
		{"empty", "", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := b.Convert(tc.input)
			if ok != tc.wantOK || got != tc.want {
				t.Fatalf("Convert(%q) = (%q, %v), want (%q, %v)", tc.input, got, ok, tc.want, tc.wantOK)
			}
		})
	}
}

func TestWebBasePrecedence(t *testing.T) {
	original := webBase
	t.Cleanup(func() { webBase = original })

	webBase = ""
	t.Setenv(EnvWebBase, "edu.gitee.com/")
	if got := WebBase(); got != "https://edu.gitee.com" {
		t.Fatalf("env web base = %q", got)
	}

	SetWebBase("https://custom.example.com/")
	if got := WebBase(); got != "https://custom.example.com" {
		t.Fatalf("flag web base = %q", got)
	}

	SetWebBase("")
	t.Setenv(EnvWebBase, "")
	if got := WebBase(); got != DefaultWebBase {
		t.Fatalf("default web base = %q", got)
	}
}
