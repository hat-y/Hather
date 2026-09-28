# Apply Progress — hather-macos-theme-manager (PR1 / Phase 0)

## Status consumed

Native SDD status pointed at `/Users/facundogayoso/projects/HatDots` (stale; no active change).
Parent granted this run for `/Users/facundogayoso/projects/Hather`, attempt token
`sha256:fedd490e9c87dffd00580199e310dbb5252ecd7b9eadec140bfa8494903e97ce`. HatDots untouched.
Target actionContext: `repo-local`, allowed edit roots = Hather repo; no warnings.

## Delivery gate

`Decision needed before apply: No`; chained PRs `Yes` (stacked-to-main); budget risk `High`.
PR1 = Phase 0 only; fixtures/docs, no production code.

## Completed tasks (all Phase 0; tasks.md boxes checked)

Wallhaven fixtures + `docs/wallhaven.md` · Peachy probe + `docs/peachy-spike.md` · palette
fixtures + `expected-palette.json` · Apple Events exec fixtures + path-escaping note ·
Ghostty/herdr fixtures + `docs/platform-verification.md` · VS Code JSONC fixture + hujson
decision · dependency decisions + pinned versions (`docs/dependencies.md`).

## Files changed

- `internal/source/testdata/` — 6 JSON fixtures + live JPEG download (binary).
- `internal/palette/testdata/` — `fixture.{png,jpg,webp}` + `expected-palette.json`.
- `internal/platform/testdata/` — 5 `exec_*.json`.
- `internal/adapter/{ghostty,herdr,vscode}/testdata/` — Ghostty theme + 3 config samples, herdr
  `config.toml`, VS Code `settings.json`.
- `docs/` — `wallhaven.md`, `peachy-spike.md`, `platform-verification.md`, `dependencies.md`.
- `tasks.md` — 7 Phase 0 boxes checked.

## Checks / probes run

Wallhaven: live search/keyless/empty (200), invalid apikey (401), rate headers (45), thumb
download (200 `image/jpeg`). Peachy: generate+export in temp HOME (17 files). Ghostty 1.3.1
(`--version`, `+list-themes`); herdr 0.8.2 (`--default-config`, `config check` ok,
`server reload-config --help`); osascript localized-error probe; appearance-key `defaults read`.
Fixture sanity: `file` types + `json.load` on JSON fixtures.
`go test ./...` — **not run**: no Go toolchain and no module/test surface by design. Not green.

## TDD evidence

Strict TDD active, but all Phase 0 tasks are fixture/evidence artifacts with no production seam
to RED against; probe outputs above are the evidence. Phase 1 starts RED→GREEN once the
toolchain + module exist.

## Deviations from design

- herdr owned keys: observed 0.8.2 set is `sidebar_bg,active_row_bg,selection_bg,panel_bg,
  accent,red,green` (no blue/yellow — design assumed them).
- `herdr theme list` missing on 0.8.2; use `--default-config` / `config check` /
  `server reload-config`.
- Wallhaven "per-30s download limit" not on current help page; only 45 API calls/min.
- `expected-palette.json` holds unpinned shape placeholders; Phase 1 pins golden values.

## Remaining tasks

Phase 0 done; Phases 1–5 (PR2–PR10) remain per `tasks.md`.

## Workload / PR boundary

PR1 = Phase 0 only. Reviewable slice (fixtures + docs) = 329 text lines + 4 binary
fixtures, under the 380-line budget; no `size:exception` needed. SDD bookkeeping
(`apply-progress.md` + `tasks.md` checkbox edits) is process output, not PR review content.

---

# Apply Progress — PR2 / Phase 1a resumed

## Status consumed

Native status: `hather-macos-theme-manager`, `applyState: ready`, OpenSpec authoritative;
`actionContext.mode: repo-local`, workspace/allowed root `/Users/facundogayoso/projects/Hather`,
no warnings. Parent supplied runtime token `sha256:f9bae6a728286019cbeed06314f6c664f7513d715781650b002e546882bd3b3b`;
no acquire/reset/settle action was performed.

## Completed tasks and persisted checkboxes

All eight Phase 1a implementation rows are visibly `- [x]` in `tasks.md`. This resumed run
completed the state RED/GREEN pair and the app RED/GREEN pair immediately after their evidence:
state/cache persistence, bounded run retention, and the shared source → palette → runner service.
The pre-existing config proof now covers both flag and environment secret values without adding a
persisted credential field.

## Files changed

- `internal/config/state.go` — user-only JSON state/cache/run persistence, 20-entry pruning.
- `internal/config/{state_test.go,config_test.go}` — temp-dir persistence/retention and secret proof.
- `internal/app/{app.go,app_test.go}` — injectable shared orchestration and fake-seam proof.
- `openspec/changes/hather-macos-theme-manager/{tasks.md,apply-progress.md}` — completion/progress.

## TDD Cycle Evidence

| Cycle | RED observed | GREEN / triangulation observed | Refactor |
| --- | --- | --- | --- |
| state/cache | `go test ./internal/config` failed on missing `NewStore`, limits, and persistence APIs | state, cache pruning, and 20-run retention pass in isolated temp dirs | stdlib JSON/filesystem only; shared JSON writer retained |
| app orchestration | `go test ./internal/app` failed on absent app contracts; expanded bad-artifact case then failed because runner ran | prerequisite errors prevent runner calls; selected IDs arrive in canonical order; partial sibling result persists; `go test ./...` passes | kept source/engine/runner seams synchronous and minimal |

## Verification

- `go test ./internal/config` — pass (after RED and GREEN).
- `go test ./internal/app` — pass (after RED and bad-artifact triangulation).
- `go test ./...` — pass.
- `gofmt` applied to changed Go files; `git diff --check` — pass.

## Deviations and risks

No design deviation. The app intentionally supplies an ordered selected-adapter request but does
not implement adapters or their runner; Phase 2b owns actual adapter execution. State persistence
is limited to complete/partial-success results, so shared-prerequisite and cancelled operations do
not create `last-applied.json`.

## Remaining tasks

- [ ] RED: write `internal/source/wallhaven_test.go` using `httptest.Server` for valid search, no-key search, optional auth header, rate-limit classification with retry guidance, auth failure, malformed/empty response, and download failure; assert stable error classes (`invalid_input|unavailable|rate_limited|authentication|download|decode|cancelled`). <!-- sdd-owner: implementation -->
- [ ] GREEN+REFACTOR: implement `internal/source/wallhaven.go` (query encoding, response decode, optional key from flag/env — never from config, request throttle, `internal/config` cache lookup/reuse, download validation, attribution metadata) and `internal/source/local.go` selecting a local path directly with `Fetch` and no `Search`. Expose the `Source` interface. <!-- sdd-owner: implementation -->
- [ ] RED: write `internal/palette/palette_test.go` asserting version-pinned determinism across repeated extraction and across JPEG/PNG/WebP fixtures, exact role/order output, 4.5:1 foreground and 3:1 muted contrast, and a "contrast limitation" diagnostic for a low-contrast fixture. <!-- sdd-owner: implementation -->
- [ ] GREEN+REFACTOR: implement `internal/palette/engine.go` (stdlib image decode for JPEG/PNG, `golang.org/x/image/webp` for WebP, deterministic downsample + fixed-order quantizer, WCAG relative luminance), returning classified errors for unsupported formats (e.g. HEIC). <!-- sdd-owner: implementation -->
- [ ] RED: write `internal/cli/cli_test.go` proving stdlib `flag.FlagSet` subcommands (`search`, `preview`, `apply`, `version`) share the app core, `preview` works offline on a local image, and stable exit codes: `0` complete/preview, `2` partial_failure, `3` prerequisite_failure, `4` usage/config error, `5` cancelled/internal. JSON output uses the same model fields as text. <!-- sdd-owner: implementation -->
- [ ] GREEN+REFACTOR: implement `internal/cli/*.go` (flag parsing, rendering) and `cmd/hather/main.go` composition root wiring real source/engine/persistence/platform executor and the exit-code mapping. <!-- sdd-owner: implementation -->
- [ ] RED: write `internal/adapter/macos/macos_test.go` injecting `CommandExecutor` fixtures for all-desktop wallpaper success, `-1743` Automation denial (classified `permission` with `System Settings → Privacy & Security → Automation` remediation), appearance-scheduling detected → safe `conflict` (never silently disables scheduling), malformed platform output, and safe AppleScript path escaping (no shell interpolation). <!-- sdd-owner: implementation -->
- [ ] GREEN+REFACTOR: implement `internal/adapter/macos/macos.go` running `/usr/bin/osascript` via the injected executor for all-desktop wallpaper + light/dark appearance; appearance-scheduling preflight returns explicit conflict/override instead of mutating schedule. Per-Space and Tinted are not exposed. <!-- sdd-owner: implementation -->
- [ ] Add `internal/platform/exec.go` command-execution abstraction (no shell) with `CommandExecutor` interface used by macOS. <!-- sdd-owner: implementation -->
- [ ] RED: write `internal/adapter/ghostty/ghostty_test.go` proving it writes/updates only `~/.config/ghostty/themes/hather.conf` and only the marked `theme = hather` selection pair; an existing unmarked `theme=` or inline palette returns `conflict` without clobbering; an absent Ghostty config dir returns `skipped`/`unavailable` and is non-fatal; unrelated config lines are preserved. <!-- sdd-owner: implementation -->
- [ ] GREEN+REFACTOR: implement `internal/adapter/ghostty/ghostty.go` (Hather-owned theme artifact writer + marked-selection editor via `internal/edit`) and back it with the shared line-oriented edit helper. <!-- sdd-owner: implementation -->
- [ ] RED: extend `internal/adapter/runner_test.go` (or new `internal/adapter/runner.go` tests) proving fixed order `macos,ghostty,herdr,neovim,vscode`, isolated per-adapter results, partial-failure aggregation, and that one failed adapter leaves sibling artifacts/results intact. <!-- sdd-owner: implementation -->
- [ ] GREEN+REFACTOR: implement `internal/adapter/adapter.go` (Adapter contract, `ApplyContext` narrow scope, no cross-adapter calls) and `internal/adapter/runner.go` wired with macOS + Ghostty adapters only in this slice. <!-- sdd-owner: implementation -->
- [ ] RED: write `internal/edit/edit_test.go` proving atomic temp-write + rename, mode preservation, no-symlink-follow, prior owned-value capture, and byte-for-byte no-op on malformed/ambiguous input. <!-- sdd-owner: implementation -->
- [ ] GREEN+REFACTOR: implement `internal/edit/edit.go` shared by all file adapters (used here by Ghostty; reused by Phases 3a–c). <!-- sdd-owner: implementation -->
- [ ] RED: write `internal/adapter/herdr/herdr_test.go` proving only `[theme]`/`[theme.custom]` owned keys change (`theme.name`, `panel_bg`, `accent`, `green`, `blue`, `red`, `yellow` per Phase 0 verification) while shell/keybind/server/unknown values survive byte-level in semantic terms; `HERDR_CONFIG_PATH` honored when set; malformed/duplicate/ambiguous TOML → no write + remediation; every applied result includes `restart_required` future action without implying a running herdr changed. <!-- sdd-owner: implementation -->
- [ ] GREEN+REFACTOR: implement `internal/adapter/herdr/herdr.go` using `internal/edit` line-preserving TOML edits; never invokes `herdr server`, LaunchAgent commands, or process signals. Missing theme sections appended with a Hather-owned comment. <!-- sdd-owner: implementation -->
- [ ] RED: write `internal/adapter/neovim/neovim_test.go` proving atomic write of only `~/.config/nvim/colors/hather.vim` with a stable Hather header + ANSI/editor palette; an existing colorscheme file without the Hather header → `conflict` not overwritten; existing plugin/colorscheme Lua config untouched; result reports next-start/manual `:colorscheme hather` and never attempts a server socket. <!-- sdd-owner: implementation -->
- [ ] GREEN+REFACTOR: implement `internal/adapter/neovim/neovim.go` (generator only via `internal/edit`; writes header-prefixed `hather.vim`). Explicit deterministic colors-dir path, no LazyVim/plugin edits. <!-- sdd-owner: implementation -->
- [ ] RED: write `internal/adapter/vscode/vscode_test.go` proving the parser (hujson) handles comments/trailing commas, merges only Hather-owned color keys beneath `workbench.colorCustomizations` and `editor.tokenColorCustomizations`, preserves unrelated settings, reports `unavailable` when the settings parent dir is absent, returns actionable failure with no full-file replacement on invalid/ambiguous root, and recommends window reload without claiming live apply. No `workbench.colorTheme`, extension install, or process control. <!-- sdd-owner: implementation -->
- [ ] GREEN+REFACTOR: implement `internal/adapter/vscode/vscode.go` using `internal/edit` + the confirmed JSONC parser; safe no-write on unsupported syntax. <!-- sdd-owner: implementation -->
- [ ] RED: write `internal/tui/model_test.go` proving update commands/events route to fake app services and return the same `OperationResult` the CLI uses; render maps results to views for query/results, selected wallpaper, palette preview, selected-adapter toggle, apply progress, and per-adapter results; explicit empty/error/permission/partial-failure views; pressing apply requires a user-selected adapter set (nothing applied by default). <!-- sdd-owner: implementation -->
- [ ] GREEN+REFACTOR: implement `internal/tui/*` (Bubble Tea model minimizing view state; Lip Gloss/Bubbles styling) over `internal/app` only; no source/palette/adapter logic added; no second pipeline. <!-- sdd-owner: implementation -->
- [ ] RED+REFACTOR: add TUI test coverage for unreadable local path and empty Wallhaven-results recovery guidance, proving no blank/destructive state. <!-- sdd-owner: implementation -->
- [ ] RED: write a formula-level test manifest (`test do` block invoking the installed binary) in the owned tap repo asserting `--help` and version output work and the executable is present on PATH; draft as a failing check before packaging wiring. <!-- sdd-owner: implementation -->
- [ ] GREEN+REFACTOR: add release packaging metadata (tagged source archive + checksums, recipe formula/release workflow in the owned `homebrew-hather` tap), pinned Go dep with `-mod=vendor` so formula install fetches no modules from the network, `-ldflags` version injection surfaced by `hather version`/`--version`. Ghostty/herdr/Nvim/VS Code are not formula dependencies. <!-- sdd-owner: implementation -->
- [ ] Add documentation (`docs/` + release notes): supported macOS scope (14+), optional Wallhaven credentials (flag/env only, never config), attribution, Automation permission remediation, per-adapter apply timing (instant/hot-reload vs restart/next-start/manual), missing-target behavior, apply-is-not-a-transaction, and known non-goals; homebrew-core is explicitly outside the gate. <!-- sdd-owner: implementation -->
- [ ] Run a bounded macOS acceptance rehearsal on this machine (scripted, recorded in the release record): fresh `brew install`, `--help`/version, offline local-image `preview`/`apply` with selected adapters, macOS Automation denial path, absent-target adapters returning independent `unavailable` results, and `go test ./...` green from the vendored build. <!-- sdd-owner: implementation -->

## Workload / PR boundary

PR2 is complete: resumed state/app work is approximately 320 authored source/test lines, below the
400-line budget. Delivery remains `stacked-to-main`; no `size:exception`, commit, Phase 1b,
adapter, TUI, packaging, `/usr/bin`, or HatDots work was performed.

---

# Verifier remediation — Phase 1a C1–C3

## Status consumed

Native SDD status: `hather-macos-theme-manager`, `applyState: ready`, OpenSpec authoritative;
`actionContext.mode: repo-local`, workspace/allowed root
`/Users/facundogayoso/projects/Hather`, no warnings. Parent granted a correction-only run with
acquire token `sha256:5bf5a7c57a18f118a9a546d12033ce27e8785d1498409d831edaac0daaeb3292`.
No acquire/reset/settle action was performed; settlement remains parent-owned and is bound to
failed verification `sha256:00b0aa412a1cfc1a3182f37c472201252776f095dc54f3f5a76a8ec89a621a79`.

## Scope and completed remediation

- C1: Added only this fresh correction-cycle evidence. It does not claim or reconstruct a
  historical Phase 1a RED cycle.
- C2: `internal/model/model_test.go` now asserts the exact ordered adapter and operation status
  wire values rather than only checking for non-empty constants.
- C3: `ResolveWallhavenAPIKey` resolves flag over environment without a config fallback. The
  test passes a real flag secret through that resolver, redacts it before result construction,
  writes last-applied and run state, and verifies config, state, and run JSON contain no secret.

## TDD Cycle Evidence

| Correction | Safety net | RED | GREEN | TRIANGULATE | REFACTOR |
| --- | --- | --- | --- | --- | --- |
| C2 exact status contract | `go test ./internal/model ./internal/config` passed before edits | No RED: the existing constants already had the required values; the strengthened exact assertion is a truthful characterization that would fail on any changed value or order. | `go test ./internal/model ./internal/config` passed. | Skipped: this is a structural exact-literal contract with one required ordered output. | No production refactor needed; `gofmt` applied and focused tests remained green. |
| C3 transient credential boundary | `go test ./internal/model ./internal/config` passed before edits | `go test ./internal/config -run '^TestWallhavenAPIKeyIsTransientAndRedacted$'` failed: `undefined: ResolveWallhavenAPIKey`. | Added the minimal flag/environment-only resolver; `go test ./internal/model ./internal/config` passed. | Added empty-flag/environment-key edge; `go test ./internal/config -run '^TestWallhavenAPIKey'` passed. | Kept the resolver as a one-line use of the existing precedence helper; no new persistence or request abstraction. |

## Verification

- `go test ./internal/model ./internal/config` — passed (safety net, GREEN, and refactor check).
- `go test ./internal/config -run '^TestWallhavenAPIKeyIsTransientAndRedacted$'` — RED observed
  as above before production code.
- `go test ./internal/config -run '^TestWallhavenAPIKey'` — passed (triangulation).
- `go test ./...` — passed.
- `gofmt` applied to changed Go files; `git diff --check` — passed.

## Files changed

- `internal/model/model_test.go` — exact stable-status assertions.
- `internal/config/config.go` — transient flag/environment API-key resolution.
- `internal/config/config_test.go` — real secret-bearing resolution plus config/state/run
  non-persistence and environment-fallback proof.
- `openspec/changes/hather-macos-theme-manager/apply-progress.md` — this truthful correction
  record.

## Remaining tasks / delivery boundary

No task checkboxes changed, as required. All Phase 1b-and-later unchecked implementation rows
remain exactly as recorded above. This correction is limited to Phase 1a C1–C3 and adds roughly
40 source/test lines, well within the 120-line correction budget. The existing 419-line PR2
`size:exception` remains recorded and was not broadened; no Phase 1b, adapter, `/usr/bin`,
HatDots, commit, or settlement work occurred.

---

# Corrective retry — Phase 1b source slice (partial; stopped)

## Status consumed

Native SDD status was authoritative and ready for `hather-macos-theme-manager`; action context was
`repo-local` with `/Users/facundogayoso/projects/Hather` as the only allowed edit root. The parent
restricted this retry to the two Wallhaven/local source rows and supplied a 220-line limit.

## TDD Cycle Evidence

| Task | Safety net | RED | GREEN | TRIANGULATE | REFACTOR |
| --- | --- | --- | --- | --- | --- |
| Wallhaven/local source | New package; no pre-existing source tests | `go test ./internal/source` failed on missing `Wallhaven`, `Local`, and `Class` | `go test ./internal/source` passed after the minimal source implementation | Fixtures plus `httptest` covered success, auth, rate, unavailable, malformed, empty, download, cancellation, and local fetch paths | `gofmt` and `git diff --check` passed; no refactor beyond the initial implementation |

## Verification and boundary

- `go test ./internal/source` — passed.
- `go test ./...` — passed.
- `git diff --check` — passed.
- Created `internal/source/{wallhaven_test.go,wallhaven.go,local.go}`; the three files total 309
  lines after `gofmt`, exceeding the parent-provided 220-line ceiling. The slice is therefore
  stopped **partial**: neither Phase 1b task checkbox was changed, and a further retry must choose
  a smaller scoped work unit before completing the task contract.

## Remaining tasks

- [ ] RED: write `internal/source/wallhaven_test.go` using `httptest.Server` for valid search, no-key search, optional auth header, rate-limit classification with retry guidance, auth failure, malformed/empty response, and download failure; assert stable error classes (`invalid_input|unavailable|rate_limited|authentication|download|decode|cancelled`). <!-- sdd-owner: implementation -->
- [ ] GREEN+REFACTOR: implement `internal/source/wallhaven.go` (query encoding, response decode, optional key from flag/env — never from config, request throttle, `internal/config` cache lookup/reuse, download validation, attribution metadata) and `internal/source/local.go` selecting a local path directly with `Fetch` and no `Search`. Expose the `Source` interface. <!-- sdd-owner: implementation -->

---

# Maintainer-approved Phase 1b source finalization

## Status and delivery boundary

Authoritative status: `hather-macos-theme-manager`, `applyState: ready`; `repo-local` edits were
restricted to Hather. The maintainer accepted Wallhaven/local as a standalone 309-line
`stacked-to-main` slice; palette and CLI remain deferred and unchecked. No Go code changed in
this finalization.

## Completed tasks and evidence

The two Wallhaven/local rows are now visibly `- [x]` in `tasks.md`.
`go test ./...` passed as the safety net (app, config, model, and source packages).
Parent settlement uses token `sha256:c05f71f496e6d3187d12baa550e99601c0245157e36a6ac12e2ce358810d345b`
and must bind remediation to failed evidence `sha256:319d203e13b6750322acf5cac49d2fa65ea2f81b8823aee9ff369f8d7b9d4384`.

## TDD Cycle Evidence

| Task | Safety net | RED | GREEN | TRIANGULATE | REFACTOR |
| --- | --- | --- | --- | --- | --- |
| Wallhaven/local source | `go test ./...` passed | Prior source test recorded | Prior source implementation passed | `httptest` success/error/local cases recorded | No code change in finalization |

## Remaining tasks / PR boundary

Palette and CLI Phase 1b rows remain unchecked for separate slices. This bookkeeping-only
finalization changes two task checkboxes and this record, well below the 60-line limit; no commit,
settlement, or other edits were performed.

---

# Phase 1b native palette slice

## Status and scope

Authoritative native status was `ready` for `hather-macos-theme-manager`; `actionContext` was
`repo-local` with `/Users/facundogayoso/projects/Hather` as the sole allowed edit root and no warnings.
Parent scope was only the two native-palette tasks; no source, CLI, adapter, `/usr/bin`, HatDots, commit,
or settlement work occurred.

## Completed tasks and persisted checkboxes

The two palette rows are visibly `- [x]` in `tasks.md`: the RED proof and native engine GREEN/REFACTOR row.

## TDD Cycle Evidence

| Task | Test file | Layer | Safety net | RED | GREEN | TRIANGULATE | REFACTOR |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Palette engine | `internal/palette/palette_test.go` | Unit | N/A (new package) | `go test ./internal/palette -run TestExtract` failed on undefined `New`, `Class`, and `contrast` | focused package tests passed after stdlib decode plus WebP registration | low-contrast diagnostic and HEIC class cases passed | fixed one-bit ordering to keep lossy JPEG/WebP fixtures parity; focused tests stayed green |

## Verification

- `go test ./internal/palette -run 'TestExtract' -v` — passed.
- `go test ./...` — passed.
- `gofmt` on both Go files and `git diff --check` — passed.

## Files changed

- `internal/palette/{palette_test.go,engine.go}` — deterministic engine and fixture-driven tests.
- `internal/palette/testdata/expected-palette.json` — pinned v1 fixture palette.
- `go.mod`, `go.sum` — pinned `golang.org/x/image v0.46.0` for WebP decoding.
- `tasks.md`, `apply-progress.md` — completion evidence only.

## Design notes, remaining work, and boundary

No design deviation: the fixed-order quantizer intentionally uses stable two-level RGB bins so lossy format
fixtures retain one exact ordered palette. The implementation plus golden/module text is 299 lines, within the
300-line parent boundary. All remaining unchecked implementation rows are unchanged in `tasks.md`; the next rows are:

- [ ] RED: write `internal/cli/cli_test.go` proving stdlib `flag.FlagSet` subcommands (`search`, `preview`, `apply`, `version`) share the app core, `preview` works offline on a local image, and stable exit codes: `0` complete/preview, `2` partial_failure, `3` prerequisite_failure, `4` usage/config error, `5` cancelled/internal. JSON output uses the same model fields as text. <!-- sdd-owner: implementation -->
- [ ] GREEN+REFACTOR: implement `internal/cli/*.go` (flag parsing, rendering) and `cmd/hather/main.go` composition root wiring real source/engine/persistence/platform executor and the exit-code mapping. <!-- sdd-owner: implementation -->

The authoritative tasks artifact contains the exact unchanged later adapter, edit, TUI, packaging,
documentation, and acceptance rows. PR boundary: only Phase 1b native palette, a stacked-to-main sub-slice;
no `size:exception` needed.

---

# Phase 1b CLI/main finalization

## Status and scope

Consumed the authoritative `ready` status for `hather-macos-theme-manager`: repo-local action context,
only `/Users/facundogayoso/projects/Hather` allowed, no warnings. This is the selected stacked-to-main
PR3 sub-slice and completes only the two CLI/main implementation rows. The parent owns settlement for
runtime token `sha256:3ddaa77888d87b1685d9e6c87048e31a6600beaaf76ec682b324042dda6da6bf`.

## Completed tasks and checkbox evidence

Both Phase 1b CLI/main rows are visibly `- [x]` in `tasks.md`: the focused CLI RED proof and the CLI/main
GREEN+REFACTOR implementation. No Phase 2 or later checkbox changed.

## TDD Cycle Evidence

| Task | Test file | Layer | Safety net | RED | GREEN | TRIANGULATE | REFACTOR |
| --- | --- | --- | --- | --- | --- | --- | --- |
| CLI commands | `internal/cli/cli_test.go` | Unit/integration seam | `go test ./internal/app ./internal/source ./internal/palette` passed | `go test ./internal/cli -run 'TestRun|TestPreview'` failed: `CLI`/`ExitCode` absent | focused CLI tests passed | fake-core routes/JSON plus a real local fixture preview, usage and all status-code branches passed | `gofmt`; focused tests remained green |
| CLI composition | `internal/cli/cli_test.go` | Integration seam | same baseline | covered by the absent CLI core in the RED command | `cmd/hather` builds with real local/Wallhaven source, palette engine and state store | empty runner returns per-selected-adapter `unavailable`, hence partial result rather than a live-apply claim | kept the runner local to `main`; no Phase 2 platform or adapter abstraction |

## Verification

- `go test ./internal/cli -run 'TestRun|TestPreview' -v` — passed after GREEN and triangulation.
- `go test ./internal/app` — passed after app-core additions.
- `go test ./...` — passed.
- `go build ./...` — passed.
- `gofmt` and `git diff --check` — passed.

## Files changed

- `internal/app/app.go` — shared `Search` and offline `Preview` operations.
- `internal/cli/{cli.go,cli_test.go}` — `flag.FlagSet` command routing, model rendering, status mapping and focused proof.
- `cmd/hather/main.go` — real source/palette/state composition with an honest unavailable adapter runner.
- `openspec/changes/hather-macos-theme-manager/{tasks.md,apply-progress.md}` — only the two persisted CLI/main checks and cumulative evidence.

## Deviation, remaining work, and workload boundary

No real adapter or platform executor was composed because those are Phase 2 work and this slice must not
claim live adapter application. Selected adapters return independent `unavailable` results. The current
exact unchecked implementation rows remain in `tasks.md` (starting with the Phase 2 macOS RED row and
ending with bounded macOS acceptance); none were edited here.

This code/test slice adds 326 authored Go lines over the prior source baseline, below the user-provided
350-line limit. SDD process artifacts are bookkeeping only; no `size:exception`, commit, `/usr/bin`,
HatDots, TUI, packaging, or external target was touched. Next recommended: `sdd-verify`.

---

# Verifier correction — Phase 1b C1–C2

## Status and scope consumed

Consumed the authoritative native SDD status for `hather-macos-theme-manager`: `applyState: ready`,
OpenSpec artifact store, `repo-local` action context, and `/Users/facundogayoso/projects/Hather` as the
only allowed edit root, with no warnings. This correction is restricted to C1 and C2 and the allowed
source, CLI, main, and progress surfaces. No task checkbox or verify report was changed. Parent owns
settlement for attempt `sha256:faa3e2e7429dd46a85f401c43c6215ce15755a1aea5854174321d6e463c4b5b6`
and failed evidence `sha256:c3705ec81ed5098a96f172ef9618a2bd68c8c7756cb051219c6b85395ebfbc14`.

## Correction evidence

- **C1:** `Wallhaven.RequestPolicy` is an injectable synchronous admission function. The composition
  root installs `NewRequestPolicy`, which enforces the documented Wallhaven limit of 45 API calls per
  minute. A local block returns `rate_limited` with a one-minute retry instruction and makes no HTTP
  request; Hather neither sleeps nor retries. Download behavior remains bounded by the existing 25 MiB
  read limit and cache reuse because Wallhaven documents no image-host rate limit.
- **C2:** `search --wallhaven-api-key` is transient and invokes a narrow composition-root callback.
  That callback resolves explicit input over `WALLHAVEN_API_KEY` immediately before remote search. The
  key is used only in the in-memory Wallhaven Authorization header and is absent from CLI output and
  does not enter app request/state structures.

## TDD Cycle Evidence

| Correction | Test file | Layer | Safety net | RED | GREEN | TRIANGULATE | REFACTOR |
| --- | --- | --- | --- | --- | --- | --- | --- |
| C1 local Wallhaven request policy | `internal/source/wallhaven_test.go` | Unit/`httptest` | `go test ./internal/source ./internal/cli ./cmd/hather -count=1` passed before edits | `go test ./internal/source -run '^TestWallhavenRequestPolicyBlocksLocally$' -count=1` failed: unknown `RequestPolicy` field | same focused test passed after the policy seam and actionable local block | `TestNewRequestPolicyAllowsDocumentedMinuteBudget` proved 45 allowed requests and the 46th blocked | one clock read and `>=` limit guard; `gofmt` and focused source tests stayed green |
| C2 transient CLI key composition | `internal/cli/cli_test.go`, `cmd/hather/main_test.go` | Unit/`httptest` | same focused baseline passed before edits | `go test ./internal/cli ./cmd/hather -run 'Test(SearchWallhavenAPIKeyIsTransient|ConfigureWallhavenAPIKeyPrefersExplicitInput)$' -count=1` failed: missing CLI callback and composition function | same command passed after flag parsing and composition-root callback | CLI test proves keyless output while main test proves explicit key wins over an environment key at the remote Authorization header | `gofmt`; focused CLI/main tests remained green |

## Verification

- Focused source/CLI/main safety net — passed before changes.
- C1 RED and C2 RED — observed as recorded above.
- `go test ./internal/source ./internal/cli ./cmd/hather -count=1` — passed.
- `go test ./... -count=1` — passed.
- `go build ./...` — passed.
- `gofmt -d` on all changed Go files and `git diff --check` — no output.

## Files changed

- `internal/source/{wallhaven.go,wallhaven_test.go}` — injectable request admission, documented
  one-minute bound, local rate-limit result, and focused proof.
- `internal/cli/{cli.go,cli_test.go}` — transient search flag, callback seam, and no-output-leak proof.
- `cmd/hather/{main.go,main_test.go}` — policy/key composition and explicit-over-environment remote-search proof.
- `openspec/changes/hather-macos-theme-manager/apply-progress.md` — this correction evidence.

## Remaining tasks and delivery boundary

All persisted task checkboxes remain unchanged by instruction; the Phase 2 macOS RED row and every later
unchecked implementation row remain open exactly as recorded in `tasks.md`. This is a stacked-to-main PR3
correction of approximately 145 authored Go test/source lines, below the 400-line review budget; no
`size:exception`, commit, adapter/TUI/packaging work, or external platform operation occurred.

---

# Independent-verifier correction — Phase 1b C1 clock window

## Scope and implementation

Closed the remaining C1 verifier gap without changing the public `NewRequestPolicy()` API. It now
uses a package-private clock-injected helper; the policy test fixes time, exhausts 45 admissions,
advances the fake clock by more than one minute, and proves the next request is admitted. No task
checkbox or verify report changed.

## Strict TDD evidence

- **RED:** `go test ./internal/source -run '^TestNewRequestPolicyAllowsDocumentedMinuteBudget$' -count=1`
  failed with `undefined: newRequestPolicy` after the clock-window assertion was added.
- **GREEN:** The same command passed after extracting `newRequestPolicy(clock func() time.Time)` and
  keeping `NewRequestPolicy()` as its `time.Now` wrapper.
- **TRIANGULATE:** `go test ./internal/source -count=1` passed, preserving local blocking and all
  existing Wallhaven paths.

## Validation

- `go test ./... -count=1` — passed.
- `go build ./...` — passed.
- `gofmt -w internal/source/wallhaven.go internal/source/wallhaven_test.go` — completed with no output.
- `git diff --check` — passed.

---

# Phase 2a — macOS adapter

## Status and scope consumed

Authoritative status was `ready` for `hather-macos-theme-manager`; action context was `repo-local`,
with `/Users/facundogayoso/projects/Hather` as the only allowed edit root and no warnings. Parent
selected Phase 2a only, strict TDD, a 360-line budget, and parent-owned settlement for attempt
`sha256:96104fd8a5746cbd4eb74da25e7878671141698696011c3051e55573679b2c87`.

## Completed tasks and persisted checkboxes

All three Phase 2a rows are visibly `- [x]` in `tasks.md`: macOS RED proof, macOS GREEN/REFACTOR,
and the no-shell command executor. No other task row changed.

## TDD Cycle Evidence

| Task | Test file | Layer | Safety net | RED | GREEN | TRIANGULATE | REFACTOR |
| --- | --- | --- | --- | --- | --- | --- | --- |
| macOS adapter + executor | `internal/adapter/macos/macos_test.go`, `internal/platform/exec_test.go` | Unit | `go test ./... -count=1` passed before edits | focused packages failed on absent `Executor` | focused packages passed after direct-exec adapter | fixture cases cover permission, scheduling, malformed output, escaped light path, and dark path | `gofmt`; focused tests remained green |

## Verification

- `go test ./internal/platform ./internal/adapter/macos -count=1 -v` — passed after GREEN and triangulation.
- `go test ./... -count=1` — passed.
- `go build ./...` — passed.
- `gofmt -d` on the four changed Go files — no output; `git diff --check` — passed.

## Files changed and boundary

- `internal/platform/{exec.go,exec_test.go}` — injected, direct command executor with no shell.
- `internal/adapter/macos/{macos.go,macos_test.go}` — all-desktop AppleScript, scheduling conflict,
  `-1743` permission remediation, malformed-probe protection, and path escaping proof.
- `tasks.md`, `apply-progress.md` — the three persisted checkboxes and this evidence.

The Phase 2a Go source/test diff is exactly 260 added lines, plus 3 checkbox-line edits and this
process record; it is below the 360-line budget. No real Apple Events ran, no scheduling state was
mutated, and Per-Space/Tinted options were not added. No design deviation: the approved design has
no narrow scheduling override, so the adapter always returns conflict while scheduling is detected.

## Remaining tasks / PR boundary

All remaining unchecked implementation rows in `tasks.md` are unchanged, beginning with the Phase
2b Ghostty RED row. PR4 is the macOS-only stacked-to-main work unit; no commit or settlement occurred.

---

# Independent-verifier correction — Phase 2a launch failures

The verifier found that non-exit launch errors retained exit code `0`, while the adapter checked only
exit codes. RED tests proved a missing command could therefore look successful and that a failed
scheduling probe continued into Apple Events. `Executor` now assigns `-1` to non-exit launch errors. The adapter stops on that sentinel during the
`defaults` scheduling check, while preserving ordinary exit `1` for an absent preference; desktop-probe
and apply commands reject any execution error before reporting success.

- **RED:** `go test ./internal/platform ./internal/adapter/macos -count=1` failed in the executor launch
  error test and both adapter launch-failure tests.
- **TRIANGULATE:** a real-executor-shaped `defaults` exit `1` initially exposed an over-broad `Err`
  check; the adapter now distinguishes ordinary command exit from launch failure.
- **GREEN:** focused tests, `go test ./... -count=1`, `go build ./...`, `go vet ./...`, `gofmt`, and
  `git diff --check` all passed.

## Phase 2a budget decision

The maintainer accepted the final 375-line Phase 2a correction candidate: it exceeds the provisional
360-line sub-budget by 15 lines but remains below the canonical 400-line review budget. No further
scope, implementation, commit, or platform action was added by this decision.

---

# Phase 2b split finalization — phase2b-ghostty-edit

## Status and scope consumed

Authoritative status was `ready` for `hather-macos-theme-manager`; `repo-local` was restricted to
`/Users/facundogayoso/projects/Hather`, with that root allowed and no action-context warnings. The
maintainer-approved stacked split is Ghostty plus `internal/edit` only; the runner rows remain unchecked.
Parent owns settlement for attempt `sha256:e66552b1a9e7ac888e2881e4d5d592cf7a62d8c240bbfb95d1fd027a473760f0`.

## Completed tasks and persisted checkboxes

The two Ghostty rows and two `internal/edit` rows are visibly `- [x]` in `tasks.md`. Existing focused
proof covers owned theme writes, marked selection updates, unowned theme/inline conflicts, absent config,
atomic replacement, mode/prior preservation, symlink rejection, and no-op rejected input.

## TDD Cycle Evidence

| Scope | Safety net | RED | GREEN / triangulation | Refactor |
| --- | --- | --- | --- | --- |
| Ghostty/edit finalization | focused Ghostty/edit tests passed before changes | Added `theme-variant` regression; focused test failed because the marked selector clobbered it | selector now requires the exact `theme` key; regression and existing conflict/update cases pass | no further refactor needed |

## Verification and files changed

- `go test ./internal/edit ./internal/adapter/ghostty -count=1 -v`, `go test ./... -count=1`, and `go build ./...` — passed.
- `gofmt` found no remaining changes; `git diff --check` — passed.
- `internal/adapter/ghostty/{ghostty.go,ghostty_test.go}` and `internal/edit/{edit.go,edit_test.go}` — validated; the Ghostty regression prevents a marked non-`theme` key from being overwritten.

## Workload boundary and remaining tasks

The prior writer's four Go files totalled 292 lines; this finalizer added exactly 14 net Go lines and
replaced one selector line, for 306 total lines, within the 340-line approved split budget. Task/progress
bookkeeping is excluded from the review-code count. Runner work remains explicitly unchecked:

- [ ] RED: extend `internal/adapter/runner_test.go` (or new `internal/adapter/runner.go` tests) proving fixed order `macos,ghostty,herdr,neovim,vscode`, isolated per-adapter results, partial-failure aggregation, and that one failed adapter leaves sibling artifacts/results intact. <!-- sdd-owner: implementation -->
- [ ] GREEN+REFACTOR: implement `internal/adapter/adapter.go` (Adapter contract, `ApplyContext` narrow scope, no cross-adapter calls) and `internal/adapter/runner.go` wired with macOS + Ghostty adapters only in this slice. <!-- sdd-owner: implementation -->

All later Phase 3–5 rows remain unchecked in `tasks.md`; no commit, settlement, adapter-runner, or external
application action occurred.

---

# Independent-verifier correction — Phase 2b Ghostty/edit safety

The verifier invalidated the four checked claims despite green tests. New RED probes demonstrated that
`Edit` read through a symlinked parent before rejecting the write, Ghostty followed a symlinked config
directory and created its theme artifact, and malformed/duplicate marked selections were accepted.
The atomicity check also named the wrong temporary-file prefix.

The minimum correction rejects a symlink target or immediate parent before `Edit` reads, rejects a
symlinked Ghostty config root before artifact creation, validates exactly one marker followed by an exact
`theme =` assignment, and checks the real same-directory `.hather-*` temporary-file convention. A marked
`theme-variant` is preserved byte-for-byte as an invalid selection rather than rewritten.

- **RED:** focused Ghostty/edit tests failed on callback execution through a symlinked parent, artifact
  creation through a symlinked config root, and accepted bare/duplicate marked themes.
- **GREEN:** focused tests, `go test ./... -count=1`, `go build ./...`, `go vet ./...`, `gofmt`, and
  `git diff --check` all passed.

A follow-up verifier then found that a symlinked `config` file was read and the owned theme artifact was
written before `edit.Edit` rejected the target. A new RED test reproduced the artifact leak. Ghostty now
uses `Lstat` on the config target before reading or writing anything; focused and full test/build/vet checks
pass, and the external symlink target remains byte-identical. Final triangulation also normalizes a
trailing-slash config root before `Lstat`, checks write parents before reading target mode, and restores or
removes the owned theme if final config validation rejects. Independent verification passed with no blockers.

---

# Runner contract correction

Manual verification found that `app.Service.Runner` used `any` and retained obsolete `RunRequest` solely
for an old preview test. Changing the field to the compile-time `app.Runner` interface produced the expected
RED compile failure in that test. Removing its unused runner and the legacy type restored GREEN. Focused/full
tests, build, vet, gofmt, diff-check, and independent verification all passed.

---

# Phase 2b — runner and adapter contract

## Status and scope consumed

Authoritative status: `hather-macos-theme-manager`, `applyState: ready`, OpenSpec store; repo-local
`actionContext` permits only `/Users/facundogayoso/projects/Hather`, with no warnings. Strict TDD was
active. Parent owns settlement for `sha256:5d0a059a32aba3c6f38a5d8930061af16325faa98207bceea99ac9f0423b610b`.

## Completed tasks and persisted checkboxes

The two Phase 2b runner rows are visibly `- [x]` in `tasks.md`: runner RED proof and adapter/runner
GREEN+REFACTOR. `ApplyContext` contains only artifact and palette. The runner owns canonical ordering;
the app forwards artifact, palette, and the selected IDs without another ordering pipeline. The composition
root wraps only real macOS and Ghostty adapters; known future and unknown IDs return `unavailable`.

## TDD Cycle Evidence

| Task | Test file | Layer | Safety net | RED | GREEN | TRIANGULATE | REFACTOR |
| --- | --- | --- | --- | --- | --- | --- | --- |
| runner contract | `internal/adapter/runner_test.go` | Unit | `go test ./internal/app ./cmd/hather ./internal/adapter/{macos,ghostty}` passed | focused runner test failed: missing `ApplyContext`, `Adapter`, and `Runner` | focused runner test passed | proves canonical five-ID order, failed Ghostty with retained macOS artifact, disabled untouched, and future unavailable | single canonical order helper; no adapter cross-calls |
| app/root wiring | `internal/app/app_test.go`, `cmd/hather/main_test.go` | Unit | same baseline | focused app test failed on the new runner signature; root test failed on missing `newRunner` | focused app/root tests passed | fake executor proves macOS wiring, absent Ghostty and herdr unavailable, with no commands/files | removed `app.ordered`; kept only a source-compatible unused request type for existing preview tests |

## Verification

- Focused RED: `go test ./internal/adapter -run '^TestRunner' -count=1`; `go test ./internal/app -run '^TestApply' -count=1`; and `go test ./cmd/hather -run '^TestRunnerComposition' -count=1` failed before GREEN as recorded.
- Focused GREEN: runner, app, and root tests passed.
- `go test ./... -count=1`, `go build ./...`, `go vet ./...`, `gofmt -d`, and `git diff --check` — passed.

## Files changed

- `internal/adapter/{adapter.go,runner.go,runner_test.go}` — narrow contract, isolated canonical runner, and proofs.
- `internal/app/{app.go,app_test.go}` — context forwarding; obsolete app ordering removed.
- `cmd/hather/{main.go,main_test.go}` — real macOS/Ghostty composition through the contract and fake-executor proof.
- `tasks.md`, `apply-progress.md` — two persisted checkboxes and cumulative evidence.

## Workload, deviations, and remaining tasks

The review diff is exactly 206 added Go lines (49 removed), below the requested 240-line budget; process
artifacts excluded. No Ghostty/edit behavior changed, no external command/file ran in unit tests, and no
design deviation occurred. PR boundary: Phase 2b runner/contract only; no commit or settlement.

## Phase 3a — herdr targeted TOML adapter

Two delegated workers timed out without creating files. The maintainer authorized a bounded direct fallback
under the existing 340-line objective. The implementation uses pinned `github.com/BurntSushi/toml` v1.6.0
to validate the complete input and output, then applies line-preserving edits through `internal/edit`.

- **RED:** focused tests failed because `Adapter` did not exist.
- **GREEN:** fixture-backed tests proved exact theme ownership, preservation of unrelated settings,
  `HERDR_CONFIG_PATH`, no-write invalid/duplicate/multiline input, missing-section append, and
  `restart_required` guidance.
- **Composition RED/GREEN:** main first reported herdr unavailable, then passed after adapter wiring.
- **Verifier correction:** a focused RED proved assignment-like text inside an unowned multiline TOML
  string could be rewritten; multiline tracking and final TOML validation now preserve it.

Focused/full tests, build, vet, gofmt, and diff-check pass. No herdr process, server, LaunchAgent, signal,
shell, user-home file, or live reload was touched. Both Phase 3a task rows are checked.

Remaining exact unchecked implementation rows:
- [ ] RED: write `internal/adapter/neovim/neovim_test.go` proving atomic write of only `~/.config/nvim/colors/hather.vim` with a stable Hather header + ANSI/editor palette; an existing colorscheme file without the Hather header → `conflict` not overwritten; existing plugin/colorscheme Lua config untouched; result reports next-start/manual `:colorscheme hather` and never attempts a server socket. <!-- sdd-owner: implementation -->
- [ ] GREEN+REFACTOR: implement `internal/adapter/neovim/neovim.go` (generator only via `internal/edit`; writes header-prefixed `hather.vim`). Explicit deterministic colors-dir path, no LazyVim/plugin edits. <!-- sdd-owner: implementation -->
- [ ] RED: write `internal/adapter/vscode/vscode_test.go` proving the parser (hujson) handles comments/trailing commas, merges only Hather-owned color keys beneath `workbench.colorCustomizations` and `editor.tokenColorCustomizations`, preserves unrelated settings, reports `unavailable` when the settings parent dir is absent, returns actionable failure with no full-file replacement on invalid/ambiguous root, and recommends window reload without claiming live apply. No `workbench.colorTheme`, extension install, or process control. <!-- sdd-owner: implementation -->
- [ ] GREEN+REFACTOR: implement `internal/adapter/vscode/vscode.go` using `internal/edit` + the confirmed JSONC parser; safe no-write on unsupported syntax. <!-- sdd-owner: implementation -->
- [ ] RED: write `internal/tui/model_test.go` proving update commands/events route to fake app services and return the same `OperationResult` the CLI uses; render maps results to views for query/results, selected wallpaper, palette preview, selected-adapter toggle, apply progress, and per-adapter results; explicit empty/error/permission/partial-failure views; pressing apply requires a user-selected adapter set (nothing applied by default). <!-- sdd-owner: implementation -->
- [ ] GREEN+REFACTOR: implement `internal/tui/*` (Bubble Tea model minimizing view state; Lip Gloss/Bubbles styling) over `internal/app` only; no source/palette/adapter logic added; no second pipeline. <!-- sdd-owner: implementation -->
- [ ] RED+REFACTOR: add TUI test coverage for unreadable local path and empty Wallhaven-results recovery guidance, proving no blank/destructive state. <!-- sdd-owner: implementation -->
- [ ] RED: write a formula-level test manifest (`test do` block invoking the installed binary) in the owned tap repo asserting `--help` and version output work and the executable is present on PATH; draft as a failing check before packaging wiring. <!-- sdd-owner: implementation -->
- [ ] GREEN+REFACTOR: add release packaging metadata (tagged source archive + checksums, recipe formula/release workflow in the owned `homebrew-hather` tap), pinned Go dep with `-mod=vendor` so formula install fetches no modules from the network, `-ldflags` version injection surfaced by `hather version`/`--version`. Ghostty/herdr/Nvim/VS Code are not formula dependencies. <!-- sdd-owner: implementation -->
- [ ] Add documentation (`docs/` + release notes): supported macOS scope (14+), optional Wallhaven credentials (flag/env only, never config), attribution, Automation permission remediation, per-adapter apply timing (instant/hot-reload vs restart/next-start/manual), missing-target behavior, apply-is-not-a-transaction, and known non-goals; homebrew-core is explicitly outside the gate. <!-- sdd-owner: implementation -->
- [ ] Run a bounded macOS acceptance rehearsal on this machine (scripted, recorded in the release record): fresh `brew install`, `--help`/version, offline local-image `preview`/`apply` with selected adapters, macOS Automation denial path, absent-target adapters returning independent `unavailable` results, and `go test ./...` green from the vendored build. <!-- sdd-owner: implementation -->

---

# Phase 3b — Neovim additive colorscheme

## Status and completion

Consumed authoritative `ready` status: OpenSpec, strict TDD, `repo-local`, Hather-only edit root, no action-context warnings. The two Phase 3b rows are visibly `- [x]` in `tasks.md` after proof.

## TDD Cycle Evidence

| Task | Safety net | RED | GREEN | TRIANGULATE | REFACTOR |
| --- | --- | --- | --- | --- | --- |
| Neovim colorscheme | `go test ./cmd/hather ./internal/adapter ./internal/edit` passed | focused Neovim/main tests failed on absent `Adapter` and `Header` | focused tests passed after the generator and composition wire | owned update and cancellation paths passed in isolated temp dirs | minimal formatter/helper retained; focused tests stayed green |

## Evidence and files

- `internal/adapter/neovim/{neovim.go,neovim_test.go}` — `internal/edit.Write` atomically writes only `colors/hather.vim`; exact header ownership, ANSI/editor palette output, conflict byte preservation, untouched Lua sentinel, manual/next-start guidance, and no-server diagnostic are covered.
- `cmd/hather/{main.go,main_test.go}` — runner uses `~/.config/nvim` through `DefaultConfigRoot`; test uses a temp home.
- `go test ./internal/adapter/neovim ./cmd/hather -count=1`, `go test ./... -count=1`, `go build ./...`, `go vet ./...`, `gofmt -d`, and `git diff --check` — passed.

## Boundary and remaining work

No plugin, Lua, server socket, process, shell, or remote-send behavior was added or invoked. No design deviation. Phase 3b initially used 217 changed lines, below the parent 320-line ceiling.

Independent verification found that a symlinked path component below the configured home could escape the
Neovim root. A focused RED added `BoundaryRoot` coverage for a symlinked `.config`; the adapter now walks
existing components between the trusted home and config root, rejecting symlinks or lexical escape before
creating directories. Main supplies `home` as the boundary. Focused/full tests, build, vet, diff-check, and
independent re-verification pass. All later unchecked rows remain exactly in `tasks.md`; Phase 3c is next.

Remaining exact unchecked implementation rows:
- [ ] RED: write `internal/adapter/vscode/vscode_test.go` proving the parser (hujson) handles comments/trailing commas, merges only Hather-owned color keys beneath `workbench.colorCustomizations` and `editor.tokenColorCustomizations`, preserves unrelated settings, reports `unavailable` when the settings parent dir is absent, returns actionable failure with no full-file replacement on invalid/ambiguous root, and recommends window reload without claiming live apply. No `workbench.colorTheme`, extension install, or process control. <!-- sdd-owner: implementation -->
- [ ] GREEN+REFACTOR: implement `internal/adapter/vscode/vscode.go` using `internal/edit` + the confirmed JSONC parser; safe no-write on unsupported syntax. <!-- sdd-owner: implementation -->
- [ ] RED: write `internal/tui/model_test.go` proving update commands/events route to fake app services and return the same `OperationResult` the CLI uses; render maps results to views for query/results, selected wallpaper, palette preview, selected-adapter toggle, apply progress, and per-adapter results; explicit empty/error/permission/partial-failure views; pressing apply requires a user-selected adapter set (nothing applied by default). <!-- sdd-owner: implementation -->
- [ ] GREEN+REFACTOR: implement `internal/tui/*` (Bubble Tea model minimizing view state; Lip Gloss/Bubbles styling) over `internal/app` only; no source/palette/adapter logic added; no second pipeline. <!-- sdd-owner: implementation -->
- [ ] RED+REFACTOR: add TUI test coverage for unreadable local path and empty Wallhaven-results recovery guidance, proving no blank/destructive state. <!-- sdd-owner: implementation -->
- [ ] RED: write a formula-level test manifest (`test do` block invoking the installed binary) in the owned tap repo asserting `--help` and version output work and the executable is present on PATH; draft as a failing check before packaging wiring. <!-- sdd-owner: implementation -->
- [ ] GREEN+REFACTOR: add release packaging metadata (tagged source archive + checksums, recipe formula/release workflow in the owned `homebrew-hather` tap), pinned Go dep with `-mod=vendor` so formula install fetches no modules from the network, `-ldflags` version injection surfaced by `hather version`/`--version`. Ghostty/herdr/Nvim/VS Code are not formula dependencies. <!-- sdd-owner: implementation -->
- [ ] Add documentation (`docs/` + release notes): supported macOS scope (14+), optional Wallhaven credentials (flag/env only, never config), attribution, Automation permission remediation, per-adapter apply timing (instant/hot-reload vs restart/next-start/manual), missing-target behavior, apply-is-not-a-transaction, and known non-goals; homebrew-core is explicitly outside the gate. <!-- sdd-owner: implementation -->
- [ ] Run a bounded macOS acceptance rehearsal on this machine (scripted, recorded in the release record): fresh `brew install`, `--help`/version, offline local-image `preview`/`apply` with selected adapters, macOS Automation denial path, absent-target adapters returning independent `unavailable` results, and `go test ./...` green from the vendored build. <!-- sdd-owner: implementation -->

---

## Phase 3c — VS Code JSONC settings merge

## TDD and verification evidence

- RED: focused tests initially failed because the VS Code adapter and composition were absent.
- GREEN: `internal/adapter/vscode` parses JSONC with pinned hujson, merges only four owned color entries, rejects invalid/non-object/duplicate targets without writing, and reports missing parent directories as unavailable.
- TRIANGULATE: tests cover comments, trailing commas, unrelated nested/root values, absent parents, lexical boundary escape, and deterministic section/key order.
- Verifier correction: replaced map iteration with fixed ordered entries and avoided `Standardize`, preserving JSONC comments while retaining deterministic output.
- Composition uses the default macOS VS Code settings path constrained beneath the trusted home.
- Applied results recommend a window reload; no theme selection, extension installation, process, shell, or live-control behavior exists.
- `go test ./internal/adapter/vscode ./cmd/hather -count=1`, full tests, build, vet, gofmt, and `git diff --check` pass.

Both Phase 3c rows are checked. Phase 4 TUI is next.

---

## Phase 4 — Bubble Tea TUI

Phase 4 is complete in bounded slices. Phase 4a adds a presentation-only Bubble Tea model over the
existing app service, with Bubbles input, minimal Lip Gloss styling, Wallhaven/local selection, result
navigation, palette preview, deterministic adapter toggles (none selected by default), apply progress,
and shared `OperationResult` rendering. Tests execute returned commands against fakes and cover empty
searches, unreadable local paths, service errors, permission diagnostics, partial failures, and explicit
no-adapter guidance. No second source, palette, or adapter pipeline exists.

The initial Phase 4a accounting was 409 lines. After the maintainer-authorized budget reset, spinner-only
state and redundant test ceremony were removed without behavior loss; the final equivalent slice is 397
lines, below the canonical 400-line budget. Phase 4b separately wires explicit `hather tui` dispatch via
an injected program runner. It preserves all CLI and bare-argument behavior and reports terminal startup
errors through stderr with a non-zero exit. Focused tests (20 runs), full tests, build, vet, gofmt, and
diff-check pass; no network, Apple Events, target app, process, or user-home mutation occurred.

Independent static verification accepted composition and identified model-state gaps. A separate correction slice
now gives focused text input ownership of printable keys, clears stale selections/results on source changes, keeps
apply single-flight, labels every adapter choice, and renders complete palette and per-adapter evidence. Its RED
regressions pass 20 consecutive focused runs plus the full test/build/vet/diff suite. A final delegated re-verifier
was unavailable by timeout; no finding remains reproducible in the bounded tests or parent static readback.

---

# Phase 5c — documentation and bounded release rehearsal

## Status consumed

Authoritative native status: `hather-macos-theme-manager`, OpenSpec store, `applyState: ready`, strict TDD,
repo-local action context, and Hather as the only allowed edit root; no action-context warnings. The parent selected
the Phase 5 documentation/rehearsal closeout with token
`sha256:c823ce7d3bc2a3994988e83383a2ea7c338fbd711872ca458b0fe023a2b9214d` and a 360-line limit.

## Completed tasks and persisted checkboxes

All four Phase 5 implementation rows are visibly `- [x]` in `tasks.md`:

- The owned-tap `Formula/hather.rb.in` has the RED `test do` manifest asserting installed executability, `--help`, and
  `--version`; the rendered formula passed `brew test`.
- Prepared release metadata is the tag-triggered release workflow, source archive plus SHA-256 sidecar, vendored
  `-mod=vendor` formula build, and `-ldflags` version injection. Optional adapters are not formula dependencies.
- `README.md` and `docs/release-v0.1.0.md` document the supported scope, credentials, attribution, remediation,
  timing, missing targets, non-transactional apply, and non-goals.
- The bounded rehearsal below passed. It used a temporary local tap and archive; no production tag, GitHub release,
  release asset, owned-tap push, or public formula install was performed.

## TDD Cycle Evidence

Documentation and release records do not change runtime behavior, so a RED/GREEN production-code cycle was not
applicable. The formula test manifest was the packaging behavior proof: it passed only after release wiring built an
installed `hather` binary with the injected version. The fixture-injected Automation-denial check was retained rather
than changing live TCC state.

## Rehearsal and verification

- In a fresh temporary tap, rendered the owned formula against a local vendored `v0.1.0` source archive and SHA-256.
  `brew install --build-from-source`, `brew test`, installed `--help`, and installed `--version` (`0.1.0`) passed;
  the temporary formula and tap were removed.
- From the temporary archive source, `go test -mod=vendor ./...` and the vendored `-ldflags` build passed.
- The installed binary passed offline local-image `preview`; selected Ghostty/herdr/Neovim/VS Code `apply` returned
  exit `2` with independent unavailable results for absent config targets and an owned Neovim artifact.
- `go test -mod=vendor ./internal/adapter/macos -run '^TestApplyClassifiesAutomationDenial$' -count=1` passed.
  This proves the injected denial/remediation path only; no live Automation denial or TCC mutation was attempted.
- Final root checks passed: `go test ./...`, `go build ./...`, `go vet ./...`, `git diff --check`, and
  no-whitespace checks for all four untracked/changed phase artifacts.

## Files changed

- `README.md` — install, safe usage, limits, and adapter reference.
- `docs/release-v0.1.0.md` — prepared-metadata and temporary-rehearsal record, explicitly not a production publish.
- `tasks.md` and this progress record — all Phase 5 completion evidence.

## Workload, deviations, and remaining tasks

This documentation/process slice is 151 authored lines, below the 360-line limit. No design deviation: release
metadata remains prepared until a real tag runs the workflow. There are no unchecked implementation-owned task rows.

---

# Final-verification correction 6.1 — runtime configuration wiring

## Status and scope consumed

Consumed the authoritative `ready` status for `hather-macos-theme-manager`: OpenSpec, strict TDD,
`repo-local` action context, and `/Users/facundogayoso/projects/Hather` as the only allowed edit root,
with no warnings. The parent selected only task 6.1 under the inherited 360-line ceiling. No external
application, network, user-home, commit, tag, push, or publication action occurred.

## Completed task and checkbox evidence

The final-verification Slice 1 / task 6.1 checkbox is visibly `- [x]` in `tasks.md`; no other final-
verification checkbox changed. `config.Load` reads the defined Hather config path and accepts ordinary
`enabled_adapters` only. The composition root resolves CLI adapters over `HATHER_ENABLED_ADAPTERS`,
config, and an empty default before the request reaches the real app service and runner. The config
model has no Wallhaven-key field; the root proof supplies config, environment, and explicit secrets and
proves none appears in CLI JSON or persisted state.

## TDD Cycle Evidence

| Task | Test file | Layer | Safety net | RED | GREEN | TRIANGULATE | REFACTOR |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 6.1 runtime configuration wiring | `internal/config/config_test.go`, `cmd/hather/main_test.go` | composition/integration | `go test ./internal/config ./cmd/hather -count=1` passed | focused command failed on undefined `Load` and `newProgram` | focused tests passed after TOML loading and the composition wrapper | root cases prove flag > environment > config > empty default, while the config-root case proves credential non-persistence | simplified the precedence assertion; focused tests stayed green |

## Verification

- Focused RED: `go test ./internal/config ./cmd/hather -run 'TestLoadReadsOrdinarySettingsWithoutCredential|TestProgramCompositionLoadsConfiguredAdapters' -count=1` failed on missing `Load` and `newProgram`.
- Focused GREEN/refactor: `go test ./internal/config ./cmd/hather -run 'TestLoadReadsOrdinarySettingsWithoutCredential|TestProgramComposition' -count=1 -v` passed.
- Focused triangulation: `go test ./cmd/hather -run '^TestProgramComposition' -count=1 -v` passed with flag, environment, config, and default cases.
- Final: `go test ./... -count=1 -timeout 90s`, `go build ./...`, `go vet ./...`, and `git diff --check` passed.

## Files, workload, and residual risk

- `internal/config/{config.go,config_test.go}` — TOML ordinary-settings loader, adapter precedence, and no-credential config representation proof.
- `cmd/hather/{main.go,main_test.go}` — real composition wrapper and fixture-backed root proofs.
- `tasks.md`, `apply-progress.md` — this one persisted completion and cumulative evidence.

The source/test delta is exactly **140 changed lines** (131 additions, 9 deletions), below the inherited
360-line ceiling; process artifacts are excluded from this review count. The repository has no tracked
baseline, so `git diff --check` completed successfully but reports no untracked-file patch. No target-path
option existed in the ordinary Hather config surface, so none was invented; existing adapters retain their
safe documented target defaults and environment handling.

## Remaining implementation tasks

- [ ] Slice 2 (target ≤320 lines) — RED → GREEN → TRIANGULATE → REFACTOR in `internal/cli/{cli.go,cli_test.go}` and `internal/tui/{model.go,model_test.go}`: prove CLI and TUI select and render the same Wallhaven result labels/metadata, preview the same selected local or Wallhaven wallpaper, and gate apply until a successful preview plus an explicit adapter selection; keep all paths on `internal/app`. <!-- sdd-owner: implementation -->
- [ ] Slice 3 (target ≤180 lines) — RED → GREEN → TRIANGULATE → REFACTOR in `internal/adapter/herdr/{herdr.go,herdr_test.go,testdata/config.toml}`: replace unverified `blue`/`yellow` ownership with the Phase 0 observed `[theme.custom]` keys from `docs/platform-verification.md` (`sidebar_bg`, `active_row_bg`, `selection_bg`, `panel_bg`, `accent`, `red`, `green`), preserving all non-owned TOML bytes/semantics and restart-only guidance. <!-- sdd-owner: implementation -->
- [ ] Slice 4 (target ≤320 lines) — RED → GREEN → TRIANGULATE → REFACTOR in `internal/edit/{edit.go,edit_test.go}`, `internal/adapter/ghostty/{ghostty.go,ghostty_test.go}`, and `internal/adapter/herdr/{herdr.go,herdr_test.go}`: reject every symlinked ancestor or lexical escape before either adapter reads, writes, or creates an owned artifact; prove external targets and config bytes remain unchanged for `HERDR_CONFIG_PATH` and Ghostty config/theme roots. <!-- sdd-owner: implementation -->
- [ ] Slice 5 (target ≤360 lines) — RED → GREEN → TRIANGULATE → REFACTOR across `go.mod`, `.github/workflows/release.yml`, and the `facundogayoso/homebrew-hather/Formula/hather.rb.in` discovery target: align the Go toolchain used by CI/release/formula, retain vendored offline builds, and add formula/installed-binary acceptance proving a local-image preview/apply remains usable while every selected missing target yields its own actionable unavailable result. <!-- sdd-owner: implementation -->
- [ ] Slice 6 (target ≤120 lines) — reconcile executable evidence and unchecked work across `openspec/changes/hather-macos-theme-manager/{tasks.md,apply-progress.md,verify-report.md}`: correct completed-task claims, verifier counts, evidence revisions, and residual-risk wording so no task or report claims behavior without the corresponding focused proof. <!-- sdd-owner: implementation -->
- [ ] Slice 7 (target ≤120 lines) — run and record final focused regression checks plus `go test ./... -count=1 -timeout 90s`, `go build ./...`, `go vet ./...`, and `git diff --check`; perform the bounded macOS/Homebrew acceptance from `docs/` and update `openspec/changes/hather-macos-theme-manager/verify-report.md` with the observed verdict, failures, and remaining manual limitations. <!-- sdd-owner: implementation -->

---

# Final-verification correction 6.2 — CLI/TUI Wallhaven parity

## Status and delivery boundary

Consumed the authoritative `ready` status for `hather-macos-theme-manager`: OpenSpec, strict TDD,
`repo-local` action context, and `/Users/facundogayoso/projects/Hather` as the only allowed edit root,
with no warnings. The human resolved the workload gate as `auto-chain` with `stacked-to-main`; this is
that bounded work unit only. No commit, branch, push, PR, network call, user-home mutation, or external
application action occurred.

## Completed task and checkbox evidence

Only final-verification Slice 2 / correction 6.2 is visibly `- [x]` in `tasks.md`. Wallhaven search
now assigns deterministic `Wallhaven <id>` titles plus resolution/category/purity metadata. CLI text
search output and TUI result views show those values; CLI preview/apply accepts a selected Wallhaven ID
and image URL, retains the transient flag-over-environment credential callback, and renders status,
artifact, operation diagnostics, adapter errors, changed paths, generated artifacts, and follow-up.
TUI blocks apply unless the current selected wallpaper has completed preview and at least one adapter is
selected. `internal/app` classifies context cancellation as `cancelled`, not prerequisite failure.

## TDD Cycle Evidence

| Task | Test file | Layer | Safety net | RED | GREEN | TRIANGULATE | REFACTOR |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 6.2 Wallhaven parity and preview gate | `internal/source/wallhaven_test.go`, `internal/cli/cli_test.go`, `internal/tui/model_test.go`, `internal/app/app_test.go` | unit/composition | `go test ./internal/source ./internal/cli ./internal/tui ./internal/app -count=1` passed before edits | Same command failed after new proofs: absent Wallhaven title/metadata, no CLI Wallhaven selection path, missing TUI metadata/gate, and cancellation reported as prerequisite failure | Same focused command passed after the minimum source/app/CLI/TUI changes | TUI proves an unpreviewed selected result and a changed selection both block apply, then a successful preview permits it; app proves both direct and context-observed cancellation | No abstraction added; fixed metadata keys and the existing app pipeline remain the single path |

## Verification

- Focused RED: `go test ./internal/source ./internal/cli ./internal/tui ./internal/app -count=1` failed as recorded above.
- Focused GREEN/TRIANGULATE: the same command passed.
- Final: `go test ./... -count=1 -timeout 90s`, `go build ./...`, `go vet ./...`, and `git diff --check` passed.

## Files and workload

- `internal/source/{wallhaven.go,wallhaven_test.go}` — deterministic Wallhaven title and fixed metadata.
- `internal/cli/{cli.go,cli_test.go}` — selected remote-wallpaper flags, credential callback use, and complete text result rendering.
- `internal/tui/{model.go,model_test.go}` — metadata rendering and current-selection preview gate.
- `internal/app/{app.go,app_test.go}` — cancellation status classification.
- `tasks.md` and this cumulative record — Slice 2 checkbox and evidence.

The repository has no tracked baseline, so Git cannot calculate a patch line count for its all-untracked
worktree; the bounded source/test edit stayed under the 400-line objective. Delivery boundary is
stacked-to-main only; no delivery operation was performed.

## Remaining implementation tasks

- [ ] Slice 3 (target ≤180 lines) — RED → GREEN → TRIANGULATE → REFACTOR in `internal/adapter/herdr/{herdr.go,herdr_test.go,testdata/config.toml}`: replace unverified `blue`/`yellow` ownership with the Phase 0 observed `[theme.custom]` keys from `docs/platform-verification.md` (`sidebar_bg`, `active_row_bg`, `selection_bg`, `panel_bg`, `accent`, `red`, `green`), preserving all non-owned TOML bytes/semantics and restart-only guidance. <!-- sdd-owner: implementation -->
- [ ] Slice 4 (target ≤320 lines) — RED → GREEN → TRIANGULATE → REFACTOR in `internal/edit/{edit.go,edit_test.go}`, `internal/adapter/ghostty/{ghostty.go,ghostty_test.go}`, and `internal/adapter/herdr/{herdr.go,herdr_test.go}`: reject every symlinked ancestor or lexical escape before either adapter reads, writes, or creates an owned artifact; prove external targets and config bytes remain unchanged for `HERDR_CONFIG_PATH` and Ghostty config/theme roots. <!-- sdd-owner: implementation -->
- [ ] Slice 5 (target ≤360 lines) — RED → GREEN → TRIANGULATE → REFACTOR across `go.mod`, `.github/workflows/release.yml`, and the `facundogayoso/homebrew-hather/Formula/hather.rb.in` discovery target: align the Go toolchain used by CI/release/formula, retain vendored offline builds, and add formula/installed-binary acceptance proving a local-image preview/apply remains usable while every selected missing target yields its own actionable unavailable result. <!-- sdd-owner: implementation -->
- [ ] Slice 6 (target ≤120 lines) — reconcile executable evidence and unchecked work across `openspec/changes/hather-macos-theme-manager/{tasks.md,apply-progress.md,verify-report.md}`: correct completed-task claims, verifier counts, evidence revisions, and residual-risk wording so no task or report claims behavior without the corresponding focused proof. <!-- sdd-owner: implementation -->
- [ ] Slice 7 (target ≤120 lines) — run and record final focused regression checks plus `go test ./... -count=1 -timeout 90s`, `go build ./...`, `go vet ./...`, and `git diff --check`; perform the bounded macOS/Homebrew acceptance from `docs/` and update `openspec/changes/hather-macos-theme-manager/verify-report.md` with the observed verdict, failures, and remaining manual limitations. <!-- sdd-owner: implementation -->

---

# Reopened Slice 2 correction — stale preview authorization

## Status and delivery boundary

Consumed the authoritative `ready` status for `hather-macos-theme-manager`: OpenSpec, strict TDD,
`repo-local` action context, allowed root `/Users/facundogayoso/projects/Hather`, and no warnings.
The resolved delivery path is `auto-chain`, `stacked-to-main`; this correction is the Slice 2 work unit
only. No delivery action, network request, external application/user-home mutation, commit, or publication occurred.

## Completed task and files

Slice 2 is visibly `- [x]` in `tasks.md`. `internal/tui/model.go` now returns the wallpaper captured
when preview starts with `previewResult`, and successful preview state records that payload rather than
the current selection. `internal/tui/model_test.go` proves a preview of A that finishes after selection
moves to B does not authorize apply B.

## TDD Cycle Evidence

| Task | Test file | Layer | Safety net | RED | GREEN | TRIANGULATE | REFACTOR |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Slice 2 stale preview gate | `internal/tui/model_test.go` | Unit | `go test ./internal/tui -count=1` passed | New delayed-preview test failed because completion assigned `m.Selected` (B) to `m.previewed` | Focused regression passed after carrying the captured wallpaper in `previewResult` | Existing successful-preview/current-selection path remains covered by `TestWallhavenLabelsAndPreviewGateApply`; the full TUI package passed | No additional refactor needed; `gofmt` applied |

## Verification

- Focused RED: `go test ./internal/tui -run '^TestPreviewCompletingAfterSelectionChangesDoesNotAuthorizeNewSelection$' -count=1` failed as expected.
- Focused GREEN: the same command passed; `go test ./internal/tui -count=1` passed.
- Final: `go test ./... -count=1 -timeout 90s`, `go build ./...`, `go vet ./...`, and `git diff --check` passed.

## Remaining implementation tasks

- [ ] Slice 3 (target ≤180 lines) — RED → GREEN → TRIANGULATE → REFACTOR in `internal/adapter/herdr/{herdr.go,herdr_test.go,testdata/config.toml}`: replace unverified `blue`/`yellow` ownership with the Phase 0 observed `[theme.custom]` keys from `docs/platform-verification.md` (`sidebar_bg`, `active_row_bg`, `selection_bg`, `panel_bg`, `accent`, `red`, `green`), preserving all non-owned TOML bytes/semantics and restart-only guidance. <!-- sdd-owner: implementation -->
- [ ] Slice 4 (target ≤320 lines) — RED → GREEN → TRIANGULATE → REFACTOR in `internal/edit/{edit.go,edit_test.go}`, `internal/adapter/ghostty/{ghostty.go,ghostty_test.go}`, and `internal/adapter/herdr/{herdr.go,herdr_test.go}`: reject every symlinked ancestor or lexical escape before either adapter reads, writes, or creates an owned artifact; prove external targets and config bytes remain unchanged for `HERDR_CONFIG_PATH` and Ghostty config/theme roots. <!-- sdd-owner: implementation -->
- [ ] Slice 5 (target ≤360 lines) — RED → GREEN → TRIANGULATE → REFACTOR across `go.mod`, `.github/workflows/release.yml`, and the `facundogayoso/homebrew-hather/Formula/hather.rb.in` discovery target: align the Go toolchain used by CI/release/formula, retain vendored offline builds, and add formula/installed-binary acceptance proving a local-image preview/apply remains usable while every selected missing target yields its own actionable unavailable result. <!-- sdd-owner: implementation -->
- [ ] Slice 6 (target ≤120 lines) — reconcile executable evidence and unchecked work across `openspec/changes/hather-macos-theme-manager/{tasks.md,apply-progress.md,verify-report.md}`: correct completed-task claims, verifier counts, evidence revisions, and residual-risk wording so no task or report claims behavior without the corresponding focused proof. <!-- sdd-owner: implementation -->
- [ ] Slice 7 (target ≤120 lines) — run and record final focused regression checks plus `go test ./... -count=1 -timeout 90s`, `go build ./...`, `go vet ./...`, and `git diff --check`; perform the bounded macOS/Homebrew acceptance from `docs/` and update `openspec/changes/hather-macos-theme-manager/verify-report.md` with the observed verdict, failures, and remaining manual limitations. <!-- sdd-owner: implementation -->

## Workload and residual risk

The surgical source/test correction is well below Slice 2's 320-line target. The only remaining risk is
that later unchecked slices have not been implemented or verified; this change preserves the existing
`internal/app` path and makes no adapter or delivery change.

---

# Final-verification correction Slice 3 — observed herdr custom keys

## Status and delivery boundary

Consumed authoritative `ready` status for `hather-macos-theme-manager`: OpenSpec, strict TDD,
`repo-local` action context, `/Users/facundogayoso/projects/Hather` as the sole allowed edit root, and no
warnings. The user resolved the workload gate as `auto-chain`, `stacked-to-main`; this work unit is Slice 3
only. No network, user-home, external application, delivery, commit, or publication action occurred.

## Completed task and persisted checkbox

Slice 3 is visibly `- [x]` in `tasks.md`. herdr now owns only `theme.name` and the Phase 0 observed
`[theme.custom]` keys: `sidebar_bg`, `active_row_bg`, `selection_bg`, `panel_bg`, `accent`, `red`, and
`green`. The fixture retains unowned `blue` and `yellow`, and the focused proof confirms they remain
byte-identical. Existing malformed, duplicate, multiline, and restart-only behavior remains covered.

## TDD Cycle Evidence

| Task | Test file | Layer | Safety net | RED | GREEN | TRIANGULATE | REFACTOR |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Slice 3 observed herdr keys | `internal/adapter/herdr/herdr_test.go` | Unit | `go test ./internal/adapter/herdr -count=1` passed | The owned-key fixture assertion failed because `sidebar_bg`, `active_row_bg`, and `selection_bg` were absent and existing `blue`/`yellow` were rewritten | `go test ./internal/adapter/herdr -run '^TestApplyChangesOnlyOwnedThemeValues$' -count=1` passed after the minimal owned-key map update | `TestApplyAppendsMissingOwnedSections` proves an absent section appends all seven observed keys and never appends `blue` or `yellow` | No further refactor: the existing two maps remain the smallest line-preserving implementation |

## Verification

- Focused safety net, RED, GREEN, and triangulation commands above completed with the recorded outcomes.
- `go test ./internal/adapter/herdr -count=1` — passed.
- `go test ./... -count=1 -timeout 90s` — passed.
- `go build ./...` — passed.
- `go vet ./...` — passed.
- `gofmt -d internal/adapter/herdr/herdr.go internal/adapter/herdr/herdr_test.go` and `git diff --check` — no output.

## Files and workload

- `internal/adapter/herdr/herdr.go` — replace unverified `blue`/`yellow` ownership with exactly the observed keys.
- `internal/adapter/herdr/herdr_test.go` — prove owned updates, unowned `blue`/`yellow` preservation, and missing-section behavior.
- `internal/adapter/herdr/testdata/config.toml` — retain unowned `blue` and `yellow` fixture values.
- `tasks.md` and this cumulative record — persisted completion and evidence.

The source/test/fixture edit is approximately 25 changed lines, below the 180-line target. This repository is
entirely untracked, so Git cannot calculate a slice patch count; no size exception is needed. No design
deviation: file updates retain restart-only guidance and the existing TOML ambiguity/multiline protection.

## Remaining implementation tasks

- [ ] Slice 4 (target ≤320 lines) — RED → GREEN → TRIANGULATE → REFACTOR in `internal/edit/{edit.go,edit_test.go}`, `internal/adapter/ghostty/{ghostty.go,ghostty_test.go}`, and `internal/adapter/herdr/{herdr.go,herdr_test.go}`: reject every symlinked ancestor or lexical escape before either adapter reads, writes, or creates an owned artifact; prove external targets and config bytes remain unchanged for `HERDR_CONFIG_PATH` and Ghostty config/theme roots. <!-- sdd-owner: implementation -->
- [ ] Slice 5 (target ≤360 lines) — RED → GREEN → TRIANGULATE → REFACTOR across `go.mod`, `.github/workflows/release.yml`, and the `facundogayoso/homebrew-hather/Formula/hather.rb.in` discovery target: align the Go toolchain used by CI/release/formula, retain vendored offline builds, and add formula/installed-binary acceptance proving a local-image preview/apply remains usable while every selected missing target yields its own actionable unavailable result. <!-- sdd-owner: implementation -->
- [ ] Slice 6 (target ≤120 lines) — reconcile executable evidence and unchecked work across `openspec/changes/hather-macos-theme-manager/{tasks.md,apply-progress.md,verify-report.md}`: correct completed-task claims, verifier counts, evidence revisions, and residual-risk wording so no task or report claims behavior without the corresponding focused proof. <!-- sdd-owner: implementation -->
- [ ] Slice 7 (target ≤120 lines) — run and record final focused regression checks plus `go test ./... -count=1 -timeout 90s`, `go build ./...`, `go vet ./...`, and `git diff --check`; perform the bounded macOS/Homebrew acceptance from `docs/` and update `openspec/changes/hather-macos-theme-manager/verify-report.md` with the observed verdict, failures, and remaining manual limitations. <!-- sdd-owner: implementation -->

---

# Final-verification correction Slice 5 — blocked by stale composition proof

## Status and delivery boundary

Consumed the authoritative `ready` status for `hather-macos-theme-manager`: OpenSpec, strict TDD,
`repo-local` action context, Hather workspace root allowed, and no warnings. The parent resolved the
workload gate as `auto-chain`, `stacked-to-main`; this is Slice 5 only. The declared edit surfaces exclude
`cmd/hather/main_test.go`, so its stale expectation was not changed. No Homebrew install, tap mutation,
network action, user-home action, external application action, commit, tag, push, or release occurred.

## Work completed before the blocker

- Neovim now returns actionable `unavailable` with `config` error class when its config root is absent, before
  creating `colors/` or `hather.vim`; the new focused proof confirms no artifact is created.
- The formula template retains `depends_on "go" => :build`, `-mod=vendor`, and version injection. Its `test do`
  now declares installed-binary local-PNG preview plus empty-`HOME` apply acceptance, asserting partial status and
  independent Ghostty, herdr, Neovim, and VS Code unavailable results.
- Release documentation now states the honest Go contract: `go.mod` sets the Go 1.27 minimum, setup-go reads it,
  and Homebrew supplies an unpinned current `go` formula. It records the formula test as unexecuted acceptance
  contract rather than a completed production or Homebrew rehearsal.

## TDD Cycle Evidence

| Task | Test file | Layer | Safety net | RED | GREEN | TRIANGULATE | REFACTOR |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Missing Neovim target | `internal/adapter/neovim/neovim_test.go` | Unit | `go test ./internal/adapter/neovim -count=1` passed | New missing-config proof failed because `Apply` created `colors/hather.vim` and returned `applied` | The focused proof passed after the pre-create unavailable guard | Existing writable-root test plus the missing-root proof passed together | Minimal direct guard; no abstraction added |
| Formula acceptance manifest | formula `test do` | Static Ruby contract | `ruby -c` passed | A Ruby structural assertion failed because preview/apply and unavailable assertions were absent | Ruby syntax and the structural assertion passed after the manifest | Four explicit target assertions require independent unavailable output | No runnable Homebrew test is authorized in this slice |

## Verification and blocker

- Passed: focused Neovim safety net/RED/GREEN/triangulation; `ruby -c` plus formula structural acceptance check;
  `go build ./...`; `go vet ./...`; `gofmt -d`; Hather and sibling `git diff --check`.
- Blocked: `go test ./... -count=1 -timeout 90s` fails only at
  `cmd/hather.TestRunnerCompositionUsesDefaultNeovimConfigRoot`, which still expects an empty HOME to create
  `.config/nvim/colors/hather.vim` and report `applied`. That contradicts Slice 5's required missing-target
  `unavailable` behavior. Updating this stale composition proof requires explicit permission to edit
  `cmd/hather/main_test.go`; it is outside the user-provided edit surfaces.

## Persisted task status and remaining work

Slice 5 remains visibly unchecked because its full Go evidence is blocked. No task checkbox was changed.
The exact remaining Slice 5 line is:

- [ ] Slice 5 (target ≤360 lines) — RED → GREEN → TRIANGULATE → REFACTOR across `go.mod`, `.github/workflows/release.yml`, and the `facundogayoso/homebrew-hather/Formula/hather.rb.in` discovery target: align the Go toolchain used by CI/release/formula, retain vendored offline builds, and add formula/installed-binary acceptance proving a local-image preview/apply remains usable while every selected missing target yields its own actionable unavailable result. <!-- sdd-owner: implementation -->

The implementation/doc/formula work is below the 360-line slice target, but cannot be completed until the
stale test expectation is reconciled. PR boundary remains this stacked-to-main Slice 5 work unit.

---

# Final-verification correction Slice 5 — composition proof reconciled

## Status and delivery boundary

Consumed the authoritative `ready` status for `hather-macos-theme-manager`: OpenSpec, strict TDD,
`repo-local` action context, and `/Users/facundogayoso/projects/Hather` as the allowed root with no
warnings. The resolved delivery path remains `auto-chain`, `stacked-to-main`; this completes Slice 5
only. The user authorized the previously excluded `cmd/hather/main_test.go` correction. No Homebrew
execution, network, external application or user-home action, commit, or publication occurred.

## Completed task and persisted checkbox

Slice 5 is visibly `- [x]` in `tasks.md`. The stale composition proof now confirms an empty Neovim
config root returns one independent `unavailable` result with the `config` error class, yields
`partial_failure`, and creates no `colors/hather.vim` artifact. This matches the existing missing-target
contract without changing production behavior.

## TDD Cycle Evidence

| Task | Test file | Layer | RED | GREEN | TRIANGULATE | REFACTOR |
| --- | --- | --- | --- | --- | --- | --- |
| Slice 5 missing Neovim composition | `cmd/hather/main_test.go` | Composition | `go test ./cmd/hather -run '^TestRunnerCompositionUsesDefaultNeovimConfigRoot$' -count=1` failed because it expected `applied` and an artifact | `go test ./cmd/hather -run '^TestRunnerCompositionReportsMissingNeovimConfigIndependently$' -count=1` passed after only updating the stale assertion | `go test ./cmd/hather ./internal/adapter/neovim -count=1` passed | No production refactor; the focused proof is the smallest contract-aligned expectation. |

## Verification

- Focused composition and Neovim tests — passed as recorded above.
- `go test ./... -count=1 -timeout 90s` — passed.
- `go build ./...` — passed.
- `go vet ./...` — passed.
- Hather `git diff --check` and sibling tap `git -C /Users/facundogayoso/projects/homebrew-hather diff --check` — passed.
- `ruby -c /Users/facundogayoso/projects/homebrew-hather/Formula/hather.rb.in` — `Syntax OK`.

## Files, deviations, and remaining work

- `cmd/hather/main_test.go` — corrected the stale empty-root expectation and artifact assertion.
- `openspec/changes/hather-macos-theme-manager/{tasks.md,apply-progress.md}` — persisted completion and evidence.

No design deviation and no broadened behavior. The current Hather and sibling tap repositories are entirely
untracked, so their diff checks cannot calculate a patch size; the active Slice 5 implementation remains
within its recorded 360-line target. Formula execution remains a bounded manual Homebrew acceptance for
Slice 7 and was not run.

Remaining implementation tasks:

- [ ] Slice 6 (target ≤120 lines) — reconcile executable evidence and unchecked work across `openspec/changes/hather-macos-theme-manager/{tasks.md,apply-progress.md,verify-report.md}`: correct completed-task claims, verifier counts, evidence revisions, and residual-risk wording so no task or report claims behavior without the corresponding focused proof. <!-- sdd-owner: implementation -->
- [ ] Slice 7 (target ≤120 lines) — run and record final focused regression checks plus `go test ./... -count=1 -timeout 90s`, `go build ./...`, `go vet ./...`, and `git diff --check`; perform the bounded macOS/Homebrew acceptance from `docs/` and update `openspec/changes/hather-macos-theme-manager/verify-report.md` with the observed verdict, failures, and remaining manual limitations. <!-- sdd-owner: implementation -->

PR boundary: stacked-to-main Slice 5 only; no commit or delivery action occurred.

---

# Final-verification correction Slice 6 — evidence reconciliation

Documentation-only reconciliation under the approved `auto-chain` / `stacked-to-main` strategy. Slice 6 is checked; Slice 7 remains unchecked. The initial whole-change verifier recorded FAIL: 7/14 requirements, 27/37 scenarios, seven blocker groups, evidence `sha256:3967d018f9e31aba6d90da69664cd598d18e8eb8de27ccf694d7ea311d25035e`.

Slices 1–5 retain only their recorded evidence: focused proofs and historical full-suite checks for Slices 1 and 3; generic verifier PASS for Slice 2 after two dedicated `sdd-verify` timeouts; independent verifier PASS for Slice 4 at 345 lines after maintainer authorization under the canonical 400-line budget (above its 320-line estimate); and focused Neovim/composition plus static formula evidence for Slice 5. No Homebrew execution followed the formula change. The historical rehearsal's Neovim `applied` result mismatches the later missing-target contract, historical RED cannot be reconstructed without a baseline, and no production delivery occurred. Slice 7 owns final verification; no final counts or PASS are claimed here.

---

# Second-final-verification correction Slice 8 — Wallhaven and CLI attribution

## Status, delivery, and completed task

Consumed authoritative `ready` OpenSpec status for `hather-macos-theme-manager`, `repo-local` action context allowing the Hather root only, no warnings. Strict TDD and the resolved `auto-chain` / `stacked-to-main` path apply. Slice 8 alone is visibly checked in `tasks.md`; Slice 7 and Slices 9–11 remain unchecked. No network, user-home, external-app, commit, or delivery action occurred.

## TDD Cycle Evidence

| Task | Test file / layer | Safety net | RED | GREEN | TRIANGULATE | REFACTOR |
| --- | --- | --- | --- | --- | --- | --- |
| Slice 8 Wallhaven + CLI | `internal/source/wallhaven_test.go`, `internal/cli/cli_test.go` / httptest and CLI fake-core | `go test ./internal/source ./internal/cli ./internal/tui -count=1` passed | Focused new tests failed: download 429 was `download`, empty search lacked recovery guidance, search omitted page attribution | `go test ./internal/source ./internal/cli -count=1` passed after bounded changes | Distinct missing-ID response proves no invented page/attribution; existing search rate policy, keyless/auth, fetch/cache and CLI JSON/local tests passed | Reused shared `renderSource` for search and operation; gofmt and focused/full tests stayed green |

## Files, checks, and boundaries

- `internal/source/{wallhaven.go,wallhaven_test.go}`: download 429 class/retry, empty-search guidance, canonical page link fallback only for known ID, and no attribution when ID/URL absent.
- `internal/cli/{cli.go,cli_test.go}`: search and apply text expose source kind, page, attribution, and available resolution/category/purity; JSON remains untouched.
- `openspec/changes/hather-macos-theme-manager/{tasks.md,apply-progress.md}`: checkbox and cumulative evidence.
- Focused source/CLI/TUI tests, `go test ./... -count=1 -timeout 90s`, `go build ./...`, `go vet ./...`, and `git diff --check` passed. TUI already renders source errors, so its existing path consumes the actionable empty-result error without a new pipeline.
- No design deviation. Source/test changes are approximately 100 authored changed lines, below the 280-line Slice 8 target; Git cannot compute patch size because the repository files are untracked. PR boundary is Slice 8 only, stacked-to-main; no delivery action.

## Remaining implementation tasks

- [ ] Slice 7 (target ≤120 lines) — run and record final focused regression checks plus `go test ./... -count=1 -timeout 90s`, `go build ./...`, `go vet ./...`, and `git diff --check`; perform the bounded macOS/Homebrew acceptance from `docs/` and update `openspec/changes/hather-macos-theme-manager/verify-report.md` with the observed verdict, failures, and remaining manual limitations. <!-- sdd-owner: implementation -->
- [ ] Slice 9 (target ≤320 lines) — RED → GREEN → TRIANGULATE → REFACTOR in `internal/cli/{cli.go,cli_test.go}`, `internal/config/{config.go,config_test.go}`, `cmd/hather/{main.go,main_test.go}`, and `internal/adapter/macos/{macos.go,macos_test.go}`: prove CLI palette output has semantic-role parity with the shared palette, config provenance is reported and configured queries are consumed, and macOS results state the verified propagation limitation without implying all Spaces/apps update instantly; run focused tests and `go test ./...`. <!-- sdd-owner: implementation -->
- [ ] Slice 10 (target ≤180 lines) — RED → GREEN → TRIANGULATE → REFACTOR in `internal/adapter/ghostty/{ghostty.go,ghostty_test.go}`: prove an existing unmarked Ghostty theme selection is user-owned, returns conflict, and causes no write to either config or Hather theme artifact; preserve marked Hather selection behavior and run focused tests plus `go test ./...`. <!-- sdd-owner: implementation -->
- [ ] Slice 11 (target ≤240 lines) — reconcile only observed evidence and residual claims in `openspec/changes/hather-macos-theme-manager/{tasks.md,apply-progress.md,verify-report.md}`, then run focused checks and final `go test ./... -count=1 -timeout 90s`, `go build ./...`, `go vet ./...`, and `git diff --check`; update the verify report with actual outcomes and remaining platform limitations, without checking Slice 7 unless that separately scoped work is completed. <!-- sdd-owner: implementation -->

---

# Reopened Slice 8 — CLI selection metadata and attribution correction

Authoritative OpenSpec status: ready, repo-local Hather root allowed, no action-context warnings. Approved auto-chain/stacked-to-main Slice 8 only; no external actions. Persisted Slice 8 checkbox is now [x].

## TDD Cycle Evidence

| Task / layer | Safety net | RED | GREEN | TRIANGULATE / REFACTOR |
| --- | --- | --- | --- | --- |
| CLI selection metadata and attribution / CLI fake-core | Offline `go test ./internal/cli ./internal/source -count=1` passed | New apply text/JSON metadata test failed with usage exit 4; missing-page test emitted `attribution: Wallhaven: ` | Focused new tests passed after three explicit selection flags and conditional attribution | Known-page text/JSON and missing-page text/JSON exercise both branches; existing source tests cover known-ID canonical fallback. No abstraction needed; gofmt and focused tests passed. |

## Checks, scope, and remaining work

`go test ./internal/cli ./internal/source -count=1`, `go test ./... -count=1 -timeout 90s`, `go build ./...`, `go vet ./...`, and `git diff --check` passed with GOPROXY/GOSUMDB off for Go commands. Changed only `internal/cli/{cli.go,cli_test.go}` plus tasks/progress artifacts. The existing source fallback and 429/empty-result behavior were not changed. Authored CLI additions are under 100 lines, below the 280-line Slice 8 target (repository remains untracked, so Git cannot measure a patch). No design deviation. PR boundary is Slice 8 only; no commit or delivery.

Remaining exact unchecked implementation rows are Slice 7 and Slices 9–11 as recorded in `tasks.md`; Slice 7 manual macOS/Homebrew acceptance remains unperformed in this correction.

---

# Second-final-verification correction Slice 9 — palette, ordinary config, macOS

Authoritative status: ready, OpenSpec, repo-local Hather edit root, no action-context warnings. Approved auto-chain / stacked-to-main Slice 9 only. Its persisted checkbox is [x]. No network, live TCC/macOS, home, commit or delivery actions.

## TDD Cycle Evidence

| Task / layer | Safety net | RED | GREEN | TRIANGULATE / REFACTOR |
| --- | --- | --- | --- | --- |
| Slice 9 / CLI/config/composition/macOS unit and fake executor | Focused `go test ./internal/cli ./internal/config ./cmd/hather ./internal/tui ./internal/adapter/macos -count=1` passed | New focused run failed on missing config source resolver and CLI fields and macOS propagation assertion | Focused run passed after smallest implementation | Preview and apply each show all four palette roles; explicit query beats configured query; flag/environment/config/default source cases and TUI initial environment query covered. Existing TUI user-input override passes. gofmt and focused tests remained green; no further refactor needed. |

## Changed files and verification

`internal/cli/{cli.go,cli_test.go}`, `internal/config/{config.go,config_test.go}`, `cmd/hather/{main.go,main_test.go}`, `internal/tui/model_test.go`, `internal/adapter/macos/{macos.go,macos_test.go}`, `tasks.md`, and this cumulative progress. CLI text displays semantic roles; JSON model remains unchanged. Ordinary query uses HATHER_QUERY then config; TUI starts with resolved query but accepts user input. Search text and apply text report ordinary-setting source without printing any credential; macOS success warns that Spaces and running apps may propagate later. No design deviation; provenance is a text diagnostic seam rather than new persistence.

Focused tests, `go test ./... -count=1 -timeout 90s`, `go build ./...`, `go vet ./...`, and `git diff --check` passed with GOPROXY/GOSUMDB off on Go commands. Approximately 130 authored source/test lines, below the 320-line target; repository remains untracked so Git cannot calculate a slice diff. PR boundary: Slice 9 only; no delivery action. macOS propagation is a reported limitation, not live acceptance proof.

## Remaining implementation tasks (exact unchecked rows)

- [ ] Slice 7 (target ≤120 lines) — run and record final focused regression checks plus `go test ./... -count=1 -timeout 90s`, `go build ./...`, `go vet ./...`, and `git diff --check`; perform the bounded macOS/Homebrew acceptance from `docs/` and update `openspec/changes/hather-macos-theme-manager/verify-report.md` with the observed verdict, failures, and remaining manual limitations. <!-- sdd-owner: implementation -->
 - [ ] Slice 10 (target ≤180 lines) — RED → GREEN → TRIANGULATE → REFACTOR in `internal/adapter/ghostty/{ghostty.go,ghostty_test.go}`: prove an existing unmarked Ghostty theme selection is user-owned, returns conflict, and causes no write to either config or Hather theme artifact; preserve marked Hather selection behavior and run focused tests plus `go test ./...`. <!-- sdd-owner: implementation -->
 - [ ] Slice 11 (target ≤240 lines) — reconcile only observed evidence and residual claims in `openspec/changes/hather-macos-theme-manager/{tasks.md,apply-progress.md,verify-report.md}`, then run focused checks and final `go test ./... -count=1 -timeout 90s`, `go build ./...`, `go vet ./...`, and `git diff --check`; update the verify report with actual outcomes and remaining platform limitations, without checking Slice 7 unless that separately scoped work is completed. <!-- sdd-owner: implementation -->

---

# Reopened Slice 9 — TUI empty-query fallback

Authoritative OpenSpec status was ready, repo-local Hather root allowed, no warnings. Approved auto-chain/stacked-to-main Slice 9 correction only; persisted Slice 9 checkbox is [x].

## TDD Cycle Evidence

| Safety net | RED | GREEN | TRIANGULATE / REFACTOR |
| --- | --- | --- | --- |
| `go test ./internal/tui ./cmd/hather -count=1` passed | New TUI regression failed to compile on absent DefaultQuery | Focused regression passed after empty-input fallback in search command | Cleared preloaded input, return via `s`, and nonempty ocean override pass; gofmt and focused tests pass. |

## Files, checks, and boundary

`internal/tui/{model.go,model_test.go}` and `cmd/hather/{main.go,main_test.go}` preserve the configured query separately from mutable input, without changing local-path behavior. `tasks.md` and this cumulative progress record completion. Focused TUI/main tests, `go test ./... -count=1 -timeout 90s`, `go build ./...`, `go vet ./...`, and `git diff --check` passed with Go proxy/sumdb disabled. No design deviation. Approximately 50 authored Go lines, below Slice 9 target and 400-line review budget; repository files are untracked, so Git cannot measure patch size. No network, live platform, user-home, commit or delivery action.

Remaining unchecked implementation rows: Slice 7, Slice 10 and Slice 11 exactly as persisted in `tasks.md`; Slice 7 macOS/Homebrew acceptance remains unperformed.

## Exact remaining task lines

- [ ] Slice 7 (target ≤120 lines) — run and record final focused regression checks plus `go test ./... -count=1 -timeout 90s`, `go build ./...`, `go vet ./...`, and `git diff --check`; perform the bounded macOS/Homebrew acceptance from `docs/` and update `openspec/changes/hather-macos-theme-manager/verify-report.md` with the observed verdict, failures, and remaining manual limitations. <!-- sdd-owner: implementation -->
 - [ ] Slice 10 (target ≤180 lines) — RED → GREEN → TRIANGULATE → REFACTOR in `internal/adapter/ghostty/{ghostty.go,ghostty_test.go}`: prove an existing unmarked Ghostty theme selection is user-owned, returns conflict, and causes no write to either config or Hather theme artifact; preserve marked Hather selection behavior and run focused tests plus `go test ./...`. <!-- sdd-owner: implementation -->
 - [ ] Slice 11 (target ≤240 lines) — reconcile only observed evidence and residual claims in `openspec/changes/hather-macos-theme-manager/{tasks.md,apply-progress.md,verify-report.md}`, then run focused checks and final `go test ./... -count=1 -timeout 90s`, `go build ./...`, `go vet ./...`, and `git diff --check`; update the verify report with actual outcomes and remaining platform limitations, without checking Slice 7 unless that separately scoped work is completed. <!-- sdd-owner: implementation -->
---

# Second-final-verification Slice 10 — Ghostty artifact ownership

Authoritative status: ready, OpenSpec, repo-local Hather edit root, no warnings. Strict TDD; approved auto-chain/stacked-to-main Slice 10 only. The Slice 10 task checkbox is visibly `[x]` in `tasks.md`. No commit, delivery, real home, Ghostty process, or network action.

## TDD Cycle Evidence

| Task / test layer | Safety net | RED | GREEN | TRIANGULATE | REFACTOR |
| --- | --- | --- | --- | --- | --- |
| Ghostty artifact ownership / temp-dir adapter integration (`internal/adapter/ghostty/ghostty_test.go`) | `go test ./internal/adapter/ghostty -count=1` passed | New existing-unmarked-artifact proof failed: marked selection overwrote user theme and modified config | Header guard before directory creation or either write made the focused RED test pass | Unmarked selection and marked selection with unowned artifact both conflict with byte-identical files; owned artifact with marked selection applies twice with deterministic output and preserved config | Shared stable generated header constant; focused package tests stayed green |

## Files and verification

- `internal/adapter/ghostty/ghostty.go`: exact generated ownership header required before replacing existing `themes/hather.conf`; conflict occurs before config/theme writes. Existing config preflight, rollback, and path confinement remain unchanged.
- `internal/adapter/ghostty/ghostty_test.go`: two conflict cases assert both config and theme bytes unchanged; owned marked artifact updates repeatably. Existing tests cover first creation, rejected symlinks/escapes, and rejected config edits.
- `tasks.md`, this cumulative `apply-progress.md`: checkbox and evidence. The checked-in theme fixture already has the stable header; it was not modified.
- Focused Ghostty tests, `GOPROXY=off GOSUMDB=off go test ./... -count=1 -timeout 90s`, offline `go build ./...`, offline `go vet ./...`, `gofmt`, and `git diff --check` passed. Git reports no tracked patch because this repository's files are untracked; authored Go edits are approximately 75 lines, below 180.

No design deviation. PR boundary: Slice 10 only; no release acceptance or verifier reconciliation performed. Remaining exact unchecked implementation rows (historical unchecked lists above describe their own earlier snapshots):

- [ ] Slice 7 (target ≤120 lines) — run and record final focused regression checks plus `go test ./... -count=1 -timeout 90s`, `go build ./...`, `go vet ./...`, and `git diff --check`; perform the bounded macOS/Homebrew acceptance from `docs/` and update `openspec/changes/hather-macos-theme-manager/verify-report.md` with the observed verdict, failures, and remaining manual limitations. <!-- sdd-owner: implementation -->
- [ ] Slice 11 (target ≤240 lines) — reconcile only observed evidence and residual claims in `openspec/changes/hather-macos-theme-manager/{tasks.md,apply-progress.md,verify-report.md}`, then run focused checks and final `go test ./... -count=1 -timeout 90s`, `go build ./...`, `go vet ./...`, and `git diff --check`; update the verify report with actual outcomes and remaining platform limitations, without checking Slice 7 unless that separately scoped work is completed. <!-- sdd-owner: implementation -->
---

# Final certification — Slices 7 and 11

Final independent verdict: **PASS WITH WARNINGS**, 14/14 requirements, 37/37 scenarios, 0 blockers. Full Go tests (16 packages), build, vet, Hather/tap diff checks, workflow YAML parse, and formula Ruby syntax passed. The exact-current-candidate temporary Homebrew rehearsal passed vendored tests, install, formula test, installed help/version/preview/apply, and cleanup. Evidence: archive `541d001ab20b7c14f209da9d25bb92971cc604efe261cd517908bb2f8b57bd11`; vendored tests `a03c0063632b8276ca9b97f93679e9f4cb66b8722a793cd150a32ccbd3a4ea49`; install `04e381ddc6bca2216bd6ada16e2e3244be59fe30e5e5ff33ed0be255501a4891`; brew test `c336acbd8147f7104a040fa20a267bc032d0bef35be724e8c135223dfcff375f`. Formula, tap, and workspace cleanup were verified absent.

Wallhaven authentication now uses the observed URL-encoded `apikey` query parameter; keyless and image-download requests omit credentials. Remaining limitations are non-blocking: both repositories have no tracked `HEAD`, live TCC denial remains fixture-only, and no production tag, GitHub release, push, or public tap publication occurred. Slices 7 and 11 are checked in `tasks.md`; local archive readiness does not authorize delivery.

During final evidence persistence, this file was accidentally overwritten. Its exact 1,228-line pre-write content was restored from the active attempt's immutable begin candidate tree `95dd8b9f38d9d33ca7b6487cbd36e3541ec618c4` (restored SHA-256 `365c205e6dafd4f902aedc08b57523962ee375feae40f19e3746c38a61e146b9`) before this section was appended.
