# Peachy compatibility spike — observed contract

Run 2026-09-08 on macOS 26.6.2. Peachy is comparison material only — **not** a production or
test dependency; Hather's palette engine is the native Go engine.

## Binary

`~/.local/bin/peachy` — Go Mach-O binary (14.6 MB), no Homebrew formula.

## Commands observed

```sh
peachy generate <abs-image> --save hather-spike --no-apply --output <dir>
peachy export hather-spike <dir>
```

`generate --help`: modes `normal|monochromatic|analogous|pastel|material`, `--light`, `--save`,
`--no-apply` (files only), `--output`, `--random`. Ran with `HOME` redirected to a temp dir.

## Palette shape (printed by generate)

`Normal:` 8 colors, `Bright:` 8 colors, then `FG:` and `BG:` hex values.

## Export file shapes (observed, 17 files)

`alacritty.toml`, `btop.theme`, `chromium.theme`, `colors.toml`, `ghostty.conf`, `hyprland.conf`,
`hyprlock.conf`, `icons.theme`, `kitty.conf`, `mako.ini`, `neovim.lua`, `swayosd.css`,
`vscode.empty.json`, `walker.css`, `warp.yaml`, `waybar.css`, `wofi.css`.

Notable:

- `ghostty.conf`: `background = #hex`, `foreground = #hex`, 16 `palette = N=#hex` lines.
- `colors.toml`: `accent`, `active_border_color`, `active_tab_background`, `cursor`,
  `foreground`, `background`, `selection_foreground`, `selection_background`, `color0..15`.
- `neovim.lua`: LazyVim plugin spec for `bjarneo/aether.nvim` (branch `v2`), not a plain
  `colors/*.vim` file — Hather instead writes an additive `hather.vim`.
- `vscode.empty.json`: `{}` — Peachy exports **no** VS Code customizations; Hather needs its
  own JSON/JSONC merge.

## Saved theme (observed)

`generate --save` writes `~/.config/peachy/themes/<name>.toml` (`hather-spike.toml`).

## Decision

Native Go palette engine confirmed. Peachy stays out of `go.mod`, tests, and the formula.
