# Hather macOS Theme Manager Specification

## Purpose

Define the first standalone Hather release: a macOS CLI and TUI that turns a Wallhaven or local wallpaper into a deterministic palette and applies it independently to macOS, Ghostty, herdr, Neovim, and VS Code. Hather MUST NOT modify, import, symlink to, or depend at runtime on HatDots.

## Requirements

### Requirement: Shared pipeline and result contract

Hather MUST expose one application-core pipeline for both CLI and TUI use. The pipeline MUST represent a wallpaper source, retrieved wallpaper artifact, versioned adapter-neutral palette, requested adapters, and per-adapter results. A result MUST distinguish shared prerequisite failure from adapter failure and MUST retain successful results and generated artifact paths when another adapter fails.

#### Scenario: Equivalent CLI and TUI inputs share the same outcome model

- GIVEN the CLI and TUI select the same wallpaper artifact, palette options, and enabled adapters
- WHEN each runs the apply pipeline
- THEN both use the same palette and per-adapter result semantics
- AND neither surface implements a second adapter or palette pipeline

#### Scenario: Shared prerequisite failure prevents unsafe adapter execution

- GIVEN the wallpaper cannot be read or the palette cannot be produced
- WHEN an apply operation starts
- THEN the operation reports a shared prerequisite failure
- AND no adapter is reported as successfully applied
- AND the failure includes actionable diagnostics

#### Scenario: Adapter failure does not erase sibling results

- GIVEN multiple adapters are enabled and one adapter fails
- WHEN the runner applies them
- THEN every adapter receives an individual result
- AND successful sibling results and generated artifacts remain available
- AND the aggregate result identifies the operation as partial failure

### Requirement: Wallhaven and local wallpaper sources

Hather MUST support Wallhaven search/download and local image input through the same source contract. Wallhaven API-key authentication MUST be optional. Local-image processing MUST work without network access or a Wallhaven key. Remote failures MUST be classified sufficiently to distinguish invalid input, unavailable service, rate limiting, authentication failure, and download failure.

#### Scenario: Public Wallhaven use works without a key

- GIVEN no Wallhaven API key is configured
- WHEN the user performs a permitted public search and selects an available image
- THEN Hather can retrieve the image through the supported public path
- AND the result records the wallpaper metadata needed for attribution

#### Scenario: Local input bypasses Wallhaven

- GIVEN the user supplies a readable supported local image
- WHEN the user previews or applies it
- THEN Hather derives the palette without contacting Wallhaven
- AND the same downstream adapters can be selected as for a remote image

#### Scenario: Remote rate limiting is actionable

- GIVEN Wallhaven reports a rate limit or Hather's request policy prevents another request
- WHEN the user requests a remote operation
- THEN Hather returns a rate-limit result with retry guidance
- AND it does not silently retry without respecting the applicable limit

### Requirement: Deterministic palette and contrast

The palette engine MUST return a versioned, adapter-neutral palette with stable semantic roles for background, foreground, muted text, and accent colors. Given the same image bytes and palette-engine version, output MUST be deterministic across CLI and TUI runs. The palette MUST provide readable foreground/background contrast for primary text and MUST record when a source cannot meet the target contrast without adjustment.

#### Scenario: Repeated extraction is stable

- GIVEN identical image bytes and the same palette-engine version
- WHEN the palette is generated repeatedly
- THEN role values and palette ordering are identical
- AND the CLI and TUI previews show the same values

#### Scenario: Contrast is enforced or reported

- GIVEN an image whose dominant colors produce insufficient text contrast
- WHEN Hather generates a palette
- THEN it adjusts eligible foreground or muted roles to meet the defined readability threshold
- OR returns an explicit contrast limitation in palette diagnostics
- AND adapters receive the same final palette values

#### Scenario: Palette state is reproducible

- GIVEN a successfully generated palette
- WHEN Hather persists the last-applied result
- THEN the stored record includes the palette version, source identity, and palette roles
- AND the record can be used to explain or reproduce the applied result

### Requirement: Hather-owned configuration and precedence

Hather MUST keep user configuration and state under Hather-owned paths, including `~/.config/hather/` for configuration, `~/.local/state/hather/` for last-applied state and operational records, and a Hather-owned cache/data location for downloaded wallpapers. Exact file retention MUST be documented. For ordinary settings, explicit CLI input MUST override environment values, which MUST override config-file values, which MUST override defaults. Wallhaven credentials MUST accept explicit input and environment input, MUST be redacted from diagnostics, and MUST NOT be written to plaintext Hather configuration in the MVP.

#### Scenario: Configuration precedence is predictable

- GIVEN a setting exists as a default, config value, environment value, and explicit CLI value
- WHEN the operation resolves configuration
- THEN the explicit CLI value wins, followed by the environment value, config value, and default
- AND the resolved source is visible without exposing secrets

#### Scenario: Credentials are not persisted in plaintext

- GIVEN a Wallhaven key is supplied through a flag or environment variable
- WHEN Hather performs a remote operation
- THEN the key is used for that operation
- AND it is absent from persisted plaintext config, normal output, and error diagnostics

### Requirement: macOS wallpaper and appearance adapter

The macOS adapter MUST support applying a wallpaper to all desktops as its MVP default and setting system light or dark appearance. It MUST not promise per-Space targeting, tinted appearance, automatic scheduling preservation, or instant propagation beyond what the supported platform behavior verifies. Apple Events/Automation denial MUST be reported as a distinct actionable permission result with remediation guidance. Appearance scheduling state MUST be preserved unless the user explicitly requests a permanent override.

#### Scenario: Default wallpaper application targets all desktops

- GIVEN the macOS adapter is enabled and the image is valid
- WHEN the adapter applies the wallpaper with default scope
- THEN it requests all-desktop application
- AND the result states the promised scope and any platform propagation limitation

#### Scenario: Automation denial is actionable

- GIVEN macOS denies Hather permission to send Apple Events to System Events
- WHEN the adapter attempts wallpaper or appearance application
- THEN the result identifies an Automation permission failure
- AND it directs the user to System Settings → Privacy & Security → Automation
- AND it does not prevent non-macOS adapters from producing their results

#### Scenario: Automatic appearance scheduling is not silently lost

- GIVEN automatic light/dark scheduling is enabled before an explicit appearance apply
- WHEN Hather changes appearance
- THEN it preserves and reports the scheduling state according to the selected MVP policy
- OR refuses the change with an explicit explanation rather than silently disabling scheduling

### Requirement: Additive Ghostty theme handling

The Ghostty adapter MUST generate a Hather-owned theme artifact with stable identity and MUST make only the smallest targeted configuration change needed to select it. It MUST preserve unrelated Ghostty settings and MUST not blindly replace an existing inline palette. When an existing theme or inline palette creates an ownership conflict that cannot be resolved safely, the adapter MUST report the conflict without clobbering the config.

#### Scenario: Theme artifact and selection are additive

- GIVEN a writable Ghostty configuration and themes directory
- WHEN the Ghostty adapter applies a palette
- THEN it writes or updates only Hather's owned theme artifact
- AND it changes only the targeted theme selection
- AND unrelated configuration remains unchanged

#### Scenario: Existing inline palette is protected

- GIVEN Ghostty uses an inline palette and no safely proven Hather ownership exists
- WHEN the adapter applies a palette
- THEN it does not blindly delete or rewrite the inline palette
- AND it either records a safe, targeted selection change or returns an actionable conflict result

#### Scenario: Missing Ghostty is non-fatal

- GIVEN Ghostty or its configuration directory is absent
- WHEN the adapter runs
- THEN it returns a clear skipped or unavailable result
- AND other enabled adapters continue

### Requirement: herdr TOML ownership and reload guidance

The herdr adapter MUST honor `HERDR_CONFIG_PATH` when set and otherwise use the documented user configuration path. It MUST own only the theme selection and explicitly supported custom theme keys, including the safe initial `[theme.custom]` color keys verified during implementation. It MUST preserve shell, keybind, server, unknown, and unrelated TOML settings. It MUST report whether restart or reload is required and MUST not claim synchronous application when that behavior is unverified.

#### Scenario: Theme-only TOML update preserves unrelated settings

- GIVEN a valid herdr TOML file containing personalized non-theme settings
- WHEN the adapter applies a palette
- THEN only Hather-owned theme values change
- AND shell, keybind, server, and unrelated values remain intact

#### Scenario: Unsafe or malformed TOML is not clobbered

- GIVEN the selected herdr configuration cannot be safely parsed or written
- WHEN the adapter runs
- THEN it returns a write/configuration failure
- AND it leaves the original file unchanged
- AND it gives the user a remediation path

#### Scenario: Reload semantics are explicit

- GIVEN herdr requires a restart or reload for configuration changes
- WHEN the adapter completes its file update
- THEN the result states the required follow-up action and expected apply timing
- AND it does not imply that running herdr instances changed immediately unless verified

### Requirement: Additive Neovim colorscheme

The Neovim adapter MUST generate an additive `hather` colorscheme artifact in the user's configured Neovim colors directory or documented default. MVP apply MUST be write-only; live server application MUST NOT be attempted unless the user explicitly configures and opts into a verified server socket. The adapter MUST preserve existing plugins and colorscheme configuration and MUST report the next-start or manual apply instruction.

#### Scenario: Colorscheme generation is non-destructive

- GIVEN a writable Neovim colors directory
- WHEN the adapter applies a palette
- THEN it creates or updates only the Hather-owned `hather` colorscheme
- AND existing Neovim configuration and plugin files remain unchanged

#### Scenario: Default apply does not assume a server

- GIVEN no explicit verified Neovim server socket is configured
- WHEN the adapter runs
- THEN it writes the colorscheme without attempting remote application
- AND the result explains how to select it on the next start or manually

### Requirement: VS Code settings merge

The VS Code adapter MUST update only supported color customization sections in the user settings file, preserving unrelated settings and formatting/data that can be safely preserved. MVP output MUST be limited to the documented minimal workbench/editor color customization surface and MUST NOT claim full semantic-token or installed-extension theme parity. Missing VS Code installations and window-reload requirements MUST be reported without failing the pipeline prerequisite.

#### Scenario: Settings merge preserves unrelated values

- GIVEN a valid VS Code `settings.json` with unrelated user settings
- WHEN the adapter applies a palette
- THEN only Hather-owned supported color customization entries change
- AND unrelated settings remain unchanged

#### Scenario: Invalid settings are protected

- GIVEN the settings file cannot be safely parsed or written
- WHEN the adapter runs
- THEN it returns an actionable configuration failure
- AND it does not replace the complete settings file

#### Scenario: Reload behavior is honest

- GIVEN VS Code may require a window reload to reflect changed colors
- WHEN the adapter writes settings
- THEN the result states whether reload or reopening is recommended
- AND it does not claim that every running window updates immediately

### Requirement: Adapter isolation and partial-failure aggregation

Adapters MUST be independently enabled, MUST not call one another, and MUST receive only the palette and narrow context required for their work. The runner MUST use a deterministic adapter order, continue after an isolated failure, and aggregate complete, partial, skipped, and shared-prerequisite outcomes. An absent application, permission error, unsupported configuration, or write failure MUST remain attributable to the affected adapter.

#### Scenario: One adapter failure leaves successful work intact

- GIVEN macOS, Ghostty, and VS Code are enabled and VS Code writing fails
- WHEN the runner applies the operation
- THEN macOS and Ghostty results remain successful if they completed
- AND the VS Code result contains the failure cause
- AND the aggregate status is partial failure

#### Scenario: Disabled adapters are not touched

- GIVEN an adapter is not selected
- WHEN an apply operation runs
- THEN no files, processes, or platform state belonging to that adapter are changed
- AND its result is reported as disabled or omitted according to the shared result contract

### Requirement: TUI and CLI parity

The TUI MUST provide source selection, palette preview, adapter selection, apply, and per-adapter result display over the tested core. The CLI MUST provide equivalent source, preview/apply, adapter-selection, and diagnostic operations suitable for automation. Both surfaces MUST expose the same partial-failure meaning, adapter follow-up instructions, and safe empty/error states.

#### Scenario: TUI previews before apply

- GIVEN the user selects a Wallhaven result or local image in the TUI
- WHEN the palette is available
- THEN the TUI shows the palette and enabled adapters before applying changes
- AND applying requires the user-selected adapter set

#### Scenario: CLI is automation-friendly

- GIVEN a script supplies a source and adapter selection
- WHEN the CLI runs
- THEN it emits stable machine-usable success, partial-failure, and prerequisite-failure outcomes
- AND it returns a non-success exit status for partial or failed application without hiding successful results

#### Scenario: TUI and CLI handle empty or unavailable sources

- GIVEN Wallhaven returns no usable results or a local path is unreadable
- WHEN either surface requests selection or apply
- THEN it displays a clear error and recovery guidance
- AND it does not enter an unexplained blank or destructive state

### Requirement: Attribution and request policy

Hather MUST preserve enough Wallhaven metadata to attribute selected artwork and MUST display or document attribution in the user-facing result/help documentation. Remote requests MUST use a bounded rate-limiting policy, MUST avoid unnecessary downloads, and MUST classify server rate-limit responses. Local images MUST not require attribution to Wallhaven.

#### Scenario: Wallhaven attribution is available

- GIVEN a wallpaper was selected from Wallhaven
- WHEN Hather shows, persists, or reports the applied result
- THEN the source identity and required attribution text or link are available to the user

#### Scenario: Download policy is bounded

- GIVEN a remote image has already been cached under the same source identity
- WHEN the user applies it again and the cache entry is valid
- THEN Hather may reuse the cached artifact instead of downloading again
- AND cache reuse does not bypass a required freshness or permission check

### Requirement: Phased delivery and review budget

Implementation MUST be delivered as independently testable reviewable phases: contract fixtures/spikes; core, sources, palette, and CLI; macOS and Ghostty; herdr, Neovim, and VS Code; TUI completion; and Homebrew release hardening. Each phase MUST remain at or below the canonical 400 changed-line review budget by default. If a phase exceeds 400 changed lines, work MUST pause and MUST proceed only after an explicit human decision selects chaining/splitting it into independently testable phases within the budget, or approves a `size:exception` for that specific oversized phase. An exception MUST NOT be inferred and MUST NOT change the default budget. Work MUST NOT hide changes, weaken tests, or combine phases to evade review.

#### Scenario: Phase boundaries remain independently testable

- GIVEN an implementation phase is submitted
- WHEN its focused checks run
- THEN the phase has a tested seam and does not require an unimplemented later adapter or TUI to be buildable

#### Scenario: Review-budget risk pauses delivery

- GIVEN a phase is estimated or observed above 400 changed lines
- WHEN the phase is prepared for review
- THEN work pauses until an explicit human decision selects chaining/splitting or approves `size:exception` for that phase
- AND unapproved oversized work does not proceed
- AND the explicitly approved 419-line historical Phase 1a satisfies the exception only for that phase, without changing the 400-line default

### Requirement: Homebrew-first release acceptance

The initial release MUST be installable through an owned Homebrew tap or equivalent owned formula path, without requiring homebrew-core acceptance. The release MUST provide version output and help, build without network access during formula installation, and document supported macOS scope, optional Wallhaven credentials, attribution, Automation remediation, adapter apply timing, missing-target behavior, and known non-goals. Release acceptance MUST include a bounded macOS apply rehearsal and offline/local-image verification.

#### Scenario: Fresh installation exposes the product

- GIVEN a supported macOS user has access to the owned release tap
- WHEN the user installs Hather through Homebrew
- THEN the binary is available on PATH
- AND `--help` and version output work
- AND the formula's test verifies the installed executable

#### Scenario: Release does not require optional applications

- GIVEN Ghostty, herdr, Neovim, or VS Code is absent
- WHEN the installed Hather binary runs a local-image preview or apply with those adapters selected
- THEN installation and the shared pipeline remain usable
- AND each absent target produces an actionable independent result

#### Scenario: Release acceptance proves offline behavior

- GIVEN network access is unavailable
- WHEN the installed binary processes a supported local image
- THEN it can produce a deterministic palette and run applicable adapters without Wallhaven access
- AND the release acceptance record identifies any adapter that requires a local application or permission
