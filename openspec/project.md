# Hather — Project Context

## Identity

Hather is a fresh, standalone, open-source macOS application. It is separate from HatDots.

## Product direction

Hather turns a wallpaper source into a coherent color palette and applies that palette through theme adapters. The initial adapter direction includes macOS appearance and wallpaper, Ghostty, herdr, Neovim, and VS Code, with more adapters added only through later phase decisions.

## MVP surface

- TUI for interactive use.
- CLI for scripting and automation.
- Wallhaven as the initial wallpaper source integration.
- Homebrew as the first distribution channel.

## Technical direction

Likely Go, targeting macOS first. Keep implementation minimal and phase-gated; confirm package choices, adapter boundaries, persistence, and platform details during explore/proposal/spec/design rather than prematurely implementing them.

## Non-goals for initialization

No application code, integrations, packaging implementation, or final architecture is created in this phase. HatDots is not modified.

## SDD defaults

- Artifact store: OpenSpec.
- Execution: auto.
- Delivery strategy: ask-on-risk.
- Review budget: 400 changed lines.
- Strict TDD: enabled when implementation phases begin.
