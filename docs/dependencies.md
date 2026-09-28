# Dependency decisions and go.mod scaffold notes

Phase 0 decisions only. `go.mod` is **not** created here (no Go toolchain on this host yet);
it lands in Phase 1. Versions pinned via `proxy.golang.org/@latest` on 2026-09-08.

## Decisions

- HTTP + download: stdlib `net/http`; image decode: stdlib `image/jpeg`, `image/png`.
- TOML (herdr/config): `github.com/BurntSushi/toml` — **v1.6.0**.
- JSONC (VS Code): `github.com/tailscale/hujson` — pseudo-version
  **v0.0.0-20260727124030-b80ff77dac4f** (no tagged release). Parses JSONC (comments, trailing
  commas) and standardizes to JSON. Version pinned; exact call surface confirmed in Phase 3c.
- WebP: `golang.org/x/image` — **v0.46.0**.
- TUI only: `bubbletea` **v1.3.10**, `lipgloss` **v1.1.0**, `bubbles` **v1.0.0**.
- **No** Peachy, **no** CLI framework (stdlib `flag.FlagSet`).

## go.mod scaffold (future Phase 1, illustrative)

```text
module github.com/<owner>/hather

go 1.24

require (
    github.com/BurntSushi/toml v1.6.0
    github.com/tailscale/hujson v0.0.0-20260727124030-b80ff77dac4f
    golang.org/x/image v0.46.0
    github.com/charmbracelet/bubbletea v1.3.10
    github.com/charmbracelet/lipgloss v1.1.0
    github.com/charmbracelet/bubbles v1.0.0
)
```

TUI packages are imported only by `internal/tui`; core/adapter packages must not import them.
Go version line is finalized when the toolchain is installed in Phase 1.
