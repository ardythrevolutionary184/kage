# kage

[![ci](https://github.com/tamnd/kage/actions/workflows/ci.yml/badge.svg)](https://github.com/tamnd/kage/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/tamnd/kage)](https://github.com/tamnd/kage/releases/latest)
[![Go Reference](https://pkg.go.dev/badge/github.com/tamnd/kage.svg)](https://pkg.go.dev/github.com/tamnd/kage)
[![Go Report Card](https://goreportcard.com/badge/github.com/tamnd/kage)](https://goreportcard.com/report/github.com/tamnd/kage)
[![License](https://img.shields.io/github/license/tamnd/kage)](./LICENSE)

**kage** (影, "shadow") clones a website into a self-contained folder you can
browse offline, with all the JavaScript stripped out. It renders every page in
headless Chrome, snapshots the final rendered DOM, removes every script and
event handler, then downloads the CSS, images, and fonts and rewrites them to
local paths. The result looks like the live site but runs no code.

[Install](#install) • [Commands](#commands) • [Clone](#clone) • [Pack](#pack-it-into-one-file) • [Native viewer](#a-native-window-not-a-browser-tab) • [How it works](#how-it-works)

![kage cloning a site, packing it into one file, and serving it back offline](docs/static/demo.gif)

Saving a page with "Save As" gives you a copy that still phones home, still runs
analytics, and often renders blank because the markup is built by JavaScript at
runtime. kage takes the opposite approach: it drives a real browser, waits for
the page to settle, captures the DOM a human would have seen, and then strips
every script out of it. What lands on disk is inert. No tracking, no network
calls, no surprises, just a folder of `.html` files you can open straight from
disk or pack into a single file to hand to someone.

Full reference and guides live at [kage.tamnd.com](https://kage.tamnd.com).

## Install

```bash
go install github.com/tamnd/kage/cmd/kage@latest
```

Or grab a prebuilt archive, `.deb`/`.rpm`/`.apk` package, or checksum from the
[releases](https://github.com/tamnd/kage/releases), or run the container image,
which bundles Chromium:

```bash
docker run --rm -v "$PWD/out:/out" ghcr.io/tamnd/kage clone example.com
```

kage drives a real browser, so it needs Chrome or Chromium on the host. It finds
a system install automatically; point it at a specific binary with `--chrome` or
the `KAGE_CHROME` environment variable. The container image needs nothing extra.

Shell completion is built in: `kage completion bash|zsh|fish|powershell`.

## Commands

| Command | Does |
| --- | --- |
| `kage clone <url>` | render a site in headless Chrome and write a browsable, script-free mirror |
| `kage serve [dir]` | preview a cloned folder over a local HTTP server |
| `kage pack <mirror-dir>` | collapse a mirror into one ZIM archive, or a self-contained viewer binary |
| `kage open <file.zim>` | serve a packed ZIM back for offline reading |

## Clone

```bash
# Clone a whole site into $HOME/data/kage/<host>/
kage clone https://example.com

# Limit the crawl
kage clone example.com --max-pages 200 --max-depth 3

# Only a section of the site
kage clone example.com --scope-prefix /docs

# Include subdomains, and trigger lazy-loaded images by scrolling
kage clone example.com --subdomains --scroll

# Re-render every page in place to pull in changed content
kage clone example.com --refresh
```

A clone is a polite breadth-first crawl. It honours `robots.txt`, seeds itself
from `sitemap.xml`, and scopes to the seed host unless you widen it. It is also
idempotent: each page is keyed by the file it writes, so the same URL reached
over http and https, with or without a trailing slash, is fetched once.
Re-running resumes where it left off; Ctrl-C saves state on the way out.
`--refresh` re-renders in place, `--force` wipes and starts clean.

Common flags:

| Flag | Default | Meaning |
|------|---------|---------|
| `-o, --out` | `$HOME/data/kage` | Output root; the mirror lands in `<out>/<host>/` |
| `-p, --max-pages` | `0` | Stop after N pages (0 = unlimited) |
| `-d, --max-depth` | `0` | Link-follow depth cap (0 = unlimited) |
| `--scope-prefix` | | Only crawl pages whose path starts with this prefix |
| `--subdomains` | `false` | Treat subdomains of the seed host as in scope |
| `--exclude` | | Path prefixes to skip (repeatable) |
| `--scroll` | `false` | Auto-scroll each page to trigger lazy loading |
| `--workers` | `4` | Concurrent page render workers |
| `--no-robots` | `false` | Ignore `robots.txt` |
| `-f, --force` | `false` | Delete any existing mirror for the host first |
| `--chrome` | | Path to the Chrome/Chromium binary |

Run `kage clone --help` for the full set, including the render-timing,
concurrency, and asset-size controls.

### Serve

`kage serve` runs a local static file server over a cloned folder so links and
assets resolve the way they would on a real host:

```bash
kage serve $HOME/data/kage/example.com
# open http://127.0.0.1:8800
```

## Pack it into one file

A clone is a folder, which is easy to browse but awkward to move: copying
thousands of small files is slow, and a directory is less tidy to hand over than
a single file. `kage pack` collapses a mirror into one artifact.

```bash
# An open ZIM archive, the single-file format Kiwix uses
kage pack example.com
kage open example.com.zim

# A single executable that *is* the site
kage pack example.com --format binary
./example
```

The default is a [ZIM](https://wiki.openzim.org/wiki/ZIM_file_format) archive:
the whole mirror in one file, text zstd-compressed and media stored as-is, that
`kage open` or any ZIM reader can browse. `--format binary` appends that archive
to a copy of kage and produces a single executable that serves the site offline
when run, so the recipient needs nothing installed, not even kage.

Packing is deterministic: the same mirror produces a byte-identical file, with
the archive UUID derived from the content rather than randomised, so a pack is
safe to checksum and cache. A bare host name resolves against the default output
directory, so `kage pack example.com` works right after `kage clone example.com`.

The appended archive is platform-independent; only the base executable carries
the architecture. Point `--base` at a kage built for another OS to make a viewer
for that platform from your own machine:

```bash
# From macOS, build a Windows viewer
kage pack example.com --format binary --base kage-windows-amd64.exe   # -> example.exe
```

## A native window, not a browser tab

By default a packed binary opens the system browser, which means a tab with an
address bar alongside your others. Build kage with the `webview` tag and it
opens the site in its own window instead, backed by the operating system's
WebView (WKWebView on macOS, WebView2 on Windows, WebKitGTK on Linux), so a
packed binary feels like a standalone app:

```bash
make build-webview                       # or: CGO_ENABLED=1 go build -tags webview ./cmd/kage
kage pack example.com --format binary --base bin/kage
./example                                # opens a window, no browser
```

This build needs cgo and links the platform WebView, so it stays opt-in. The
default build is pure Go (`CGO_ENABLED=0`) and the prebuilt release binaries
open the browser, which keeps the cross-compiled release pipeline simple.
`kage open` honours the same tag.

## How it works

```
seed URL ─▶ headless Chrome ─▶ final DOM ─▶ strip JS ─▶ localise assets ─▶ disk
              (render)          (snapshot)   (sanitize)   (rewrite links)
```

Pages are rendered by a pool of Chrome tabs; assets are fetched over plain HTTP
by a separate worker pool. Every URL maps deterministically to a local path, so
links can be rewritten before the asset they point at has finished downloading.
Output layout:

```
example.com/
├── index.html                 # the home page, scripts stripped
├── about/index.html           # /about
├── _kage/                      # reserved: assets and crawl state
│   ├── example.com/site.css    # localised stylesheet (url() rewritten)
│   ├── example.com/logo.png
│   └── state.json              # visited set, for resuming
└── ...
```

The same model underlies `pack`: the mirror's links are already mirror-relative
paths, and those map one-to-one onto the archive's content entries, so a click
in a served page hits the right entry with no rewriting.

## Building from source

```bash
git clone https://github.com/tamnd/kage
cd kage
make build          # -> bin/kage (pure Go, opens the browser)
make build-webview  # -> bin/kage with the native-window viewer (needs cgo)
make test           # full suite, including the Chrome-driven end-to-end tests
make test-short     # skip the tests that launch a browser
```

The repository is laid out by concern:

```
cmd/kage/   thin main: pins the main thread, then hands off to cli.Execute
cli/        the cobra command tree and flag wiring
clone/      the crawl: frontier, render workers, asset workers, resume state
browser/    headless Chrome control and DOM snapshotting
sanitize/   strip scripts, handlers, and javascript: URLs from the DOM
asset/      download and localise CSS, images, and fonts
urlx/       the deterministic URL-to-path mapping
zim/        a pure-Go ZIM reader and writer
pack/       mirror to ZIM or self-contained binary, and the offline HTTP handler
viewer/     present a served site: system browser, or native window (webview tag)
docs/       the tago documentation site
```

## Releasing

Push a version tag and GitHub Actions runs GoReleaser, which builds the
archives, the `.deb`/`.rpm`/`.apk` packages, a multi-arch GHCR image with
Chromium bundled, checksums, SBOMs, and a cosign signature:

```bash
git tag v0.1.0
git push --tags
```

The image tag carries no `v` prefix (`ghcr.io/tamnd/kage:0.1.0`). The Homebrew
and Scoop steps self-disable until their tokens exist, so the first release
works with no extra secrets.

## License

MIT. See [LICENSE](LICENSE).
