package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const AllowedFileRootsEnv = "GITEE_ENT_MCP_ALLOWED_FILE_ROOTS"

func ResolveReadableFile(path string) (string, error) {
	if !filepath.IsAbs(path) {
		return "", NewParamError("file_paths", "file paths must be absolute")
	}

	resolved, err := filepath.EvalSymlinks(filepath.Clean(path))
	if err != nil {
		return "", NewParamError("file_paths", fmt.Sprintf("cannot resolve %q: %v", path, err))
	}
	if err := ensureAllowedPath(resolved); err != nil {
		return "", err
	}

	info, err := os.Stat(resolved)
	if err != nil {
		return "", NewParamError("file_paths", fmt.Sprintf("cannot access %q: %v", path, err))
	}
	if !info.Mode().IsRegular() {
		return "", NewParamError("file_paths", fmt.Sprintf("%q is not a regular file", path))
	}

	return resolved, nil
}

func ResolveWritableDirectory(path string) (string, error) {
	if !filepath.IsAbs(path) {
		return "", NewParamError("output_directory", "output directory must be an absolute path")
	}

	resolved, err := filepath.EvalSymlinks(filepath.Clean(path))
	if err != nil {
		return "", NewParamError("output_directory", fmt.Sprintf("cannot resolve %q: %v", path, err))
	}
	if err := ensureAllowedPath(resolved); err != nil {
		return "", err
	}

	info, err := os.Stat(resolved)
	if err != nil {
		return "", NewParamError("output_directory", fmt.Sprintf("cannot access %q: %v", path, err))
	}
	if !info.IsDir() {
		return "", NewParamError("output_directory", fmt.Sprintf("%q is not a directory", path))
	}

	return resolved, nil
}

func allowedFileRoots() ([]string, error) {
	configured := strings.TrimSpace(os.Getenv(AllowedFileRootsEnv))
	if configured == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, NewInternalError(fmt.Errorf("get current working directory: %w", err))
		}
		configured = cwd
	}

	parts := filepath.SplitList(configured)
	roots := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if !filepath.IsAbs(part) {
			return nil, NewParamError(AllowedFileRootsEnv, fmt.Sprintf("allowed root %q must be absolute", part))
		}
		resolved, err := filepath.EvalSymlinks(filepath.Clean(part))
		if err != nil {
			return nil, NewParamError(AllowedFileRootsEnv, fmt.Sprintf("cannot resolve allowed root %q: %v", part, err))
		}
		roots = append(roots, resolved)
	}
	if len(roots) == 0 {
		return nil, NewParamError(AllowedFileRootsEnv, "at least one allowed root is required")
	}
	return roots, nil
}

func ensureAllowedPath(path string) error {
	roots, err := allowedFileRoots()
	if err != nil {
		return err
	}

	for _, root := range roots {
		rel, err := filepath.Rel(root, path)
		if err != nil {
			continue
		}
		if rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))) {
			return nil
		}
	}

	return NewParamError("path", fmt.Sprintf("%q is outside %s", path, AllowedFileRootsEnv))
}
