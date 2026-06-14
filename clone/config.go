// Package clone is kage's engine: it ties the Chrome pool, the JavaScript
// stripper, the asset localiser, and the URL↔path mapper into one resumable,
// polite crawl that turns a live site into a browsable offline folder.
package clone

import (
	"os"
	"path/filepath"
	"time"

	"github.com/tamnd/kage/urlx"
)

// DefaultOutDir is where mirrors land unless --out overrides it: a per-user data
// directory ($HOME/data/kage) so clones from anywhere collect in one place,
// falling back to a local kage-out when the home directory cannot be resolved.
func DefaultOutDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "kage-out"
	}
	return filepath.Join(home, "data", "kage")
}

// Config is the full set of knobs for a clone run. DefaultConfig fills the
// baseline; the CLI overlays flags on top.
type Config struct {
	OutDir   string // output root; the mirror lands in <OutDir>/<host>/
	Reserved string // reserved dir name for assets and state (default "_kage")

	Workers       int // page render workers
	AssetWorkers  int // HTTP asset download workers
	BrowserPages  int // Chrome page-pool size
	MaxPages      int // stop after N pages (0 = unlimited)
	MaxDepth      int // BFS/DFS depth cap (0 = unlimited)
	Traversal     string
	MaxAssetBytes int64

	Timeout       time.Duration // per HTTP request
	Settle        time.Duration // network-idle quiet period
	RenderTimeout time.Duration // hard cap per page render
	Scroll        bool

	UserAgent         string
	IncludeSubdomains bool
	ScopePrefix       string
	ExcludePaths      []string

	RespectRobots bool
	FollowSitemap bool
	Headless      bool
	KeepNoscript  bool
	ChromeBin     string
	ControlURL    string

	// Resume loads the prior run's visited set and skips pages already written,
	// so an interrupted or repeated clone picks up where it left off instead of
	// refetching. Refresh forces every page to be re-rendered in place (the
	// mirror is kept, files are overwritten) to pull in changed content. Force
	// deletes the mirror first for a clean-slate clone. Persist writes the
	// visited set back to state.json when the run ends.
	Resume  bool
	Refresh bool
	Force   bool
	Persist bool
}

// DefaultUserAgent is a current desktop Chrome UA, used by the asset fetcher and
// the robots fetch so a site treats kage like the browser it drives.
const DefaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) " +
	"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"

// DefaultConfig returns the baseline configuration.
func DefaultConfig() Config {
	return Config{
		OutDir:        DefaultOutDir(),
		Reserved:      urlx.DefaultReserved,
		Workers:       4,
		AssetWorkers:  8,
		BrowserPages:  4,
		MaxAssetBytes: 25 << 20,
		Traversal:     "bfs",
		Timeout:       30 * time.Second,
		Settle:        1500 * time.Millisecond,
		RenderTimeout: 30 * time.Second,
		UserAgent:     DefaultUserAgent,
		RespectRobots: true,
		FollowSitemap: true,
		Headless:      true,
		Resume:        true,
		Persist:       true,
	}
}

// HostDir returns the mirror root for a seed host: <OutDir>/<host>.
func (c Config) HostDir(host string) string {
	return filepath.Join(c.OutDir, host)
}

// scope builds the urlx scope config from the run config.
func (c Config) scope() urlx.ScopeConfig {
	return urlx.ScopeConfig{
		IncludeSubdomains: c.IncludeSubdomains,
		ScopePrefix:       c.ScopePrefix,
		ExcludePaths:      c.ExcludePaths,
	}
}
