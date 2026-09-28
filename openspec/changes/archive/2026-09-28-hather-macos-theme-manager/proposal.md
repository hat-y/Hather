# Proposal — Hather macOS Theme Manager

## Intent

Build the first usable slice of Hather, a standalone open-source macOS TUI and CLI that turns a selected Wallhaven wallpaper into a cohesive palette and applies it through theme adapters.

The product pipeline is deliberately small:

```text
Wallhaven wallpaper source → palette engine → theme adapters
```

Hather is the successor product to the workflow represented by HatDots, but it is not an extension of HatDots. This change does not modify HatDots or depend on its files at runtime.

## Product outcome

A user can install Hather through Homebrew, select or provide a Wallhaven wallpaper, preview the derived palette, and apply that palette to the integrations they have enabled. The same application core serves both scripted CLI use and interactive TUI use.

The experience should make one coherent theme available across the user's desktop, terminal, multiplexer, editor, and IDE without requiring each integration to be configured by hand. Integrations are optional and independent: an absent application, unsupported configuration, or permission failure is reported clearly and does not discard the palette or prevent other adapters from running.

## User journeys

### Interactive selection and apply

1. The user launches the TUI and searches or browses Wallhaven results.
2. Hather displays enough wallpaper metadata and a palette preview to choose an image.
3. The user selects enabled adapters and applies the result.
4. Hather reports per-adapter outcomes, including any permission, missing-application, restart, or write failure.
5. The user can continue using the successfully updated integrations even if one adapter failed.

### Scripted apply

1. The user invokes the CLI with a wallpaper source/query and an explicit or configured adapter set.
2. Hather downloads or reuses the wallpaper, derives a palette, and applies adapters in a deterministic order.
3. The CLI exits with a result that distinguishes complete success from partial adapter failure, while retaining useful output and diagnostics for automation.

### Local wallpaper apply

1. The user supplies a local image instead of searching Wallhaven.
2. Hather derives a palette without requiring a Wallhaven API key or network access.
3. The selected adapters are applied using the same core path as Wallhaven-sourced images.

## Scope

### MVP capabilities

- Go-based macOS application core shared by CLI and TUI.
- Wallhaven as the first remote wallpaper source.
- Local image input as a useful offline/apply seam where it keeps the core simple.
- Palette preview and persisted last-applied result under Hather-owned state paths.
- Initial adapter set, each independently enabled:
  - macOS wallpaper and light/dark appearance.
  - Ghostty.
  - herdr.
  - Neovim.
  - VS Code.
- TUI for selecting a wallpaper, previewing a palette, choosing adapters, applying, and viewing per-adapter results.
- CLI for the same core operations and for automation.
- Homebrew-first installation, initially through an owned tap or equivalent release formula path.
- Clear attribution and rate-limit behavior for Wallhaven.

### Adapter behavior in the first product slice

- **macOS:** use the proven System Events/Apple Events path for wallpaper and light/dark appearance, with explicit Automation-permission errors. Treat all-desktops wallpaper application as the default. Per-Space targeting is not promised until verified on the supported macOS versions.
- **Ghostty:** prefer a generated Hather theme file and the smallest possible config change needed to select it. Preserve unrelated configuration and resolve conflicts with existing inline palettes during spec/design.
- **herdr:** update only the owned theme portions of the TOML configuration, preserving shell, keybind, server, and unrelated settings. Restart or reload guidance is reported rather than implied to be instantaneous.
- **Neovim:** generate an additive `hather` colorscheme file. Write-only or opt-in headless/live application is the conservative default until a server-socket contract is explicitly configured.
- **VS Code:** update only the supported color customization sections in `settings.json`; do not assume VS Code is installed or running, and do not promise full token-theme parity in the MVP.

## MVP boundaries and non-goals

### Boundaries

- macOS is the only target platform for the initial release.
- Wallhaven is the only remote source in this change.
- Light/dark appearance is supported; Tinted appearance, automatic scheduling preservation, and per-app appearance control are later decisions.
- Adapter output is intentionally minimal and additive. Each adapter owns only the files or keys needed for its result.
- A Wallhaven API key is optional. Public search and local images must remain usable without one.
- The initial release prioritizes reliable application and clear partial results over a large theme-template catalog.

### Non-goals

- No HatDots changes, imports, symlinks, or runtime dependency.
- No cross-platform support, GUI application, menu-bar application, daemon, or background wallpaper rotation.
- No promise that every macOS Space or every running application updates instantly.
- No requirement that Ghostty, herdr, Neovim, or VS Code be installed or running.
- No destructive replacement of complete user configuration files.
- No full VS Code semantic-token theme, broad editor/plugin theme marketplace, or generated theme gallery.
- No mandatory Peachy installation or hard runtime dependency on a separate Peachy binary.
- No automatic API-key plaintext storage in Hather configuration.
- No homebrew-core submission as a prerequisite for the first release; an owned tap is sufficient for the initial distribution slice.

## Proposed architecture

Keep one small application core with explicit seams:

```text
source → wallpaper artifact → palette engine → palette
                                      ↓
                              adapter runner → results
```

- **Sources** return wallpaper metadata and a retrievable image artifact. Wallhaven-specific query and authentication behavior stays behind this seam.
- **Palette engine** accepts an image artifact and returns a versioned, adapter-neutral palette. The engine must be usable without an external Peachy installation.
- **Adapters** accept the palette plus a narrow context and return a structured result. They must not call one another or mutate shared adapter state.
- **Adapter runner** executes enabled adapters, records each result, and continues after an isolated failure. It owns ordering, cancellation, summary, and CLI/TUI presentation—not adapter-specific configuration logic.
- **Persistence** is Hather-owned: configuration in `~/.config/hather`, state/cache under `~/.local/state/hather` and an appropriate cache/data location. Exact paths and retention rules belong in the spec.

Use standard-library HTTP, image decoding where sufficient, file handling, and CLI parsing first. Bubble Tea/lipgloss may provide the TUI, but package selection should remain limited to what the TUI actually needs. Avoid adding a CLI framework solely for conventional subcommands.

All remote and platform behavior discovered during exploration remains subject to bounded verification in spec/design. This proposal does not claim external documentation or API behavior has been independently re-verified.

## Conservative proposal decisions

These defaults narrow risk without locking the implementation to speculative infrastructure:

1. **Adapter isolation:** an adapter failure is a per-adapter result, not a pipeline-wide abort, except when the shared wallpaper or palette prerequisite fails. Writes are additive and targeted; unrelated configuration is preserved.
2. **Wallhaven authentication:** API keys are optional and must not be required for MVP use. The first spec should prefer an environment/explicit-input path and defer persisted credential storage; if persistence is later required, use the macOS Keychain rather than plaintext config.
3. **Peachy/native palette work:** run a bounded compatibility spike against the observed Peachy contract. Keep a `PaletteEngine` seam either way. If reuse is not small, stable, and distributable through Homebrew, use a minimal native engine rather than making users install Peachy.
4. **macOS apply:** use System Events for MVP wallpaper and light/dark appearance. Detect and explain Apple Events denial. Do not claim instant propagation to every Space or application.
5. **Ghostty changes:** generate a Hather-owned theme artifact and make the minimum targeted selection change. Do not rewrite the user's complete config or blindly replace an inline palette; the exact precedence/merge rule must be proven before implementation.
6. **Neovim changes:** generate an additive colorscheme file first. Live application requires an explicit, verified server-socket path and is not a prerequisite for MVP success.
7. **herdr and VS Code:** report restart/window-reload requirements as adapter results or instructions rather than hiding them behind a false synchronous apply contract.
8. **Delivery:** implement in reviewable slices, each planned below the 400 changed-line budget. Do not combine all adapters, the TUI, and packaging into one review unit.

## Decisions required for spec/design

The following are product or contract decisions, not implementation details to silently infer:

- The exact supported macOS version range and whether the first release supports only the environment used during exploration.
- Wallhaven query defaults, SFW/purity policy, attribution text, download/cache retention, and behavior when the API is unavailable or rate-limited.
- Palette format: number of colors, semantic roles (background/foreground/accent), contrast expectations, and deterministic output guarantees.
- The result and exit-code contract for partial adapter failure.
- Config precedence and credential input precedence (`flag`, environment, config, and future Keychain).
- Ghostty conflict behavior when an existing inline palette or theme setting is present, including whether Hather may add or replace one owned setting.
- herdr reload/restart behavior and the exact `[theme.custom]` keys that are safe to own.
- Neovim default apply behavior: write-only, next-start headless apply, or explicitly opted-in server apply.
- VS Code settings merge policy and the minimum supported customization keys.
- Whether a user can choose “all adapters” or must explicitly enable each adapter on first run.
- Whether macOS appearance changes should preserve and restore automatic appearance scheduling state.
- The exact Homebrew first-release route (owned tap naming, release artifact, and formula test) before packaging work begins.

These decisions should be resolved from bounded local/platform verification and product preference during spec/design. No research phase is created for this proposal.

## Initial phased delivery strategy

Each phase is an independently reviewable slice intended to stay below the 400-line review budget. The phases describe implementation order, not a promise that all work belongs in one pull request. The ask-on-risk delivery policy remains in force if a slice approaches the budget.

### Phase 0 — Contract spikes and test fixtures

- Confirm the Wallhaven response/error fixtures needed by the source seam.
- Exercise the Peachy compatibility contract without making Peachy a dependency.
- Verify the narrow platform behaviors that affect contracts: Apple Events errors, Ghostty theme/config precedence, herdr theme keys/reload, and safe JSON/TOML edits.
- Produce fixture-based tests and decisions for the following phase; do not add production breadth merely to close an evidence gap.

### Phase 1 — Core pipeline and CLI foundation

- Establish the Go module, source/palette/adapter interfaces, Hather-owned config/state paths, and deterministic result types.
- Implement Wallhaven retrieval with optional authentication, throttling/error classification, and local-image input if it does not expand the seam.
- Implement the selected palette-engine path behind the engine seam.
- Add a thin CLI path that can run the pipeline with no adapters or a test adapter, using offline fixtures for tests.

### Phase 2 — macOS and Ghostty application

- Add the macOS wallpaper/appearance adapter with permission-aware errors and conservative all-desktops semantics.
- Add the Ghostty theme-file/config adapter with targeted, additive edits.
- Add adapter-runner reporting and focused behavior proofs for successful, missing-target, permission, and write-failure cases.
- Keep the TUI out of this slice unless the review size remains demonstrably below budget.

### Phase 3 — herdr, Neovim, and VS Code adapters

- Add each remaining selected adapter independently, with fixture-based config merge tests.
- Preserve unrelated TOML, Lua/Vim, and JSON settings.
- Return explicit follow-up instructions for restart, next-start, or window reload behavior.
- Verify that one failed adapter leaves successful adapter results and generated artifacts intact.

### Phase 4 — TUI and user-facing completion

- Build the smallest TUI over the already-tested core: source selection, palette preview, adapter selection, apply, and per-adapter results.
- Keep TUI state and rendering separate from source, palette, and adapter behavior.
- Add accessibility-minded keyboard paths and clear empty/error states without adding a second application core.

### Phase 5 — Homebrew-first distribution and release hardening

- Add versioned release packaging through an owned tap or equivalent initial formula path.
- Verify install, `--help`, version output, offline/local-image behavior, and a bounded end-to-end apply rehearsal on macOS.
- Document Wallhaven attribution, optional key handling, Automation permission remediation, adapter latency, supported macOS scope, and known non-goals.
- Consider homebrew-core only after the first release is stable; it is outside the MVP delivery gate.

## Success criteria

The change is successful when all of the following are true:

1. A fresh macOS user can install Hather through the chosen initial Homebrew path and discover CLI help and version information.
2. A Wallhaven search and download can complete without an API key under the supported public-use path, with rate-limit and attribution behavior documented.
3. A local image can exercise the palette pipeline without network access.
4. The TUI and CLI use the same tested pipeline and produce the same palette and adapter result model for equivalent inputs.
5. The selected adapters generate or update only their owned artifacts/keys, preserve unrelated user configuration, and are safe when their target application is absent.
6. Adapter failures are isolated, clearly reported, and do not erase a successful palette or successful sibling-adapter results.
7. macOS Automation denial produces actionable remediation text instead of a silent failure.
8. Apply behavior is documented per adapter, including whether it is immediate, hot-reloaded, restart-based, next-start, or manual.
9. Offline fixture tests cover source parsing, palette determinism, config-preserving writes, and partial-failure aggregation; implementation phases retain strict TDD as configured by the project.
10. Each implementation slice remains within the 400-line review budget or pauses for the configured ask-on-risk delivery decision.

## Risks and mitigations

| Risk | Mitigation |
|---|---|
| Wallhaven API behavior, limits, or terms differ from assumptions | Verify the endpoint and response fixtures before implementation; throttle, preserve optional-key operation, document attribution, and classify failures. |
| Apple Events/TCC blocks macOS application | Detect the specific denial, explain the System Settings remediation, and continue reporting other adapter results where possible. |
| macOS Spaces or running apps do not update uniformly | Make all-desktops behavior the only initial promise, verify targeted behavior separately, and document per-adapter/per-surface latency. |
| Existing config formats are personalized or malformed | Use targeted parsers/merges, preserve unknown fields where the format permits, write owned artifacts separately, and fail without clobbering on unsafe input. |
| Peachy is unavailable or its export contract changes | Keep the engine abstraction, time-box reuse validation, and retain a native fallback path with no mandatory Peachy installation. |
| A missing or stopped target application makes apply appear broken | Treat absence and reload requirements as structured adapter outcomes with actionable instructions. |
| All five adapters plus TUI exceed the review budget | Use the explicit phase slices and stop for ask-on-risk rather than compressing review or hiding work in abstractions. |
| Credentials leak through config or logs | Keep the key optional, avoid plaintext persistence in MVP, redact diagnostics, and defer Keychain persistence until its contract is explicit. |

## Rollback and recovery

- **Code delivery:** revert the individual phase slice; no phase should require a later phase to remain buildable or testable.
- **Generated artifacts:** write Hather-owned theme files with stable names and metadata so they can be removed or regenerated without touching unrelated files.
- **Config edits:** make narrow, identifiable changes and retain enough prior owned-value information to disable Hather's selection. Never implement rollback by deleting an entire Ghostty, herdr, Neovim, or VS Code configuration.
- **Runtime apply:** if an adapter cannot safely prove its prior state or restore semantics, report that limitation and require explicit user action rather than pretending rollback succeeded. In particular, automatic restoration of every macOS Space wallpaper is not an MVP guarantee.
- **Distribution:** removing the Homebrew formula removes Hather; it does not remove user-generated configuration unless the user explicitly requests cleanup.

## Out of scope for this proposal

This artifact defines intent, boundaries, decisions, and delivery sequencing only. It does not create application code, package manifests, adapter implementations, external research claims, or HatDots changes. Those belong to later spec/design and implementation phases.
