// Package urlbuilder builds Gitee enterprise (new frontend) web URLs from the
// identifiers that the enterprise API returns.
//
// Route templates mirror the enterprise frontend router, which is the source
// of truth:
//
//	gitee-ent-web/packages/gitee-ent-web/src/utils/linkTo.js
//
// The legacy Rails implementation lives in
// gitee/app/services/enterprises/url_transform/static_path/*.rb and is kept
// here only as a cross-check.
package urlbuilder

import (
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	// DefaultWebBase is the default enterprise web origin.
	DefaultWebBase = "https://e.gitee.com"

	// EnvWebBase allows overriding the enterprise web origin without a flag.
	EnvWebBase = "GITEE_ENT_WEB_BASE"
)

var webBase string

// SetWebBase overrides the enterprise web origin, e.g. "https://edu.gitee.com".
// An empty value falls back to EnvWebBase / DefaultWebBase.
func SetWebBase(value string) {
	webBase = NormalizeBase(value)
}

// WebBase returns the configured enterprise web origin without a trailing slash.
// Precedence: SetWebBase > GITEE_ENT_WEB_BASE > DefaultWebBase.
func WebBase() string {
	if webBase != "" {
		return webBase
	}
	if value := strings.TrimSpace(os.Getenv(EnvWebBase)); value != "" {
		return NormalizeBase(value)
	}
	return DefaultWebBase
}

// NormalizeBase trims spaces and trailing slashes and adds a scheme when the
// value looks like a bare host.
func NormalizeBase(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimRight(value, "/")
	if value == "" {
		return ""
	}
	if !strings.Contains(value, "://") {
		value = "https://" + value
	}
	return value
}

// Builder renders enterprise frontend URLs for one enterprise.
type Builder struct {
	Base    string
	EntPath string
}

// New returns a Builder for the given enterprise path (for example "oschina").
func New(entPath string) Builder {
	return Builder{Base: WebBase(), EntPath: strings.Trim(entPath, "/")}
}

// WithBase returns a copy of the builder pinned to a different web origin.
func (b Builder) WithBase(base string) Builder {
	b.Base = NormalizeBase(base)
	return b
}

// WebBasePath returns the enterprise root, e.g. https://e.gitee.com/oschina.
func (b Builder) WebBasePath() string {
	return joinURL(b.Base, b.EntPath)
}

func (b Builder) path(parts ...string) string {
	all := make([]string, 0, len(parts)+1)
	all = append(all, b.EntPath)
	all = append(all, parts...)
	return joinURL(b.Base, all...)
}

// RepoLink returns the enterprise-scoped repository root, honouring the
// "/-" separator that the new frontend requires for subgroup paths.
func (b Builder) RepoLink(pathWithNamespace string) string {
	pathWithNamespace = cleanPath(pathWithNamespace)
	if pathWithNamespace == "" {
		return ""
	}
	return b.path("repos", pathWithNamespace) + SubGroupSuffix(pathWithNamespace)
}

// Repository returns the repository home page.
func (b Builder) Repository(pathWithNamespace string) string {
	return b.RepoLink(pathWithNamespace)
}

// PullRequest returns a pull request detail page.
func (b Builder) PullRequest(pathWithNamespace string, iid int) string {
	link := b.RepoLink(pathWithNamespace)
	if link == "" || iid <= 0 {
		return ""
	}
	return link + "/pulls/" + itoa(iid)
}

// PullRequestList returns a repository pull request list page.
func (b Builder) PullRequestList(pathWithNamespace string) string {
	link := b.RepoLink(pathWithNamespace)
	if link == "" {
		return ""
	}
	return link + "/pulls"
}

// Issue returns the enterprise issue page. Both enterprise-level and
// project-level issues use the same route, keyed by the issue ident.
func (b Builder) Issue(ident string) string {
	ident = strings.TrimSpace(ident)
	if ident == "" {
		return ""
	}
	return b.path("issues/list") + "?issue=" + url.QueryEscape(ident)
}

// IssueTable returns the enterprise issue table page, optionally filtered by ident.
func (b Builder) IssueTable(ident string) string {
	page := b.path("issues/table")
	if strings.TrimSpace(ident) == "" {
		return page
	}
	return page + "?issue=" + url.QueryEscape(ident)
}

// Commit returns a commit detail page.
func (b Builder) Commit(pathWithNamespace, sha string) string {
	link := b.RepoLink(pathWithNamespace)
	if link == "" || strings.TrimSpace(sha) == "" {
		return ""
	}
	return link + "/commit/" + url.PathEscape(sha)
}

// Tree returns a directory page.
func (b Builder) Tree(pathWithNamespace, ref, dir string) string {
	link := b.RepoLink(pathWithNamespace)
	if link == "" {
		return ""
	}
	out := link + "/tree/" + url.PathEscape(strings.Trim(ref, "/"))
	if dir = strings.Trim(dir, "/"); dir != "" {
		out += "/" + dir
	}
	return out
}

// Blob returns a file page.
func (b Builder) Blob(pathWithNamespace, ref, file string) string {
	link := b.RepoLink(pathWithNamespace)
	if link == "" {
		return ""
	}
	return link + "/blob/" + url.PathEscape(strings.Trim(ref, "/")) + "/" + strings.TrimLeft(file, "/")
}

// Release returns a release detail page.
func (b Builder) Release(pathWithNamespace, tagVersion string) string {
	link := b.RepoLink(pathWithNamespace)
	if link == "" || strings.TrimSpace(tagVersion) == "" {
		return ""
	}
	return link + "/releases/" + url.PathEscape(tagVersion)
}

// Wiki returns a repository wiki page.
func (b Builder) Wiki(pathWithNamespace string) string {
	link := b.RepoLink(pathWithNamespace)
	if link == "" {
		return ""
	}
	return link + "/wiki"
}

// ReleaseList returns a repository release list page.
func (b Builder) ReleaseList(pathWithNamespace string) string {
	link := b.RepoLink(pathWithNamespace)
	if link == "" {
		return ""
	}
	return link + "/releases"
}

// Program returns a program (项目) root page.
func (b Builder) Program(programID int) string {
	if programID <= 0 {
		return ""
	}
	return b.path("projects", itoa(programID))
}

// ProgramOverview returns a program overview page.
func (b Builder) ProgramOverview(programID int) string {
	root := b.Program(programID)
	if root == "" {
		return ""
	}
	return root + "/overview"
}

// Sprint returns a sprint detail page.
func (b Builder) Sprint(programID, sprintID int) string {
	root := b.Program(programID)
	if root == "" || sprintID <= 0 {
		return ""
	}
	return root + "/sprints/" + itoa(sprintID)
}

// SprintOverview returns a sprint overview page.
func (b Builder) SprintOverview(programID, sprintID int) string {
	link := b.Sprint(programID, sprintID)
	if link == "" {
		return ""
	}
	return link + "/overview"
}

// Version returns a scrum version page.
func (b Builder) Version(programID, versionID int) string {
	root := b.Program(programID)
	if root == "" || versionID <= 0 {
		return ""
	}
	return root + "/versions/" + itoa(versionID)
}

// Baseline returns a program baseline issues page.
func (b Builder) Baseline(programID, baselineID int) string {
	root := b.Program(programID)
	if root == "" || baselineID <= 0 {
		return ""
	}
	return root + "/baseline/" + itoa(baselineID) + "/issues"
}

// TestCase returns a test case library page focused on one case.
func (b Builder) TestCase(programID int, ident string) string {
	root := b.Program(programID)
	if root == "" || strings.TrimSpace(ident) == "" {
		return ""
	}
	return root + "/tests/test_cases?test_case=" + url.QueryEscape(ident)
}

// TestCaseLibrary returns the test case library page.
func (b Builder) TestCaseLibrary(programID int) string {
	root := b.Program(programID)
	if root == "" {
		return ""
	}
	return root + "/tests/test_cases"
}

// TestPlan returns a test plan detail page.
func (b Builder) TestPlan(programID, testPlanID int) string {
	root := b.Program(programID)
	if root == "" || testPlanID <= 0 {
		return ""
	}
	return root + "/tests/test_plans/" + itoa(testPlanID)
}

// TestPlanCase returns a test plan page focused on one case.
func (b Builder) TestPlanCase(programID, testPlanID int, ident string) string {
	link := b.TestPlan(programID, testPlanID)
	if link == "" || strings.TrimSpace(ident) == "" {
		return link
	}
	return link + "?test_case=" + url.QueryEscape(ident)
}

// TestReport returns a test report detail page.
func (b Builder) TestReport(programID, reportID int) string {
	root := b.Program(programID)
	if root == "" || reportID <= 0 {
		return ""
	}
	return root + "/tests/test_reports/" + itoa(reportID)
}

// TestReview returns a test review detail page.
func (b Builder) TestReview(programID, reviewID int) string {
	root := b.Program(programID)
	if root == "" || reviewID <= 0 {
		return ""
	}
	return root + "/tests/test_review/" + itoa(reviewID)
}

// MilestoneEnterprise returns an enterprise-level milestone page.
func (b Builder) MilestoneEnterprise(milestoneID int) string {
	if milestoneID <= 0 {
		return ""
	}
	return b.path("milestones", itoa(milestoneID))
}

// MilestoneProgram returns a program-level milestone page.
func (b Builder) MilestoneProgram(programID, milestoneID int) string {
	root := b.Program(programID)
	if root == "" || milestoneID <= 0 {
		return ""
	}
	return root + "/milestones/" + itoa(milestoneID)
}

// MilestoneRepo returns a repository-level milestone page.
func (b Builder) MilestoneRepo(pathWithNamespace string, milestoneID int) string {
	link := b.RepoLink(pathWithNamespace)
	if link == "" || milestoneID <= 0 {
		return ""
	}
	return link + "/milestones/" + itoa(milestoneID)
}

// Document returns a knowledge-base document page. programID may be 0 for
// enterprise-level documents.
func (b Builder) Document(programID, docNodeID, fileID int) string {
	if docNodeID <= 0 || fileID <= 0 {
		return ""
	}
	detail := "docs/" + itoa(docNodeID) + "/file/" + itoa(fileID)
	if programID > 0 {
		return b.path("projects", itoa(programID), detail)
	}
	return b.path(detail)
}

// DocumentLibrary returns the knowledge-base root.
func (b Builder) DocumentLibrary(programID int) string {
	if programID > 0 {
		return b.path("projects", itoa(programID), "docs")
	}
	return b.path("docs")
}

// Group returns a team (group) page.
func (b Builder) Group(groupPath string) string {
	groupPath = cleanPath(groupPath)
	if groupPath == "" {
		return ""
	}
	return b.path("members/team", groupPath)
}

// Members returns the enterprise member list page.
func (b Builder) Members() string {
	return b.path("members")
}

// Kanban returns a kanban board page. programID may be 0 for the enterprise board.
func (b Builder) Kanban(programID int) string {
	if programID > 0 {
		return b.path("projects", itoa(programID), "issues/kanban")
	}
	return b.path("issues/kanban")
}

// Statistics returns an enterprise statistics page.
func (b Builder) Statistics(kind string) string {
	kind = strings.Trim(kind, "/")
	if kind == "" {
		return b.path("statistics")
	}
	return b.path("statistics", kind)
}

// Admin returns an enterprise admin page.
func (b Builder) Admin(parts ...string) string {
	return b.path(append([]string{"admin"}, parts...)...)
}

// RepoWebPath converts a relative repository web path returned by the API
// (for example "/oschina/gitee/blob/master/README.md") into a full enterprise
// URL by inserting the "/repos" segment and the subgroup separator.
func (b Builder) RepoWebPath(relPath string) string {
	pathWithNamespace, rest := SplitRepoWebPath(b.EntPath, relPath)
	if pathWithNamespace == "" {
		return ""
	}
	link := b.RepoLink(pathWithNamespace)
	if link == "" {
		return ""
	}
	if rest == "" {
		return link
	}
	return link + "/" + rest
}

// SubGroupSuffix returns "/-" for subgroup paths (more than two segments),
// matching getSubGroupSuffix() in the frontend.
func SubGroupSuffix(pathWithNamespace string) string {
	if IsSubGroupPath(pathWithNamespace) {
		return "/-"
	}
	return ""
}

// IsSubGroupPath reports whether the path belongs to a repository group.
func IsSubGroupPath(pathWithNamespace string) bool {
	return len(strings.Split(cleanPath(pathWithNamespace), "/")) > 2
}

// repoActionMarkers are the path segments that follow a repository path in a
// web URL. They delimit path_with_namespace from the rest of the route.
var repoActionMarkers = []string{
	"/tree/", "/blob/", "/commit/", "/pulls", "/issues", "/releases",
	"/wiki", "/milestones", "/branches", "/tags", "/sources", "/settings",
	"/members", "/overview",
}

// SplitRepoWebPath splits a relative web path into path_with_namespace and the
// remaining route. The enterprise path is the repository namespace, so it is
// kept as the leading segment of path_with_namespace; when it is missing it is
// prepended.
func SplitRepoWebPath(entPath, relPath string) (string, string) {
	path := cleanPath(relPath)
	entPath = cleanPath(entPath)
	if path == "" {
		return "", ""
	}
	if entPath != "" {
		if path == entPath {
			return "", ""
		}
		if !strings.HasPrefix(path, entPath+"/") {
			path = entPath + "/" + path
		}
	}

	index := -1
	for _, marker := range repoActionMarkers {
		if i := strings.Index("/"+path, marker); i >= 0 && (index < 0 || i < index) {
			index = i
		}
	}
	if index < 0 {
		return path, ""
	}
	repoPath := strings.Trim(path[:index], "/")
	rest := strings.Trim(path[index:], "/")
	return repoPath, rest
}

// Convert rewrites a community (gitee.com) web URL into the equivalent
// enterprise URL. The second return value reports whether a rewrite happened.
func (b Builder) Convert(communityURL string) (string, bool) {
	parsed, err := url.Parse(strings.TrimSpace(communityURL))
	if err != nil || parsed.Host != "gitee.com" || parsed.Host == "" {
		return "", false
	}
	path := cleanPath(parsed.Path)
	entPath := cleanPath(b.EntPath)
	if entPath == "" || !strings.HasPrefix(path, entPath+"/") {
		return "", false
	}
	segments := strings.Split(path, "/")
	// Community issue: /{owner}/{repo}/issues/{ident}
	for i := 0; i+1 < len(segments); i++ {
		if segments[i] == "issues" && segments[i+1] != "" {
			if converted := b.Issue(segments[i+1]); converted != "" {
				return converted, true
			}
		}
	}

	if converted := b.RepoWebPath(path); converted != "" {
		return converted, true
	}
	return "", false
}

func joinURL(base string, parts ...string) string {
	out := strings.TrimRight(base, "/")
	for _, part := range parts {
		part = strings.Trim(part, "/")
		if part == "" {
			continue
		}
		out += "/" + part
	}
	return out
}

func cleanPath(value string) string {
	return strings.Trim(strings.TrimSpace(value), "/")
}

func itoa(value int) string {
	if value <= 0 {
		return ""
	}
	return strconv.Itoa(value)
}
