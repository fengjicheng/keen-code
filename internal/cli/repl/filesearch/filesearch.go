package filesearch

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mochow13/keen-code/internal/filesystem"
)

const defaultRefreshInterval = 10 * time.Second

type FileSearcher struct {
	workingDir      string
	guard           *filesystem.Guard
	refreshInterval time.Duration
	mu              sync.RWMutex
	cache           []string
	cached          bool
	refreshing      bool
	updatedAt       time.Time
}

func NewFileSearcher(workingDir string, guard *filesystem.Guard) *FileSearcher {
	return &FileSearcher{
		workingDir:      workingDir,
		guard:           guard,
		refreshInterval: defaultRefreshInterval,
	}
}

// Search returns up to limit paths matching query.
// Path traversal queries (starting with /, ~/, ./, ../) list directory contents on demand.
// Workspace queries search the cached git file list (with background refresh), falling back
// to on-demand directory listing when workingDir has no git index.
func (s *FileSearcher) Search(query string, limit int) []string {
	if query == "" || limit <= 0 {
		return nil
	}

	if isPathTraversalQuery(query) {
		return s.searchDirectory(query, limit)
	}

	s.ensureCache()
	cache := s.snapshotCache()
	s.refreshStaleCache()

	if len(cache) == 0 {
		return s.searchDirectory(query, limit)
	}

	query = strings.ToLower(query)
	var results []string
	for _, p := range cache {
		if strings.Contains(strings.ToLower(p), query) {
			results = append(results, p)
			if len(results) >= limit {
				break
			}
		}
	}
	return results
}

func isPathTraversalQuery(query string) bool {
	return query == "~" ||
		query == "." ||
		query == ".." ||
		strings.HasPrefix(query, "/") ||
		strings.HasPrefix(query, "~/") ||
		strings.HasPrefix(query, "./") ||
		strings.HasPrefix(query, "../")
}

func (s *FileSearcher) searchDirectory(query string, limit int) []string {
	displayDir, resolvedDir, filePrefix, ok := s.resolveDirectoryQuery(query)
	if !ok {
		return nil
	}

	if s.guard != nil && s.guard.IsBlocked(resolvedDir) {
		return nil
	}

	entries, err := os.ReadDir(resolvedDir)
	if err != nil {
		return nil
	}

	lowerPrefix := strings.ToLower(filePrefix)
	includeHidden := strings.HasPrefix(filePrefix, ".")

	var prefixMatches []string
	var substringMatches []string

	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") && !includeHidden {
			continue
		}

		fullPath := filepath.Join(resolvedDir, name)
		if s.guard != nil && s.guard.IsBlocked(fullPath) {
			continue
		}

		lowerName := strings.ToLower(name)
		isPrefix := lowerPrefix == "" || strings.HasPrefix(lowerName, lowerPrefix)
		isSubstring := !isPrefix && strings.Contains(lowerName, lowerPrefix)
		if !isPrefix && !isSubstring {
			continue
		}

		candidate := displayDir + name
		if isDirEntry(resolvedDir, entry) {
			candidate += "/"
		}

		if isPrefix {
			prefixMatches = append(prefixMatches, candidate)
			if len(prefixMatches) >= limit {
				return prefixMatches[:limit]
			}
		} else {
			substringMatches = append(substringMatches, candidate)
		}
	}

	results := append(prefixMatches, substringMatches...)
	if len(results) > limit {
		results = results[:limit]
	}
	return results
}

func (s *FileSearcher) resolveDirectoryQuery(query string) (displayDir string, resolvedDir string, filePrefix string, ok bool) {
	if query == "~" || query == "." || query == ".." || strings.HasSuffix(query, "/..") {
		query += "/"
	}

	if lastSlash := strings.LastIndex(query, "/"); lastSlash >= 0 {
		displayDir = query[:lastSlash+1]
		filePrefix = query[lastSlash+1:]
	} else {
		filePrefix = query
	}

	targetDir := displayDir
	if targetDir == "" {
		targetDir = "."
	}
	resolved, err := filesystem.ResolvePath(s.workingDir, targetDir)
	if err != nil {
		return "", "", "", false
	}

	return displayDir, resolved, filePrefix, true
}

func isDirEntry(parentDir string, entry os.DirEntry) bool {
	if entry.IsDir() {
		return true
	}
	if entry.Type()&os.ModeSymlink != 0 {
		info, err := os.Stat(filepath.Join(parentDir, entry.Name()))
		return err == nil && info.IsDir()
	}
	return false
}

func (s *FileSearcher) ensureCache() {
	s.mu.RLock()
	cached := s.cached
	s.mu.RUnlock()
	if cached {
		return
	}

	paths, _ := gitLsFiles(s.workingDir)

	s.mu.Lock()
	if !s.cached {
		s.cache = paths
		s.cached = true
		s.updatedAt = time.Now()
	}
	s.mu.Unlock()
}

func (s *FileSearcher) snapshotCache() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cache
}

func (s *FileSearcher) refreshStaleCache() {
	s.mu.Lock()
	if s.refreshing || time.Since(s.updatedAt) < s.refreshInterval {
		s.mu.Unlock()
		return
	}
	s.refreshing = true
	s.mu.Unlock()

	go s.rebuildCache()
}

func (s *FileSearcher) rebuildCache() {
	paths, ok := gitLsFilesIncludingUntracked(s.workingDir)

	s.mu.Lock()
	defer s.mu.Unlock()
	if ok {
		s.cache = paths
		s.cached = true
	}
	s.updatedAt = time.Now()
	s.refreshing = false
}

func gitLsFiles(dir string) ([]string, bool) {
	return gitLsFilesWithArgs(dir, "--cached")
}

func gitLsFilesIncludingUntracked(dir string) ([]string, bool) {
	return gitLsFilesWithArgs(dir, "--cached", "--others", "--exclude-standard")
}

func gitLsFilesWithArgs(dir string, args ...string) ([]string, bool) {
	cmdArgs := append([]string{"ls-files", "-z"}, args...)
	cmd := exec.Command("git", cmdArgs...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, false
	}

	var paths []string
	seen := make(map[string]struct{})
	for _, path := range strings.Split(string(out), "\x00") {
		if path == "" {
			continue
		}
		if _, ok := seen[path]; !ok {
			paths = append(paths, path)
			seen[path] = struct{}{}
		}
		for d := filepath.Dir(path); d != "."; d = filepath.Dir(d) {
			if _, ok := seen[d]; ok {
				continue
			}
			paths = append(paths, d)
			seen[d] = struct{}{}
		}
	}

	return paths, true
}
