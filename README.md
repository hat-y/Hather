# Hather

Hather derives a deterministic palette from a local image or Wallhaven wallpaper, then applies it to the macOS targets you explicitly select.

## Install

The first distribution path is the owned Homebrew tap; homebrew-core is not part of this release. After a tagged release is published:

```sh
brew tap hat-y/hather
brew install hather
hather --help
hather --version
```

Hather supports macOS 14 Sonoma and later. The release rehearsal host was macOS 26.6.2.

## Quick path

A local image needs neither network access nor a Wallhaven key:

```sh
hather preview --image /path/to/wallpaper.png
hather apply --image /path/to/wallpaper.png --adapters ghostty,herdr,neovim,vscode
```

Choose adapters explicitly. `apply` is not a cross-adapter transaction: one target can fail or be unavailable while the other selected targets retain their independent results.

## Wallhaven and attribution

Public Wallhaven use is keyless. When needed, provide a key only with `--wallhaven-api-key` or `WALLHAVEN_API_KEY`; the flag wins. Hather never accepts or writes this credential in its config, state, cache metadata, normal output, or diagnostics.

Credit Wallhaven artwork with its selected page URL. Local images have no Wallhaven attribution requirement. Hather follows the documented 45 API-calls-per-minute policy and reports rate limits rather than silently retrying.

## Adapter behavior

| Adapter | What Hather changes | Timing and missing target behavior |
| --- | --- | --- |
| macOS | Wallpaper on all desktops and requested appearance through System Events | Platform propagation is not promised to be instant. Automation denial is independent and actionable; see below. |
| Ghostty | Hather-owned theme plus a marked selection only | Reload behavior depends on the installed Ghostty version. Missing config is `unavailable`; unowned theme or inline-palette settings are protected as conflicts. |
| herdr | Owned `[theme]` and `[theme.custom]` values only | Restart required; Hather does not signal or reload a running herdr instance. Missing config is `unavailable`. |
| Neovim | `colors/hather.vim` only | Available next start or manually with `:colorscheme hather`; Hather does not use a server socket. |
| VS Code | Hather-owned color customization entries only | Reload the window to observe changes. Missing settings parent is `unavailable`. |

If macOS blocks System Events, enable Hather in **System Settings → Privacy & Security → Automation**, then retry. Do not disable automatic appearance scheduling: Hather reports it as a safe conflict rather than silently changing it.

## Non-goals

Hather does not support platforms before macOS 14, HatDots integration, per-Space wallpaper targeting, Tinted appearance, a GUI/menu-bar app, background rotation, a mandatory Peachy dependency, automatic plaintext credential storage, Ghostty/herdr/Neovim/VS Code installation, live Neovim control, or full VS Code theme parity.

See [the v0.1.0 release record](docs/release-v0.1.0.md) for packaging and rehearsal evidence.
