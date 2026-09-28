# Explore — hather-macos-theme-manager

Phase: explore · Project: Hather · Backend: OpenSpec · Date: from local machine state (macOS 26.6.2)

## Goal

Establish the product and technical map for Hather's first change: an open-source macOS TUI + CLI that turns a wallpaper source (Wallhaven first) into a color palette and applies it through theme adapters (macOS wallpaper + appearance, Ghostty, herdr, Neovim, VS Code), distributed via Homebrew. No implementation here; package choices and adapter contracts are to be confirmed in proposal/spec/design.

## Product map (pipeline)

```
wallpaper source (Wallhaven API) → download image → palette engine (Peachy reuse?)
→ theme adapters (macOS, Ghostty, herdr, Neovim, VS Code) → user config files + live apply
```

- Reference implementation of the same idea already exists in the user's dotfiles: `HatDots/HatMac/bin/hat-theme` (bash: peachy palette → osascript wallpaper + appearance). Hather is its production/TUI/CLI successor; the adapter contracts can be derived from it.
- "Aether-like" is treated as directional ("unified cohesive theme across desktop + terminals + editors"), not as a specific product contract; no local evidence of a concrete Aether app exists (checked HatDots git logs/READMEs).
- MVP surface confirmed by `openspec/project.md`: TUI + CLI, Wallhaven, Homebrew-first, adapters listed above.

## Technical map — findings and evidence

### 1. Wallhaven API (external knowledge; must verify)

- Search endpoint: `GET https://wallhaven.cc/api/v1/search` with query params `q`, `categories` (3-bit `general|anime|people`, e.g. `010`), `purity` (`sfw|sketchy|nsfw`, e.g. `100`), `sorting` (`relevance|date_added|random|views|favorites|toplist`), `order`, `atleast`, `ratios`, `page`, `seed`, `colors`.
- Response: `{ "data": [ { id, url, short_url, path, thumb, file_type, file_size, colors[], resolution, ratio, category, purity, tags[], uploader{}, created_at, views, favorites, source } ], "meta": { current_page, last_page, per_page, total } }`. Image and thumbnail URLs host on `w.wallhaven.cc`.
- Auth: API key optional for search and download-limit boosts; obtained from user account settings (`https://wallhaven.cc/settings/account`). Sent as `apikey` query param on search and as `Authorization: Bearer <key>` on download requests.
- Rate limits: docs state search ~45 req/min (higher with key) and download limits per 30s window (higher with key). ToS requires attribution/respect for downloaded art.
- Verification plan: `curl 'https://wallhaven.cc/api/v1/search?q=nature&categories=010&purity=100&sorting=random&page=1'` before implementing; confirm param names, response shapes, and rate-limit numbers from `https://wallhaven.cc/help/api`.
- Go: stdlib `net/http` suffices; no SDK needed. Preview thumbnails are small JPEGs (x/image not even needed for thumbnails if only palette extraction is done on the full image).

### 2. Palette engine: Peachy reuse (local evidence — strong)

- Peachy = `github.com/bjarneo/peachy`, installed on this machine as a **Go binary** at `~/.local/bin/peachy` (mach-O, `__go_buildinfo`, links libSystem + Security.framework). Not in Homebrew (hat-theme comment: "no Homebrew formula is required"); likely `go install`.
- Contract used by `hat-theme`:
  1. `peachy generate <abs-wallpaper-path> --save <name> --no-apply`
  2. `peachy export <name> <out-dir>` → writes generated app theme files (Ghostty, Neovim, terminals, …) into `<out-dir>` (`~/.config/peachy/generated/hat-wallpaper` in hat-theme).
- Validation: supports jpg/jpeg/png/heic/heif/webp/bmp/tiff/gif; requires absolute POSIX path.
- Decision for proposal: (a) shell out to the `peachy` binary (exact upstream contract, but adds a Node-free Go runtime dependency and `--no-apply` semantics are peachy-specific), or (b) embed a minimal native Go palette engine (stdlib `image` + a small quantizer; HEIC needs cgo `libheif` — recommend jpg/png/webp via stdlib + golang.org/x/image, Wallhaven serves jpg/png almost exclusively). Option (b) avoids a cross-tool dependency but diverges from the "Peachy reuse" direction; propose with a spike.
- Evidence gaps: full set of app templates Peachy's `export` writes (filenames per adapter) — verify by running `peachy export` on this machine or reading bjarneo/peachy (README/source).

### 3. macOS wallpaper & Spaces (local evidence + known platform behavior)

- Reference implementation (hat-theme): `osascript -e 'tell application "System Events" to set picture of every desktop to POSIX file "<abs>"'` — sets wallpaper on **all desktops/spaces** (every display × every Space is enumerated as `desktop` objects).
- Per-Space wallpaper: macOS stores wallpaper per "desktop" (display + Space pair). Setting picture on a specific `desktop N` targets one Space; `every desktop` covers all. TUI/CLI should expose "all" (default) and optionally per-space via System Events index.
- Alternative APIs (knowledge, verify): `NSWorkspace.setDesktopImageURL` (AppKit) is **deprecated since macOS 14** and requires Automation permission from an external process; the SwiftUI `Wallpaper` framework (macOS 14+) sets wallpaper without extra permission but is Swift-only — a Go CLI cannot call it directly without a compiled helper. **Recommend osascript/System Events for MVP** (already proven on this machine, single dependency: macOS built-in).
- Spaces note: macOS does not rotate wallpapers per-Space automatically; per-Space setting only works while the target Space is the active/configured desktop. Test targeted-desktop behavior before promising it in spec.

### 4. macOS appearance (local evidence + known platform behavior)

- Reference implementation: `osascript -e 'tell application "System Events" to tell appearance preferences to set dark mode to true|false'` — proven on this machine.
- Machine runs **macOS 26.6.2 (Tahoe)**, which (like Sequoia 15.3+) includes a **Tinted** appearance option; keep the adapter to light/dark for MVP and note Tinted as future cursor (verify exact API surface for `tinted` via System Events on 26.x before spec).
- Knowledge: writing `defaults write -g AppleInterfaceStyle Dark` also works but is undocumented/unofficial; some apps ignore in-flight appearance changes. System Events path is the sanctioned, observable one.
- Limitation to record: appearance is system-wide, not per-Space or per-app (per-app overrides exist only inside apps); users running auto mode get overwritten by an explicit set — adapter should optionally restore `AppleInterfaceStyleSwitchesAutomatically` state.

### 5. Ghostty adapter (local evidence — strong)

- Live config: `/Users/facundogayoso/.config/ghostty/config` uses the **inline palette block** style (`background =`, `foreground =`, `cursor-color =`, `selection-background =`, `selection-foreground =`, 16× `palette = N=#hex`) — the user's Ghostty practice is inline, not `theme =` files.
- Theme file format confirmed locally: `/Users/facundogayoso/.config/ghostty/themes/catppuccin-mocha.conf` holds `palette = N=#hex` lines + base colors (no `#` on background/foreground; `#` inside palette values) — matches Ghostty theme file convention for `$XDG_CONFIG_HOME/ghostty/themes/`.
- Knowledge (verify): `theme = <name|path>` config key; builtin themes; `ghostty +list-themes` enumerates builtin + user-themes-dir themes; Ghostty watches the config file and hot-reloads most settings (verify exact reload scope per installed version — `ghostty --version`).
- Adapter contract options: (a) write `<palette-name>.conf` into `~/.config/ghostty/themes/` and set `theme = <palette-name>` in config (idempotent, replaces inline block cleanly), or (b) patch the inline color block in-place (matches user's current file, but fragile to formatting). Recommend (a); note the user's existing themes dir already proves the mechanism.

### 6. herdr adapter (local evidence — strong)

- herdr = Rust terminal-server/multiplexer (`brew "herdr"` — no custom tap line in `HatMac/Brewfile`, so it ships in homebrew-core); consumed as a pane runtime inside Ghostty; TUI chrome has a theme.
- Config: `~/.config/herdr/config.toml` (env override `HERDR_CONFIG_PATH`). Theme keys observed in `HatMac/herdr/config.toml`:
  - `[theme] name = "one-dark"` (18 builtin themes per herdr README: kanagawa, catppuccin, tokyo-night, dracula, nord, gruvbox, …)
  - `[theme.custom]` overrides: `panel_bg`, `accent`, `green`, `blue`, `red`, `yellow` (more keys likely exist).
- Discovery commands from README: `herdr theme list`. Reload: herdr server is LaunchAgent-managed (`~/Library/LaunchAgents/com.user.herdr.plist`, `install.sh` unloads/loads) — theme changes likely require `herdr server` restart or config reload; **verify hot-reload** on herdr.dev/docs/configuration/.
- Adapter contract: TOML edit of `[theme]` (+ optional `[theme.custom]` overrides) + restart guidance. herdr config being deleted resets everything (`rm -rf ~/.config/herdr ~/.local/state/herdr` per README) — Hather must never clobber the whole file, only the theme table, preserving `HERDR_CONFIG_PATH` when set.

### 7. Neovim adapter (local evidence)

- User nvim is **LazyVim** (`HatMac/nvim/lua/plugins/lazy.lua`: `install.colorscheme = {"tokyonight","habamax"}`; `lua/plugins/colorscheme.lua`: `vim.cmd.colorscheme("oxocarbon")`, `all-themes.lua` loads many palettes for hot-reload).
- No `--listen`/server socket found in the nvim config → live `--remote-send` apply is NOT currently available on this machine without user config change (evidence gap closed locally: not set up).
- Least-invasive adapter: write `~/.config/nvim/colors/hather.vim` (a colorscheme file; `colors/` is on nvim's runtimepath, no plugin manager needed) + apply via `:colorscheme hather` (manual or `nvim --headless -c 'colorscheme hather' +q` for the next start, or `--remote-send` when a server socket exists). Spec must choose between "write-only; user opt-in to apply" vs "write + headless apply".

### 8. VS Code adapter (local evidence: minimal)

- VS Code not found at `~/Library/Application Support/Code/User` on this machine → **the user may not run VS Code current**; keep the adapter as a config-file writer, not an assumption of a running instance.
- Contract (knowledge, verify): write to `settings.json` (`~/Library/Application Support/Code/User/settings.json`) — `workbench.colorTheme` only accepts **installed extension themes**, so the practical path is `workbench.colorCustomizations` + `editor.tokenColorCustomizations` (picked up on settings change, occasionally requiring window reload; verify). VS Code CLI (`code`) has no official "set theme" command.
- Risk: token-level customization is the most complex template of all adapters ("semantic token colors" mapping); keep the MVP template minimal (editor background/foreground + a few UI colors) and mark full token mapping as later work.

### 9. Go TUI/CLI/library choices (knowledge; verify versions at proposal)

- TUI: **Bubble Tea** (charmbracelet `bubbletea` + `lipgloss` styling + `bubbles` components) — de-facto standard, actively maintained; alternative `tview` (older tcell-style widgets).
- CLI: start with stdlib `flag` for MVP (CLI is thin: search/download/apply/palette commands); add `cobra` only if subcommand complexity demands it. TUI and CLI should share one `internal/app` core.
- Config parsing: **`github.com/BurntSushi/toml`** — TOML matches the ecosystem (herdr uses TOML, Ghostty uses TOML), avoids viper weight.
- HTTP/download: stdlib `net/http` (no client lib needed; respect Wallhaven rate limits with a small throttle).
- Image decode: stdlib `image/jpeg`, `image/png` + `golang.org/x/image/webp` if previews needed; **HEIC needs cgo** — avoid in MVP (Wallhaven serves jpg/png).
- Palette extraction in Go (if not shelling out to peachy): minimal median-cut/k-means over downscaled image (~100 lines) — decide in proposal with a tiny spike.

### 10. Configuration & storage (knowledge + ecosystem precedent; decide in proposal)

- Recommended (matches ecosystem + herdr/Ghostty precedent): `~/.config/hather/config.toml` (user config + enabled adapters + wallpaper source pins), `~/.local/state/hather/` (last-applied palette, cache index, download history), downloaded wallpapers under `~/.local/share/hather/wallpapers/` or `~/Library/Caches/hather`.
- Peachy output currently lands in `~/.config/peachy/generated/…` — Hather should own its own state dir; if reusing peachy, `--export` target points at a Hather dir.
- Wallhaven API key: Keychain via cgo (`zalando/go-keyring`, links Security.framework — note peachy also links Security) vs plaintext in config. Plaintext file is the lazy MVP; keychain is the correct end state. Flagging for spec decision; no security-sensitive feature may ship without an explicit choice here.

### 11. Permissions (macOS TCC)

- Wallpaper + appearance via `osascript` (System Events, Apple Events) requires **Automation permission** — first invocation triggers the TCC prompt attributed to the host app (the terminal running the TUI). No Accessibility or Screen Recording needed for these operations.
- Applies-once UX note: the TUI should detect the TCC denial (osascript error `-1743` "not allowed to send Apple events") and print a clear remediation path (System Settings → Privacy & Security → Automation).
- Reading/writing Ghostty/herdr/nvim/VS Code configs under `~/` from an unsandboxed CLI: no TCC required (Homebrew CLI is unsandboxed; sandbox is only a concern if a future .app).
- `tccutil reset AppleEvents` is the user's reset path. Notarization is not applicable to source-built Homebrew CLI binaries (only .app/.pkg distribution).

### 12. Homebrew packaging (knowledge; verify conventions at packaging phase)

- herdr precedent: a Rust terminal tool shipping as a plain formula (`brew "herdr"`) — Hather (Go) has an even simpler formula story.
- Path: **own tap first** (`brew tap <user>/homebrew-hather`, formula with `url` to GitHub release tarball + `go_install` or classic `go build`), then **homebrew-core PR** once stable (requires: semver tags, `brew audit --strict --new-formula` clean, test-bot CI, a `test do` block, livecheck).
- Go-specific formula notes: build must not hit network (vendored or module cache), versions from `go.mod`/tags, `ldflags -X` for version stamping. Ghostty stays a separate cask (undeclared dependency — Hather must degrade gracefully if the target apps are absent).
- Distribution constraint captured: **peachy has no formula** — if Hather shells out to peachy, ship the formula install instructions or vendor the engine (see §2).

## Constraints and risks

1. Automation/TCC denial blocks the entire macOS adapter path and is user-attributable; must fail loudly with remediation text.
2. Wallhaven rate limits + ToS: throttle downloads; keep API key optional; document attribution.
3. Adapter surface assumptions: user's live Ghostty config uses inline palette (not `theme=`), herdr default-shell/keybinds are personalized, nvim is LazyVim, VS Code may be absent — every adapter must be additive (never clobber unrelated config), which TOML/JSON section editing must honor.
4. Apply semantics differ per adapter: macOS (instant via osascript), Ghostty (hot-reload — verify scope), herdr (likely restart), nvim (next start unless server socket), VS Code (live-ish). Spec must state per-adapter apply latency instead of promising uniform "instant".
5. 400-line review budget ⇒ this change cannot ship all five adapters + TUI + CLI + packaging in one diff: propose explicit phase slices (e.g., core+source → macOS/Ghostty adapters → herdr/nvim/VS Code → packaging).

## Open decisions for proposal/spec/design

1. Palette engine: shell out to peachy vs native Go engine (spike).
2. Wallhaven key handling: plaintext config vs Keychain (cgo).
3. Ghostty apply mode: theme-file + `theme =` vs inline-block patch.
4. Neovim apply: write-only vs write + headless apply; whether to require an opt-in nvim server socket for live apply.
5. herdr: restart-based apply contract; whether to attempt live config reload.
6. CLI arg parsing: stdlib `flag` vs cobra.
7. Phase slice order and which adapters land in change #1 vs later changes.

## Evidence gaps (honest list with verification steps)

| # | Gap | How to close |
|---|-----|--------------|
| 1 | Wallhaven API exact params/response/rate limits differ from memory | `curl` the search endpoint; read `https://wallhaven.cc/help/api` |
| 2 | Ghostty config-reload scope & `+list-themes` on installed version | `ghostty --version`; `ghostty +list-themes`; ghostty.org/docs/config/reload |
| 3 | Full `[theme.custom]` key set and hot-reload for herdr | `herdr theme list`; herdr.dev/docs/configuration/ |
| 4 | Peachy `export` template file set per adapter | run `peachy export` locally; read bjarneo/peachy source |
| 5 | System Events targeted-desktop (per-Space) wallpaper behavior on macOS 26 | one-off osascript with a known desktop index on this machine |
| 6 | Tinted appearance availability via Apple Events on macOS 26 | inspect `appearance preferences` keys via osascript |
| 7 | VS Code settings live-apply + `workbench.colorTheme` requirement | install/check VS Code; test settings.json edit; VS Code docs |
| 8 | Homebrew formula/go_install conventions current state (Go version support) | `brew audit` at packaging phase; Homebrew formula cookbook |
| 9 | "Aether" product reference scope | product question for the owner; no local evidence found |
| 10 | Wallpaper per-Space storage model (display×space indexing) | AppleScript exploration + `NSScreen` enumeration notes |

## Sources referenced

- Local: `openspec/config.yaml`, `openspec/project.md` (Hather); `HatDots/HatMac/bin/hat-theme`, `HatDots/HatMac/herdr/{config.toml,README.md,install.sh,com.user.herdr.plist.template}`, `HatDots/HatMac/Brewfile`, `HatDots/HatMac/nvim/lua/{config/lazy.lua,plugins/colorscheme.lua,plugins/all-themes.lua}`, `~/.config/ghostty/{config,themes/catppuccin-mocha.conf}`, `~/.local/bin/peachy` (Go binary), `/System/Library/CoreServices/SystemVersion.plist` (macOS 26.6.2).
- External (verify): `https://wallhaven.cc/help/api`, `https://ghostty.org/docs/config`, `https://herdr.dev/docs/configuration/`, `https://github.com/bjarneo/peachy`, Apple AppKit/SwiftUI wallpaper + TCC docs, Homebrew formula cookbook, charmbracelet Bubble Tea docs.