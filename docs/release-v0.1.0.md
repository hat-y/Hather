# Hather v0.1.0 release record

This is the release-readiness record for Hather v0.1.0. It records prepared packaging and a bounded local rehearsal; it does **not** claim that a production tag, GitHub release, tap push, or public Homebrew install has occurred.

## Release scope

- Supported platform: macOS 14 Sonoma and later; the rehearsal host was macOS 26.6.2 on Apple Silicon.
- Distribution: the owned `hat-y/homebrew-hather` tap. homebrew-core submission, audit, and acceptance are outside this release gate.
- Optional targets: Ghostty, herdr, Neovim, and VS Code are not formula dependencies and may be absent.
- Credentials: Wallhaven is keyless by default. A key is accepted only through `--wallhaven-api-key` or `WALLHAVEN_API_KEY`, with the flag taking precedence; it is never stored in config, state, cache metadata, normal output, or diagnostics.

## Prepared packaging and Go build contract

`go.mod` declares `go 1.27`. The release workflow uses `actions/setup-go` with `go-version-file: go.mod`, so CI selects a toolchain from that module requirement rather than duplicating an exact version. The formula uses Homebrew's standard `depends_on "go" => :build`; it is not a versioned formula or an exact toolchain pin. Go enforces the `go 1.27` minimum during the build, so a Homebrew Go toolchain older than that minimum fails rather than silently building an unsupported release. The observed local toolchain for this correction was Go 1.27.1.

The formula builds `./cmd/hather` with `-mod=vendor` and `-ldflags` version injection. The release workflow vendors dependencies before packaging. Local exact-current-candidate installation and formula test passed; the production release workflow has not run.

The workflow still validates a `vSemVer` tag, runs tests/builds, creates a source archive and checksum, creates a GitHub release, renders the owned-tap formula, and pushes it. Those are prepared source instructions only: no tag, GitHub release, tap push, or production formula mutation occurred in this correction.

## Bounded acceptance status

The exact-current-candidate local rehearsal passed: source archive SHA-256 `541d001ab20b7c14f209da9d25bb92971cc604efe261cd517908bb2f8b57bd11`; vendored tests `a03c0063632b8276ca9b97f93679e9f4cb66b8722a793cd150a32ccbd3a4ea49`; Homebrew install `04e381ddc6bca2216bd6ada16e2e3244be59fe30e5e5ff33ed0be255501a4891`; installed formula `brew test` `c336acbd8147f7104a040fa20a267bc032d0bef35be724e8c135223dfcff375f`. The formula test exercised installed-binary help/version, local PNG preview, and independently unavailable optional adapters on apply. Temporary install, tap, and workspace cleanup were verified absent. These hashes were observed by the parent, not independently recomputed by the verifier. This was local, not a production tag/release/push or public tap installation.

The Automation-denial path remains covered by the injected macOS adapter fixture test (`TestApplyClassifiesAutomationDenial`) rather than changing this machine's TCC permission state. It verifies the actionable remediation text: **System Settings → Privacy & Security → Automation**. A live denial prompt or permission-state mutation was intentionally not performed.

## User-facing behavior

| Concern | Release behavior |
| --- | --- |
| Attribution | Credit downloaded Wallhaven artwork with its page URL; local images need no Wallhaven attribution. |
| macOS timing | All-desktop wallpaper/appearance uses System Events; universal instant propagation is not promised. |
| Ghostty timing | Hather writes only its owned theme and marked selection; reload behavior depends on the installed version. |
| herdr timing | Theme-only TOML write; restart required and no process signal or reload is sent. |
| Neovim timing | Writes `colors/hather.vim`; use it next start or run `:colorscheme hather` manually. |
| VS Code timing | Merges only owned color customizations; reload the window to observe them. |
| Missing targets | Each selected absent target returns its own actionable `unavailable` result; installation and the shared pipeline continue. |
| Apply semantics | Apply is not a transaction: successful sibling changes are retained if another adapter fails. |

## Known non-goals

This release excludes pre-macOS-14 support, HatDots integration, per-Space wallpaper targeting, Tinted appearance, a GUI/menu-bar app, background rotation, a mandatory Peachy binary, plaintext credential persistence, automatic installation of optional targets, live Neovim control, full VS Code theme parity, and homebrew-core submission.
