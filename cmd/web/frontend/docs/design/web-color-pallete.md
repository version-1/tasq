# Web Color Palette

`src/app/globals.css` is the source of truth for color tokens. Light mode is
the default; adding `data-theme="dark"` to an ancestor (normally `html`) switches
every color and shadow token below to the dark palette. This document records
the dark values so component CSS can continue to use semantic variables only.

## Neutral and semantic tokens

| Token | Dark value | Token | Dark value |
| --- | --- | --- | --- |
| `--bg` / `--white` / `--surface-wash` | `#111315` | `--surface` | `#181B1F` |
| `--surface-strong` / `--accent` / `--primary-black` | `#F0F3F6` | `--border` / `--light-gray` | `#30363D` |
| `--text` | `#F0F3F6` | `--muted` / `--medium-gray` | `#9DA7B3` |
| `--dark-gray` | `#D0D7DE` | `--extra-light-gray` | `#21262D` |
| `--surface-hover` | `#262C36` | `--surface-row-hover` | `#1C2128` |
| `--accent-color` | `#A78BFA` | `--glow-color` | `#C4B5FD` |
| `--accent-strong` / `--control-strong-hover` | `#FFFFFF` | `--control-strong` | `#F0F3F6` |
| `--control-divider` | `#9DA7B3` | `--backdrop` | `rgb(0 0 0 / 60%)` |
| `--danger` | `#FF7B72` | `--warning` | `#F2CC60` |
| `--ok` | `#3FB950` | `--project-bg` / `--project-text` | `#12343B` / `#7DD3DC` |

## Informational and Markdown tokens

| Token | Dark value | Token | Dark value |
| --- | --- | --- | --- |
| `--info-accent` | `#79C0FF` | `--info-bg` / `--info-border` / `--info-text` | `#12253D` / `#28547A` / `#B6D7FF` |
| `--markdown-link` / `--markdown-checkbox` | `#79C0FF` | `--markdown-link-hover` | `#A5D6FF` |
| `--markdown-link-visited` | `#C4B5FD` | `--markdown-inline-code-bg` / `--markdown-inline-code-text` | `#1D2D3D` / `#A5D6FF` |
| `--markdown-quote-bg` / `--markdown-quote-border` / `--markdown-quote-text` | `#172536` / `#58A6FF` / `#B6D7FF` | `--approval-bg` | `#3D310F` |

## Priority, status, toast, and filter tokens

Each triplet is `accent / background / text`.

| Role | Dark values |
| --- | --- |
| Priority high | `#F2CC60` / `#3D310F` / `#F8E3A1` |
| Priority normal | `#58A6FF` / `#112D4A` / `#A5D6FF` |
| Priority low | `#8B949E` / `#24292F` / `#C9D1D9` |
| Status backlog | `#8B949E` / `#24292F` / `#C9D1D9` |
| Status ready | `#56D364` / `#12372A` / `#7EE787` |
| Status in progress | `#58A6FF` / `#112D4A` / `#A5D6FF` |
| Status review | `#E3B341` / `#3D2E0C` / `#F2CC60` |
| Status done | `#3FB950` / `#12372A` / `#7EE787` |
| Status blocked | `#F0883E` / `#3D260F` / `#F5B77A` |
| Status failed | `#FF7B72` / `#3D1F22` / `#FFA198` |
| Status muted | `#8B949E` / `#24292F` / `#C9D1D9` |
| Toast error (`accent / bg / border / text`) | `#FF7B72` / `#3D1F22` / `#6E2B32` / `#FFA198` |
| Toast success (`accent / bg / icon`) | `#56D364` / `#12372A` / `#7EE787` |
| Filter chip (`bg / border`) | `#24292F` / `#30363D` |

## Shadows

Dark mode increases shadow opacity to preserve hierarchy on dark surfaces.
All `--shadow-*` tokens are overridden in the theme selector; do not add
component-level shadow color literals.
