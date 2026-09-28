# Design — hather-macos-theme-manager

## Scope and design decisions

This design turns the validated explore, proposal, and theme-manager specification into a minimal Go architecture. It covers the first macOS release only. It does not implement code, modify HatDots, import HatDots files, or make HatDots a runtime dependency.

The supported platform is macOS 14 Sonoma and later. Phase 0 must verify the Apple Events and target-application behavior on the supported range before an adapter is marked production-ready. macOS 26.6.2 is the current verification host, not the only supported version.

The following decisions are explicit:

- **Peachy:** Phase 0 performs a bounded compatibility spike against the observed `generate --no-apply` and `export` commands. Peachy is not a runtime dependency, is not installed by Homebrew, and is not required by tests. The release implementation uses a native Go palette engine unless the spike proves that Peachy can be vendored, remain stable, and provide a materially better result without increasing installation risk. Peachy output may be used as a comparison fixture, not as the production contract.
- **Wallhaven credentials:** public use remains keyless. A key may come from `--wallhaven-api-key` or `WALLHAVEN_API_KEY`, with the explicit flag taking precedence. The key is never accepted from or written to plaintext Hather config, state, cache metadata, logs, JSON output, or error text. Keychain persistence is deferred until a separate security decision and implementation slice.
- **Ghostty ownership:** Hather owns `~/.config/ghostty/themes/hather.conf` and only a marked selection pair in the Ghostty config. A pre-existing `theme = ...` or inline color/palette block without a Hather ownership marker is a conflict; Hather does not replace it. Hather may append its marked selection only when no conflicting theme or inline palette is present. Existing Hather-owned selection is updated in place.
- **herdr reload:** Hather edits the configured TOML and does not signal, restart, unload, or reload a herdr process. The result is successful file application with `restart_required` follow-up until a bounded verification proves a safe reload command. No synchronous runtime change is implied.
- **Neovim apply:** MVP is write-only. Hather atomically writes `colors/hather.vim` and reports the next-start/manual `:colorscheme hather` action. It never assumes a server socket. A socket path and live `--remote-send` behavior require explicit opt-in and a later verified slice.
- **VS Code apply:** Hather merges only a small Hather-owned block in `settings.json`/JSONC under `workbench.colorCustomizations` and `editor.tokenColorCustomizations`. It does not set `workbench.colorTheme`, install extensions, or attempt to control a running VS Code process. The result recommends a window reload when necessary.
- **Adapter selection:** no adapter is silently applied on a fresh installation. The CLI requires an explicit adapter list or configured enabled list; the TUI requires the user to select adapters. Disabled adapters are not probed, written, or launched.
- **Appearance scheduling:** if automatic light/dark scheduling is detected, the default macOS operation refuses the explicit appearance change with an actionable conflict. An explicit future/CLI override may permit a permanent change only after the platform probe defines the exact state transition. Hather never silently disables scheduling.

## Data flow

```text
CLI or TUI request
    │
    ▼
internal/app request resolution
    │  explicit input > environment > config > defaults
    ▼
source: search/resolve → wallpaper metadata → fetch/cache → image artifact
    │
    ▼
palette engine: image bytes + engine version → deterministic Palette
    │
    ├── persist preview/last-applied metadata after successful generation
    ▼
adapter runner: fixed order, selected adapters only, isolated results
    │
    ├── macOS Apple Events
    ├── Ghostty owned theme + marked selection
    ├── herdr targeted TOML values
    ├── Neovim owned Vim colorscheme
    └── VS Code owned JSON/JSONC color entries
    │
    ▼
OperationResult → human output or stable JSON/exit code
```

The wallpaper and palette are shared prerequisites. If either fails, the runner does not execute adapters. Adapter execution is sequential and deterministic so file races and user-visible output are predictable; one adapter failure does not cancel later independent adapters unless the request context is cancelled.

## Package layout

The implementation should keep the public surface small and use `internal` packages:

```text
cmd/hather/main.go              process entry point and exit-code mapping
internal/model/                  source, artifact, palette, request, result types
internal/app/                    shared Preview/Apply orchestration
internal/config/                 config paths, precedence, redaction, state/cache records
internal/source/                 Source contract, local source, Wallhaven client/cache
internal/palette/                PaletteEngine contract and native deterministic engine
internal/adapter/                Adapter contract and runner
internal/adapter/macos/          System Events AppleScript adapter
internal/adapter/ghostty/        owned theme and marked config selection
internal/adapter/herdr/          targeted TOML theme editor and follow-up result
internal/adapter/neovim/         owned hather.vim generator
internal/adapter/vscode/         targeted JSON/JSONC color merge
internal/edit/                   atomic writes and format-specific safe edit helpers
internal/platform/               command execution abstraction and macOS checks
internal/cli/                    stdlib flag subcommands and text/JSON rendering
internal/tui/                    Bubble Tea model, views, and event-to-request mapping
```

`internal/tui` and `internal/cli` may depend on `internal/app` and `internal/model`, but never on adapter implementation packages. Adapter packages depend on model and narrow edit/platform helpers only. There is one composition root in `cmd/hather` that wires real sources, engine, runner, persistence, and platform executor. Tests inject fakes at those seams.

The initial dependency set is intentionally small: Go standard library for HTTP, image JPEG/PNG, filesystem, JSON, flags, and subprocesses; `github.com/BurntSushi/toml` for TOML validation/config decoding; a JSONC-capable parser such as `github.com/tailscale/hujson` for VS Code settings; `golang.org/x/image/webp` for supported WebP input; and Bubble Tea/Lip Gloss/Bubbles only in the TUI package. Phase 0 must confirm the selected JSONC API and versions before implementation. No CLI framework or Peachy dependency is planned.

## Core contracts

### Model types

`model.Wallpaper` contains source kind, stable source ID, title, page URL, image URL, attribution text, and optional source metadata. `model.Artifact` contains the local path, SHA-256, source identity, content type, and byte size. The path is local to the current operation and is never treated as a shell command.

`model.Palette` is versioned and adapter-neutral:

- `EngineVersion` and source artifact SHA-256;
- `Background`, `Foreground`, `Muted`, and `Accent` as canonical `#RRGGBB` values;
- deterministic ordered colors for adapters that need a 16-color terminal palette;
- diagnostics, including contrast adjustment or an explicit limitation.

The native engine works on decoded JPEG, PNG, and WebP images, downsamples deterministically to a bounded image, extracts colors with a fixed algorithm/order, and applies WCAG relative-luminance checks. Primary foreground/background contrast targets 4.5:1; muted text targets 3:1. Unsupported formats, including HEIC in the native MVP, return a classified error rather than a partial palette.

### Source and palette interfaces

The interfaces remain small and synchronous from the caller's perspective:

```go
type Source interface {
    Search(context.Context, SearchQuery) ([]Wallpaper, error)
    Fetch(context.Context, Wallpaper) (Artifact, error)
}

type PaletteEngine interface {
    Version() string
    Extract(context.Context, Artifact) (Palette, error)
}
```

The local source uses a `Wallpaper` carrying a local path and makes `Search` unnecessary for the local command; the application selects the local path directly and calls `Fetch`. The Wallhaven source owns query encoding, API response decoding, optional authorization, request throttling, cache lookup, download validation, and attribution metadata. `net/http` is used directly with bounded request and download sizes.

Source errors expose a stable class (`invalid_input`, `unavailable`, `rate_limited`, `authentication`, `download`, `decode`, or `cancelled`), a safe user message, and optional retry guidance. HTTP bodies and authorization values are not copied into normal diagnostics.

### Adapter and runner interfaces

```go
type Adapter interface {
    ID() string
    Apply(context.Context, Palette, ApplyContext) AdapterResult
}

type Runner interface {
    Apply(context.Context, Palette, ApplyRequest) OperationResult
}
```

`ApplyContext` contains only resolved Hather paths, target-specific options, and the platform/editor abstractions needed by the adapter. It does not expose the full config object or another adapter. `AdapterResult` contains adapter ID, status (`applied`, `skipped`, `unavailable`, `conflict`, or `failed`), stable error class when relevant, changed paths, generated artifact paths, follow-up instructions, and safe diagnostics. It may contain rollback metadata for Hather-owned values but never raw credentials or complete user configuration contents.

The runner order is fixed: `macos`, `ghostty`, `herdr`, `neovim`, `vscode`. It emits one result for each selected adapter, continues after isolated errors, and records successful sibling results. `OperationResult` contains the shared prerequisite status, palette identity, requested/selected adapters, ordered adapter results, attribution, and aggregate status (`complete`, `partial_failure`, `prerequisite_failure`, or `cancelled`).

CLI exit codes are stable: `0` for complete success or preview success, `2` for partial adapter failure, `3` for shared prerequisite failure, `4` for usage/configuration error, and `5` for cancellation/internal failure. JSON output uses the same model fields as text output so the TUI, CLI, and tests do not invent separate meanings.

## Configuration, state, and cache

Hather owns these paths and creates directories with user-only permissions where possible:

- `~/.config/hather/config.toml` — ordinary settings only;
- `~/.local/state/hather/last-applied.json` — last palette, source identity, engine version, ordered adapter results, attribution, and per-adapter ownership/rollback metadata;
- `~/.local/state/hather/runs/` — optional bounded operation records, retained only for the last 20 successful/failed operations;
- `~/Library/Caches/hather/` — downloaded wallpaper bytes and `index.json` metadata on macOS.

The cache key is the Wallhaven source ID plus a content extension. A valid cached artifact is reused; new downloads prune the oldest entries after the 20-wallpaper default. `hather cache clear` is a later CLI convenience, but removing the cache directory is always safe because state contains no required source bytes. Local images are not copied into the cache unless explicitly selected for a future data-retention feature.

Config values use explicit CLI input, then environment, then config file, then defaults. The config file may contain Wallhaven query defaults, selected adapter IDs, and non-secret path/options. It has no API-key field. Secret-bearing flags are excluded from persisted request/config/state structures and are redacted from errors.

State is written only after the relevant operation succeeds. Ownership metadata stores target path, prior Hather-owned scalar values/selection lines, and a pre-write file hash; it does not store full config files. A later rollback command may restore an owned value only when the current file still matches the post-apply hash. If another program changed the file, rollback refuses rather than overwriting it.

## Safe edit and rollback design

All file adapters use a shared `internal/edit` helper:

1. resolve and validate the target path without following an unexpected directory or symlink;
2. read bytes and preserve the existing mode when replacing an existing file;
3. parse/validate before mutation;
4. compute a narrow edit and verify that the intended ownership/target is unambiguous;
5. write a temporary file in the same directory with restrictive permissions, flush it, and atomically rename it;
6. return changed paths and prior owned values for state persistence.

No adapter rewrites a complete user config through an unrelated formatter when a targeted edit is possible. A malformed or ambiguous file is left byte-for-byte unchanged. There is no cross-adapter transaction: macOS side effects cannot be rolled back atomically with file edits, and a later adapter failure does not undo successful earlier adapters. Results say this plainly.

- **TOML:** validate the complete document with the TOML parser, then perform line-preserving edits only in `[theme]` and `[theme.custom]`. The initial herdr owned keys are `theme.name`, `panel_bg`, `accent`, `green`, `blue`, `red`, and `yellow`, subject to Phase 0 verification. Duplicate/ambiguous target keys, unsupported multiline forms, or malformed TOML cause a no-write conflict/failure. Missing theme sections are appended with a clear Hather-owned comment.
- **JSON/JSONC:** parse with the selected JSONC-capable parser, require an object root, update only the Hather-owned color keys beneath the two supported customization objects, and serialize using the parser's safe formatter. Invalid syntax, unsupported root shape, or ambiguous duplicate keys causes no write. Unrelated semantic values must survive fixture tests; formatting preservation is attempted but not promised where the parser normalizes whitespace.
- **Vim:** do not parse or edit user Lua/Vim configuration. Generate only the Hather-owned `colors/hather.vim` file with a stable header and palette assignments. An existing file without the Hather header is a conflict and is not overwritten.
- **Ghostty:** treat the config as a line-oriented key/value format. Write the Hather-owned theme file first only after target preflight. Add or update exactly the marked `theme = hather` selection pair when safe. Existing unmarked theme selection or inline palette is protected as a conflict; no broad replacement is attempted.

## Adapter behavior

### macOS

The adapter uses an injectable `CommandExecutor` to run `/usr/bin/osascript` without a shell. It sends the proven System Events script for all-desktop wallpaper application and light/dark appearance. Paths are escaped as AppleScript literals and are never interpolated into a shell command. The default wallpaper scope is all desktops; per-Space indexing is not exposed until Phase 0 verifies it.

The adapter classifies Apple Events denial, including the known `-1743` path, as `permission` and includes `System Settings → Privacy & Security → Automation` remediation. It runs an appearance-scheduling preflight before changing appearance. Automatic scheduling causes a safe conflict by default; the result identifies the required explicit override rather than silently changing the schedule. Tinted appearance and automatic propagation to every running application are outside this design.

### Ghostty

The generated artifact is `~/.config/ghostty/themes/hather.conf`, containing the semantic palette and terminal colors. The adapter discovers the config at `~/.config/ghostty/config` unless a future explicit path option is added. It will not create or rewrite a personalized config merely to force selection. A safe empty config can receive the marked selection; an existing inline palette or unmarked `theme` setting returns `conflict` with the exact remediation. The result describes Ghostty reload behavior as `applied; reload behavior depends on the installed version` until the Phase 0 version check proves the scope.

### herdr

The config path is `HERDR_CONFIG_PATH` when set, otherwise `~/.config/herdr/config.toml`. Only theme selection and verified custom color keys are owned. Shell, keybind, server, unknown, and unrelated values remain untouched. The adapter writes the file but never calls `herdr server`, LaunchAgent commands, or process signals. Every applied result includes `restart_required` and a manual follow-up. Phase 0 may narrow this to a verified reload command, but no implementation may assume one.

### Neovim

The default colors directory is `~/.config/nvim/colors`; path discovery is explicit and deterministic rather than inferred from plugin state. The adapter writes `hather.vim` with a small ANSI/editor palette, reports the path, and instructs the user to select it on next start or with `:colorscheme hather`. It does not edit LazyVim/plugin files or invoke `nvim`.

### VS Code

The default settings path is `~/Library/Application Support/Code/User/settings.json`. The adapter may report unavailable when the parent directory is absent. It merges a small fixed set of editor/workbench colors under Hather-owned entries and does not set an extension theme. JSONC comments and trailing commas are supported only if Phase 0 fixtures validate the selected parser; otherwise the adapter safely reports unsupported syntax rather than replacing the file. The result recommends a window reload and does not claim live application.

## CLI and TUI

The CLI uses standard-library `flag.FlagSet` subcommands: `search`, `preview`, `apply`, `version`, and `config` only as needed for the first slice. `preview` can use a local image without network. `apply` accepts a local path or Wallhaven selection, adapter IDs, optional JSON output, and the optional transient API key. It does not contain source, palette, or adapter logic.

The Bubble Tea model has only view state: query/results, selected wallpaper, palette preview, selected adapters, progress, and operation results. Commands/events call `internal/app`; rendering maps model results to human-readable controls. Long operations run through Bubble Tea commands and return the same `OperationResult` used by the CLI. Empty results, unreadable local paths, permission errors, and partial failure all have explicit views. No second pipeline is allowed in the TUI.

## Phase boundaries and review budget

Each slice is a separate review unit, with a hard target below 400 changed lines. The estimate includes tests and release files. If implementation or generated dependency/vendor changes put a slice above 400 lines, work stops under ask-on-risk; no exception or chain is inferred.

1. **Phase 0 — contract spikes and fixtures:** Wallhaven HTTP/error fixtures, Peachy compatibility notes, Apple Events executor fixtures, Ghostty/herdr reload and ownership probes, TOML/JSONC/Vim/Ghostty edit fixtures, and final dependency decisions. No production adapter breadth.
2. **Phase 1 — core, sources, palette, CLI foundation:** Go module, model/contracts, config/state/cache paths, Wallhaven/local source, native palette engine, runner with fake adapter, deterministic result/exit contracts, and offline CLI preview/apply seam. No real adapters or TUI.
3. **Phase 2 — macOS and Ghostty:** real Apple Events adapter, permission classification, owned Ghostty theme/selection editor, focused runner integration, and platform-gated tests. Keep TUI and later adapters out.
4. **Phase 3 — herdr, Neovim, and VS Code:** one narrow adapter plus fixtures at a time; preserve unrelated TOML/JSON values and owned-file conflicts; include follow-up semantics and partial-result tests.
5. **Phase 4 — TUI completion:** Bubble Tea selection, palette preview, adapter selection, apply progress, and per-adapter result screens over the existing app service. No new behavior in adapters or palette.
6. **Phase 5 — Homebrew and release hardening:** owned tap formula/release metadata, vendored build with no network during formula installation, help/version checks, offline local-image acceptance, documentation, and a bounded macOS rehearsal. Homebrew-core submission is explicitly outside the gate.

Every implementation step follows strict TDD: write the smallest failing fixture or behavior proof, implement the seam, run the focused package test, then run `go test ./...` before closing the slice. External verification that cannot run in CI is recorded as a bounded macOS acceptance check, never disguised as an offline unit test.

## Test strategy

Tests are table-driven and fixture-based, with no network or user-home mutation:

- Wallhaven tests use `httptest.Server` for valid search, no-key search, auth failure, rate limit/retry guidance, malformed response, and download failure. The cache is injected into a temporary directory.
- Palette tests use checked-in small PNG/JPEG/WebP fixtures, assert exact engine version, deterministic role/order output, and contrast diagnostics. The Peachy spike is an opt-in local verification, not a required `go test ./...` test.
- Core/runner tests use a fake source, fake palette engine, and fake adapters to prove prerequisite gating, fixed order, disabled adapters, sibling success retention, cancellation, and stable aggregate/exit results.
- Config tests prove precedence and secret redaction. State/cache tests prove path ownership, metadata persistence only after success, cache reuse, and bounded pruning.
- Edit tests compare fixture bytes or parsed semantic values before/after. They cover malformed/duplicate TOML, preserved herdr settings, Ghostty inline/theme conflicts, existing Hather markers, Vim ownership conflicts, JSON/JSONC comments/trailing commas, invalid roots, and VS Code unrelated settings.
- macOS adapter tests inject command output/error fixtures for success, `-1743`, scheduling conflict, malformed platform response, and safe path escaping. Actual Apple Events are a manual acceptance check on supported macOS versions.
- TUI tests exercise update commands and rendered state using fake app services; they do not launch Wallhaven, AppleScript, or target applications.
- Formula checks run in packaging CI plus a bounded install test for `--help`, `--version`, and local offline preview. The release test documents absent optional applications as independent unavailable results.

## Homebrew packaging

The first distribution path is an owned tap, separate from homebrew-core. The Hather release publishes a tagged source archive and checksum; the tap formula builds with a pinned Go dependency and `-mod=vendor` so installation does not fetch modules from the network. Version is injected with `-ldflags` and exposed through `hather version` and `--version`. The formula test invokes the installed binary and checks help/version output. Ghostty, herdr, Neovim, and VS Code are not formula dependencies; Hather must remain installable and useful with all of them absent.

Packaging is Phase 5 only. The formula and release workflow are not part of the core or adapter slices, and homebrew-core audit/submission is a later optional effort.

## Verification gates and unresolved evidence

Implementation may proceed only after Phase 0 records bounded evidence for: Wallhaven response/rate-limit details; Peachy export shape; Ghostty reload/selection behavior; verified herdr custom keys and reload semantics; JSONC parser behavior; System Events appearance scheduling properties; and supported macOS version behavior. A failed or unavailable probe narrows the adapter contract to the safe fallback described here. No external evidence is promoted to a guarantee merely because it appeared in exploration.

## Rollout and recovery

Roll out in phase order with each phase buildable and testable without later phases. Release notes must state that apply is not a cross-adapter transaction and must list per-adapter timing and limitations. Reverting a code slice does not remove user files. Hather-owned theme artifacts can be regenerated or removed; user configuration is never deleted as cleanup. Config restoration is permitted only for recorded Hather-owned values when the current target hash proves no intervening edit. Otherwise Hather reports a manual recovery path and leaves the file unchanged.
