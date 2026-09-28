# Platform verification — Phase 0 bounded probes

Verified 2026-09-08 on macOS 26.6.2. "Observed" = ran here; "docs" = official docs read, not
live-verified; "unverified" = not tested on this machine.

## Ghostty (observed 1.3.1 stable)

- `ghostty --version` → 1.3.1. `+list-themes` enumerates builtin + user themes.
- User config at `~/.config/ghostty/config` (inline palette style); `~/.config/ghostty/themes`
  exists.
- `theme` option (docs): built-in name, custom theme name, or absolute path; a custom name is
  searched in `~/.config/ghostty/themes` first; a theme file is another Ghostty config file.
- Reload (docs): runtime reload via `cmd+shift+,` / `reload_config` action; docs caveat "Some
  configuration options cannot be reloaded at runtime; others may only apply to newly created
  terminals." Whether `theme = hather` hot-reloads on 1.3.1 is **unverified** (not tested
  live). Adapter result stays `applied; reload behavior depends on installed version`.

## herdr (observed 0.8.2)

- `herdr --version` → 0.8.2. `--default-config` prints the full default config.
- **Correction:** `herdr theme list` does **not** exist on 0.8.2 (`unknown command: theme`).
  Use `--default-config` / `config check` (observed `config: ok`).
- Owned theme keys (observed in `--default-config`):
  - `[theme]`: `name`, plus `auto_switch`, `dark_name`, `light_name`.
  - `[theme.custom]`: `sidebar_bg`, `active_row_bg`, `selection_bg`, `panel_bg`, `accent`,
    `red`, `green` — **no `blue` or `yellow`**; Phase 3a must own only the observed set.
  - Built-ins: catppuccin, terminal, tokyo-night, dracula, nord, gruvbox, one-dark, solarized,
    kanagawa, rose-pine, vesper.
- Reload (observed command): `herdr server reload-config`. Hather still only writes the file +
  reports the follow-up; it never invokes this.

## Apple Events / macOS

- osascript error format (observed): `<line>:<col>: execution error: <message> (<code>)`.
  Messages are **localized** (this host reports Spanish); classify on the numeric code. `-1743`
  = Automation denial (fixture `exec_error_1743.json`).
- Appearance scheduling key `AppleInterfaceStyleSwitchesAutomatically` is absent on this host
  (auto off; `AppleInterfaceStyle` = `Dark`). Scheduling conflict is a documented fixture.
- Path escaping (fixture `exec_path_escaping.json`): AppleScript string literals
  (`POSIX file "..."`); escape `\` and `"` only, run via `exec.Command` arg array (no shell).

## VS Code (unverified on host)

`~/Library/Application Support/Code/User` is absent here, so the adapter is a config-file
writer only; no live-apply claim. Fixture: `internal/adapter/vscode/testdata/settings.json`.
