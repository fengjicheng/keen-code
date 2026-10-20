package filesearch

import (
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/mochow13/keen-code/internal/filesystem"
)

func TestSearchTrackedFilesAndDirectories(t *testing.T) {
	requireGit(t)
	dir := t.TempDir()
	writeFile(t, dir, "internal/pkg/file.go", "package pkg")
	writeFile(t, dir, "README.md", "readme")
	runGit(t, dir, "init")
	runGit(t, dir, "add", ".")

	searcher := NewFileSearcher(dir, nil)
	got := searcher.Search("PKG", 10)
	if !containsPath(got, "internal/pkg") || !containsPath(got, "internal/pkg/file.go") {
		t.Fatalf("Search() = %q, want directory and file", got)
	}
	if got := searcher.Search("", 10); got != nil {
		t.Fatalf("empty Search() = %q, want nil", got)
	}
	if got := searcher.Search(".", 1); len(got) != 1 {
		t.Fatalf("limited Search() = %q, want one result", got)
	}
}

func TestSearchParentAndSiblingTraversal(t *testing.T) {
	root := t.TempDir()
	workDir := filepath.Join(root, "current-repo")
	writeFile(t, workDir, "main.go", "package main")
	writeFile(t, root, "sibling-repo/service.go", "package service")
	writeFile(t, root, "sibling-repo/sub/helper.go", "package sub")
	writeFile(t, root, "sibling-repo/.secret", "hidden")
	writeFile(t, root, "notes.txt", "notes")

	searcher := NewFileSearcher(workDir, nil)

	parentEntries := searcher.Search("..", 10)
	if !containsPath(parentEntries, "../sibling-repo/") || !containsPath(parentEntries, "../notes.txt") {
		t.Fatalf("Search(\"..\") = %q, want ../sibling-repo/ and ../notes.txt", parentEntries)
	}

	siblingEntries := searcher.Search("../sibling-repo/", 10)
	if !containsPath(siblingEntries, "../sibling-repo/service.go") || !containsPath(siblingEntries, "../sibling-repo/sub/") {
		t.Fatalf("Search(\"../sibling-repo/\") = %q, want service.go and sub/", siblingEntries)
	}
	if containsPath(siblingEntries, "../sibling-repo/.secret") {
		t.Fatalf("Search(\"../sibling-repo/\") unexpectedly included hidden file: %q", siblingEntries)
	}

	hiddenEntries := searcher.Search("../sibling-repo/.", 10)
	if !containsPath(hiddenEntries, "../sibling-repo/.secret") {
		t.Fatalf("Search(\"../sibling-repo/.\") = %q, want ../sibling-repo/.secret", hiddenEntries)
	}

	filtered := searcher.Search("../sibling-repo/ser", 10)
	if len(filtered) != 1 || filtered[0] != "../sibling-repo/service.go" {
		t.Fatalf("Search(\"../sibling-repo/ser\") = %q, want [../sibling-repo/service.go]", filtered)
	}
}

func TestSearchHomeAndAbsolutePaths(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeFile(t, home, "docs/guide.md", "guide")
	writeFile(t, home, "todo.txt", "todo")
	writeFile(t, home, ".ssh/id_rsa", "private")

	workDir := t.TempDir()
	guard := filesystem.NewGuard(workDir, nil)
	searcher := NewFileSearcher(workDir, guard)

	homeEntries := searcher.Search("~/", 10)
	if !containsPath(homeEntries, "~/docs/") || !containsPath(homeEntries, "~/todo.txt") {
		t.Fatalf("Search(\"~/\") = %q, want ~/docs/ and ~/todo.txt", homeEntries)
	}

	// Guard blocks ~/.<hidden> directories even when explicitly queried with leading dot.
	if got := searcher.Search("~/.ssh/", 10); len(got) != 0 {
		t.Fatalf("Search(\"~/.ssh/\") = %q, want blocked empty result", got)
	}

	absQuery := filepath.Join(home, "docs") + "/gu"
	absEntries := searcher.Search(absQuery, 10)
	wantAbs := filepath.Join(home, "docs", "guide.md")
	if !containsPath(absEntries, wantAbs) {
		t.Fatalf("Search(%q) = %q, want %q", absQuery, absEntries, wantAbs)
	}
}

func TestSearchNonGitWorkingDirFallback(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "alpha.go", "package main")
	writeFile(t, dir, "nested/beta.go", "package nested")

	searcher := NewFileSearcher(dir, nil)
	got := searcher.Search("alp", 10)
	if !containsPath(got, "alpha.go") {
		t.Fatalf("non-git Search(\"alp\") = %q, want alpha.go", got)
	}
	nested := searcher.Search("nested/", 10)
	if !containsPath(nested, "nested/beta.go") {
		t.Fatalf("non-git Search(\"nested/\") = %q, want nested/beta.go", nested)
	}
}

func TestRefreshAddsUntrackedFiles(t *testing.T) {
	requireGit(t)
	dir := t.TempDir()
	writeFile(t, dir, "tracked.txt", "tracked")
	runGit(t, dir, "init")
	runGit(t, dir, "add", "tracked.txt")
	searcher := NewFileSearcher(dir, nil)
	_ = searcher.Search("tracked", 10)
	writeFile(t, dir, "new-file.txt", "new")
	searcher.refreshInterval = 0
	searcher.updatedAt = time.Time{}
	_ = searcher.Search("new-file", 10)

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if containsPath(searcher.Search("new-file", 10), "new-file.txt") {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("background refresh did not add untracked file")
}

func TestGitLsFilesFailure(t *testing.T) {
	if paths, ok := gitLsFiles(t.TempDir()); ok || paths != nil {
		t.Fatalf("gitLsFiles() = %q, %v, want nil false", paths, ok)
	}
}

func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
}
