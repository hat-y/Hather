# Tasks — hather-macos-theme-manager

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~900–1,300 across final-verification and second-final-verification corrections |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | Slices 1–7 (historical) → Slice 8: Wallhaven/CLI gaps → Slice 9: palette/config/macOS gaps → Slice 10: Ghostty ownership → Slice 11: evidence reconciliation/final verification |
| Delivery strategy | auto-chain |
| Chain strategy | stacked-to-main |

```text
Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: High
```

Historical completed slices remain recorded below. The final-verification corrections use the approved `auto-chain` / `stacked-to-main` strategy; each remains independently bounded below 400 changed lines.

## Guiding contract (applies to every slice)

- Strict TDD: RED (failing fixture/proof) → GREEN (smallest implementation) → TRIANGULATE (second probe/edge) → REFACTOR (collapsed, still green). Run `go test ./...` green before closing every slice.
- Each slice must build and pass offline with no network and no user-home mutation; tests use `httptest.Server`, checked-in fixtures, and injected temporary dirs.
- All real platform behavior (Wallhaven HTTP, Apple Events, Ghostty/herdr/nvim/VS Code, Homebrew) that cannot run in CI is recorded as a bounded macOS acceptance check in Phase 0/5, never disguised as an offline unit test.
- No slice depends on a later phase. No HatDots files at any point.

---

## Phase 0 — Contract spikes, fixtures, and dependency decisions

Slice: PR1. Dependency-ordered; produces fixtures and final decisions consumed by Phases 1–5. No production adapter breadth.

- [x] Write a Wallhaven response/error fixture set (`internal/source/testdata/wallhaven_*.json`): valid search, no-key search, rate-limit, auth-failure, malformed, empty-data, and a download (small JPEG). Record exact param names, `data`/`meta` shape, keyless path, and per-30s download limit from the live endpoint into a `docs/wallhaven.md` note. <!-- sdd-owner: implementation -->
- [x] Run a bounded Peachy compatibility probe (`peachy generate --no-apply` then `peachy export`) on this machine and record observed export file shapes in `docs/peachy-spike.md`; keep it as comparison material, not a production dependency. <!-- sdd-owner: implementation -->
- [x] Add checked-in tiny palette fixtures (`internal/palette/testdata/fixture.{jpg,png,webp}`) plus a reference expected-palette JSON, so palette determinism and contrast tests have stable inputs. <!-- sdd-owner: implementation -->
- [x] Add Apple Events `CommandExecutor` output/error fixtures (`internal/platform/testdata/exec_*.json`): success, `-1743` Automation denial, appearance-scheduling error, malformed output, and a path-escaping vector. <!-- sdd-owner: implementation -->
- [x] Add Ghostty and herdr ownership/reload fixtures: a live `themes/hather.conf` sample, an inline-palette config sample, a `theme=` config sample, and a herdr `config.toml` with personalized shell/keybind/server values. Verify and record Ghostty reload scope and herdr managed `[theme]`/`[theme.custom]` keys in `docs/platform-verification.md`. <!-- sdd-owner: implementation -->
- [x] Add JSON/JSONC edit fixtures for VS Code (`settings.json` with comments and trailing commas) and a JSONC parser decision recorded for `internal/edit` (design names `github.com/tailscale/hujson`; confirm API/version). <!-- sdd-owner: implementation -->
- [x] Record final dependency decisions and pinned versions in `go.mod` scaffold notes: stdlib HTTP/JPEG/PNG, `github.com/BurntSushi/toml`, JSONC parser, `golang.org/x/image/webp`, and TUI-only Bubble Tea/Lip Gloss/Bubbles. No Peachy, no CLI framework. <!-- sdd-owner: implementation -->

Acceptance: fixtures committed; `docs/{wallhaven,peachy-spike,platform-verification}.md` record probes; dependency decisions explicit. Estimate ~280–380 changed lines.

---

## Phase 1 — Core, sources, native palette, and CLI foundation

### Phase 1a — Model, config/state/cache, app service (PR2)

- [x] RED: write `internal/model/*_test.go` asserting the `Wallpaper`, `Artifact`, `Palette`, `AdapterResult`, and `OperationResult` types with their stable fields (palette EngineVersion + 4 semantic roles + ordered 16 colors + diagnostics; result statuses `applied|skipped|unavailable|conflict|failed`; aggregate `complete|partial_failure|prerequisite_failure|cancelled`). <!-- sdd-owner: implementation -->
- [x] GREEN+REFACTOR: define `internal/model/model.go` satisfying the tests. <!-- sdd-owner: implementation -->
- [x] RED: write `internal/config/config_test.go` proving explicit-input > env > config-file > default precedence and that a `--wallhaven-api-key`/`WALLHAVEN_API_KEY` secret never enters config/state/request structures and is redacted from diagnostics. <!-- sdd-owner: implementation -->
- [x] GREEN+REFACTOR: implement `internal/config/config.go` (paths: `~/.config/hather/config.toml`, `~/.local/state/hather/last-applied.json`, `runs/` bounded-20 retention, `~/Library/Caches/hather/` + `index.json`; user-only dir perms). <!-- sdd-owner: implementation -->
- [x] RED: write `internal/config/state_test.go` proving last-applied metadata is persisted only after success and cache keying/pruning (20-wallpaper default) behaves correctly with a temp dir. <!-- sdd-owner: implementation -->
- [x] GREEN+REFACTOR: implement state/cache persistence (`internal/config/state.go`). <!-- sdd-owner: implementation -->
- [x] RED: write `internal/app/app_test.go` proving shared-prerequisite gating (bad artifact/palette → no adapters run), fixed runner order, disabled adapters untouched, sibling-success retention on adapter failure, and stable aggregate status. Uses fake source, fake engine, fake adapters. <!-- sdd-owner: implementation -->
- [x] GREEN: implement `internal/app/app.go` orchestrating source→palette→runner with injectable seams; persist last-applied record via `internal/config`. <!-- sdd-owner: implementation -->

Acceptance: model/config/state/app tested; CLI/TUI both route through this one app core (no second pipeline). Estimate ~340–390 changed lines. If this exceeds 400, stop for ask-on-risk and split state persistence into its own slice.

### Phase 1b — Wallhaven + local sources, native palette engine, CLI + main (PR3)

- [x] RED: write `internal/source/wallhaven_test.go` using `httptest.Server` for valid search, no-key search, optional auth header, rate-limit classification with retry guidance, auth failure, malformed/empty response, and download failure; assert stable error classes (`invalid_input|unavailable|rate_limited|authentication|download|decode|cancelled`). <!-- sdd-owner: implementation -->
- [x] GREEN+REFACTOR: implement `internal/source/wallhaven.go` (query encoding, response decode, optional key from flag/env — never from config, request throttle, `internal/config` cache lookup/reuse, download validation, attribution metadata) and `internal/source/local.go` selecting a local path directly with `Fetch` and no `Search`. Expose the `Source` interface. <!-- sdd-owner: implementation -->
- [x] RED: write `internal/palette/palette_test.go` asserting version-pinned determinism across repeated extraction and across JPEG/PNG/WebP fixtures, exact role/order output, 4.5:1 foreground and 3:1 muted contrast, and a "contrast limitation" diagnostic for a low-contrast fixture. <!-- sdd-owner: implementation -->
- [x] GREEN+REFACTOR: implement `internal/palette/engine.go` (stdlib image decode for JPEG/PNG, `golang.org/x/image/webp` for WebP, deterministic downsample + fixed-order quantizer, WCAG relative luminance), returning classified errors for unsupported formats (e.g. HEIC). <!-- sdd-owner: implementation -->
- [x] RED: write `internal/cli/cli_test.go` proving stdlib `flag.FlagSet` subcommands (`search`, `preview`, `apply`, `version`) share the app core, `preview` works offline on a local image, and stable exit codes: `0` complete/preview, `2` partial_failure, `3` prerequisite_failure, `4` usage/config error, `5` cancelled/internal. JSON output uses the same model fields as text. <!-- sdd-owner: implementation -->
- [x] GREEN+REFACTOR: implement `internal/cli/*.go` (flag parsing, rendering) and `cmd/hather/main.go` composition root wiring real source/engine/persistence/platform executor and the exit-code mapping. <!-- sdd-owner: implementation -->

Acceptance: offline CLI `preview`/`apply` seam works with fixtures; `go test ./...` green; no real adapters yet (fake adapter in runner tests only). Estimate ~350–400 changed lines; if over, split the palette engine or CLI out into Phase 1c under ask-on-risk.

---

## Phase 2 — macOS and Ghostty adapters + runner integration

### Phase 2a — macOS System Events adapter (PR4)

- [x] RED: write `internal/adapter/macos/macos_test.go` injecting `CommandExecutor` fixtures for all-desktop wallpaper success, `-1743` Automation denial (classified `permission` with `System Settings → Privacy & Security → Automation` remediation), appearance-scheduling detected → safe `conflict` (never silently disables scheduling), malformed platform output, and safe AppleScript path escaping (no shell interpolation). <!-- sdd-owner: implementation -->
- [x] GREEN+REFACTOR: implement `internal/adapter/macos/macos.go` running `/usr/bin/osascript` via the injected executor for all-desktop wallpaper + light/dark appearance; appearance-scheduling preflight returns explicit conflict/override instead of mutating schedule. Per-Space and Tinted are not exposed. <!-- sdd-owner: implementation -->
- [x] Add `internal/platform/exec.go` command-execution abstraction (no shell) with `CommandExecutor` interface used by macOS. <!-- sdd-owner: implementation -->

Acceptance: permission denial is actionable; scheduling state preserved; actual Apple Events is a manual Phase 5 acceptance check, not a unit test. Estimate ~300–360 changed lines.

### Phase 2b — Ghostty owned theme + marked selection; runner integration (PR5)

- [x] RED: write `internal/adapter/ghostty/ghostty_test.go` proving it writes/updates only `~/.config/ghostty/themes/hather.conf` and only the marked `theme = hather` selection pair; an existing unmarked `theme=` or inline palette returns `conflict` without clobbering; an absent Ghostty config dir returns `skipped`/`unavailable` and is non-fatal; unrelated config lines are preserved. <!-- sdd-owner: implementation -->
- [x] GREEN+REFACTOR: implement `internal/adapter/ghostty/ghostty.go` (Hather-owned theme artifact writer + marked-selection editor via `internal/edit`) and back it with the shared line-oriented edit helper. <!-- sdd-owner: implementation -->
- [x] RED: extend `internal/adapter/runner_test.go` (or new `internal/adapter/runner.go` tests) proving fixed order `macos,ghostty,herdr,neovim,vscode`, isolated per-adapter results, partial-failure aggregation, and that one failed adapter leaves sibling artifacts/results intact. <!-- sdd-owner: implementation -->
- [x] GREEN+REFACTOR: implement `internal/adapter/adapter.go` (Adapter contract, `ApplyContext` narrow scope, no cross-adapter calls) and `internal/adapter/runner.go` wired with macOS + Ghostty adapters only in this slice. <!-- sdd-owner: implementation -->
- [x] RED: write `internal/edit/edit_test.go` proving atomic temp-write + rename, mode preservation, no-symlink-follow, prior owned-value capture, and byte-for-byte no-op on malformed/ambiguous input. <!-- sdd-owner: implementation -->
- [x] GREEN+REFACTOR: implement `internal/edit/edit.go` shared by all file adapters (used here by Ghostty; reused by Phases 3a–c). <!-- sdd-owner: implementation -->

Acceptance: Ghostty additive + conflict-safe; runner isolates macOS/Ghostty; no TUI, no herdr/nvim/VS Code yet. Estimate ~330–390 changed lines. If over 400, split runner+edit into its own slice under ask-on-risk.

---

## Phase 3 — herdr, Neovim, and VS Code adapters

### Phase 3a — herdr targeted TOML editor (PR6)

- [x] RED: write `internal/adapter/herdr/herdr_test.go` proving only `[theme]`/`[theme.custom]` owned keys change (`theme.name`, `panel_bg`, `accent`, `green`, `blue`, `red`, `yellow` per Phase 0 verification) while shell/keybind/server/unknown values survive byte-level in semantic terms; `HERDR_CONFIG_PATH` honored when set; malformed/duplicate/ambiguous TOML → no write + remediation; every applied result includes `restart_required` future action without implying a running herdr changed. <!-- sdd-owner: implementation -->
- [x] GREEN+REFACTOR: implement `internal/adapter/herdr/herdr.go` using `internal/edit` line-preserving TOML edits; never invokes `herdr server`, LaunchAgent commands, or process signals. Missing theme sections appended with a Hather-owned comment. <!-- sdd-owner: implementation -->

Acceptance: herdr theme-only, restart-explicit, malformed-input safe. Estimate ~280–340 changed lines.

### Phase 3b — Neovim additive colorscheme (PR7)

- [x] RED: write `internal/adapter/neovim/neovim_test.go` proving atomic write of only `~/.config/nvim/colors/hather.vim` with a stable Hather header + ANSI/editor palette; an existing colorscheme file without the Hather header → `conflict` not overwritten; existing plugin/colorscheme Lua config untouched; result reports next-start/manual `:colorscheme hather` and never attempts a server socket. <!-- sdd-owner: implementation -->
- [x] GREEN+REFACTOR: implement `internal/adapter/neovim/neovim.go` (generator only via `internal/edit`; writes header-prefixed `hather.vim`). Explicit deterministic colors-dir path, no LazyVim/plugin edits. <!-- sdd-owner: implementation -->

Acceptance: write-only, non-destructive, server-socket-free. Estimate ~260–320 changed lines.

### Phase 3c — VS Code JSON/JSONC settings merge (PR8)

- [x] RED: write `internal/adapter/vscode/vscode_test.go` proving the parser (hujson) handles comments/trailing commas, merges only Hather-owned color keys beneath `workbench.colorCustomizations` and `editor.tokenColorCustomizations`, preserves unrelated settings, reports `unavailable` when the settings parent dir is absent, returns actionable failure with no full-file replacement on invalid/ambiguous root, and recommends window reload without claiming live apply. No `workbench.colorTheme`, extension install, or process control. <!-- sdd-owner: implementation -->
- [x] GREEN+REFACTOR: implement `internal/adapter/vscode/vscode.go` using `internal/edit` + the confirmed JSONC parser; safe no-write on unsupported syntax. <!-- sdd-owner: implementation -->

Acceptance: additive merge, unrelated values preserved, invalid input not clobbered. Estimate ~280–340 changed lines.

---

## Phase 4 — Bubble Tea TUI (PR9)

- [x] RED: write `internal/tui/model_test.go` proving update commands/events route to fake app services and return the same `OperationResult` the CLI uses; render maps results to views for query/results, selected wallpaper, palette preview, selected-adapter toggle, apply progress, and per-adapter results; explicit empty/error/permission/partial-failure views; pressing apply requires a user-selected adapter set (nothing applied by default). <!-- sdd-owner: implementation -->
- [x] GREEN+REFACTOR: implement `internal/tui/*` (Bubble Tea model minimizing view state; Lip Gloss/Bubbles styling) over `internal/app` only; no source/palette/adapter logic added; no second pipeline. <!-- sdd-owner: implementation -->
- [x] RED+REFACTOR: add TUI test coverage for unreadable local path and empty Wallhaven-results recovery guidance, proving no blank/destructive state. <!-- sdd-owner: implementation -->

Acceptance: TUI and CLI share the one app core and result model; offline/TUI tests use fake services, never launching Wallhaven/AppleScript/apps. Estimate ~330–390 changed lines. If over 400 with styling stdlib, split result-display into its own slice under ask-on-risk.

---

## Phase 5 — Owned-tap Homebrew packaging and release hardening (PR10)

- [x] RED: write a formula-level test manifest (`test do` block invoking the installed binary) in the owned tap repo asserting `--help` and version output work and the executable is present on PATH; draft as a failing check before packaging wiring. <!-- sdd-owner: implementation -->
- [x] GREEN+REFACTOR: add release packaging metadata (tagged source archive + checksums, recipe formula/release workflow in the owned `homebrew-hather` tap), pinned Go dep with `-mod=vendor` so formula install fetches no modules from the network, `-ldflags` version injection surfaced by `hather version`/`--version`. Ghostty/herdr/Nvim/VS Code are not formula dependencies. <!-- sdd-owner: implementation -->
- [x] Add documentation (`docs/` + release notes): supported macOS scope (14+), optional Wallhaven credentials (flag/env only, never config), attribution, Automation permission remediation, per-adapter apply timing (instant/hot-reload vs restart/next-start/manual), missing-target behavior, apply-is-not-a-transaction, and known non-goals; homebrew-core is explicitly outside the gate. <!-- sdd-owner: implementation -->
- [x] Run a bounded macOS acceptance rehearsal on this machine (scripted, recorded in the release record): fresh `brew install`, `--help`/version, offline local-image `preview`/`apply` with selected adapters, macOS Automation denial path, absent-target adapters returning independent `unavailable` results, and `go test ./...` green from the vendored build. <!-- sdd-owner: implementation -->

Acceptance: owned tap installs `hather`; offline behavior proven; release record lists any adapter requiring a local app/permission. Estimate ~300–360 changed lines (excluding generated release archive bytes).

---

## Final-verification corrections

No code, commit, publication, or platform mutation is authorized by this planning update. Implement each slice independently under 400 changed lines using the approved `auto-chain` / `stacked-to-main` strategy.

- [x] Slice 1 (target ≤250 lines) — RED → GREEN → TRIANGULATE → REFACTOR in `internal/config/{config.go,config_test.go}`, `internal/app/{app.go,app_test.go}`, and `cmd/hather/{main.go,main_test.go}`: prove ordinary runtime config (enabled adapters and supported target options) is loaded and resolved with CLI > environment > config > defaults, reaches the composed app/runner, and cannot persist or expose the Wallhaven key; then run focused tests and `go test ./...`. <!-- sdd-owner: implementation -->
- [x] Slice 2 (target ≤320 lines) — RED → GREEN → TRIANGULATE → REFACTOR in `internal/cli/{cli.go,cli_test.go}` and `internal/tui/{model.go,model_test.go}`: prove CLI and TUI select and render the same Wallhaven result labels/metadata, preview the same selected local or Wallhaven wallpaper, and gate apply until a successful preview plus an explicit adapter selection; keep all paths on `internal/app`. <!-- sdd-owner: implementation -->
- [x] Slice 3 (target ≤180 lines) — RED → GREEN → TRIANGULATE → REFACTOR in `internal/adapter/herdr/{herdr.go,herdr_test.go,testdata/config.toml}`: replace unverified `blue`/`yellow` ownership with the Phase 0 observed `[theme.custom]` keys from `docs/platform-verification.md` (`sidebar_bg`, `active_row_bg`, `selection_bg`, `panel_bg`, `accent`, `red`, `green`), preserving all non-owned TOML bytes/semantics and restart-only guidance. <!-- sdd-owner: implementation -->
- [x] Slice 4 (target ≤320 lines) — RED → GREEN → TRIANGULATE → REFACTOR in `internal/edit/{edit.go,edit_test.go}`, `internal/adapter/ghostty/{ghostty.go,ghostty_test.go}`, and `internal/adapter/herdr/{herdr.go,herdr_test.go}`: reject every symlinked ancestor or lexical escape before either adapter reads, writes, or creates an owned artifact; prove external targets and config bytes remain unchanged for `HERDR_CONFIG_PATH` and Ghostty config/theme roots. <!-- sdd-owner: implementation -->
- [x] Slice 5 (target ≤360 lines) — RED → GREEN → TRIANGULATE → REFACTOR across `go.mod`, `.github/workflows/release.yml`, and the `facundogayoso/homebrew-hather/Formula/hather.rb.in` discovery target: align the Go toolchain used by CI/release/formula, retain vendored offline builds, and add formula/installed-binary acceptance proving a local-image preview/apply remains usable while every selected missing target yields its own actionable unavailable result. <!-- sdd-owner: implementation -->
- [x] Slice 6 (target ≤120 lines) — reconcile executable evidence and unchecked work across `openspec/changes/hather-macos-theme-manager/{tasks.md,apply-progress.md,verify-report.md}`: correct completed-task claims, verifier counts, evidence revisions, and residual-risk wording so no task or report claims behavior without the corresponding focused proof. <!-- sdd-owner: implementation -->
- [x] Slice 7 (target ≤120 lines) — run and record final focused regression checks plus `go test ./... -count=1 -timeout 90s`, `go build ./...`, `go vet ./...`, and `git diff --check`; perform the bounded macOS/Homebrew acceptance from `docs/` and update `openspec/changes/hather-macos-theme-manager/verify-report.md` with the observed verdict, failures, and remaining manual limitations. <!-- sdd-owner: implementation -->

 ## Second-final-verification corrections

 These bounded slices address the seven newly observed gaps; delivery remains `auto-chain` / `stacked-to-main`. Keep each slice below 400 changed lines, preserve this history, and leave Slice 7 above unchecked until its own work is completed.

 - [x] Slice 8 (target ≤280 lines) — RED → GREEN → TRIANGULATE → REFACTOR in `internal/source/{wallhaven.go,wallhaven_test.go}` and `internal/cli/{cli.go,cli_test.go}`: prove HTTP 429 is classified as `rate_limited` with retry guidance, empty search results produce recovery guidance, selected result attribution retains its page URL, and CLI search/apply expose the same source metadata; run focused tests and `go test ./...`. <!-- sdd-owner: implementation -->
 - [x] Slice 9 (target ≤320 lines) — RED → GREEN → TRIANGULATE → REFACTOR in `internal/cli/{cli.go,cli_test.go}`, `internal/config/{config.go,config_test.go}`, `cmd/hather/{main.go,main_test.go}`, and `internal/adapter/macos/{macos.go,macos_test.go}`: prove CLI palette output has semantic-role parity with the shared palette, config provenance is reported and configured queries are consumed, and macOS results state the verified propagation limitation without implying all Spaces/apps update instantly; run focused tests and `go test ./...`. <!-- sdd-owner: implementation -->
 - [x] Slice 10 (target ≤180 lines) — RED → GREEN → TRIANGULATE → REFACTOR in `internal/adapter/ghostty/{ghostty.go,ghostty_test.go}`: prove an existing unmarked Ghostty theme selection is user-owned, returns conflict, and causes no write to either config or Hather theme artifact; preserve marked Hather selection behavior and run focused tests plus `go test ./...`. <!-- sdd-owner: implementation -->
 - [x] Slice 11 (target ≤240 lines) — reconcile only observed evidence and residual claims in `openspec/changes/hather-macos-theme-manager/{tasks.md,apply-progress.md,verify-report.md}`, then run focused checks and final `go test ./... -count=1 -timeout 90s`, `go build ./...`, `go vet ./...`, and `git diff --check`; update the verify report with actual outcomes and remaining platform limitations, without checking Slice 7 unless that separately scoped work is completed. <!-- sdd-owner: implementation -->