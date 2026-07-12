# Web カラーパレット

色トークンの単一情報源は `src/app/globals.css` です。ライトモードが既定で、祖先要素
（通常は `html`）に `data-theme="dark"` を付けると、以下のダークパレットにすべての
色・shadow トークンが切り替わります。component CSS は引き続き意味的な変数だけを使います。

## Neutral と semantic トークン

| トークン | ダーク値 | トークン | ダーク値 |
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

## Information と Markdown トークン

| トークン | ダーク値 | トークン | ダーク値 |
| --- | --- | --- | --- |
| `--info-accent` | `#79C0FF` | `--info-bg` / `--info-border` / `--info-text` | `#12253D` / `#28547A` / `#B6D7FF` |
| `--markdown-link` / `--markdown-checkbox` | `#79C0FF` | `--markdown-link-hover` | `#A5D6FF` |
| `--markdown-link-visited` | `#C4B5FD` | `--markdown-inline-code-bg` / `--markdown-inline-code-text` | `#1D2D3D` / `#A5D6FF` |
| `--markdown-quote-bg` / `--markdown-quote-border` / `--markdown-quote-text` | `#172536` / `#58A6FF` / `#B6D7FF` | `--approval-bg` | `#3D310F` |

## Priority・status・toast・filter トークン

各組は `accent / background / text` の順です。

| 役割 | ダーク値 |
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

## Shadow

ダークモードでは、暗い surface 上でも階層を保つために shadow の不透明度を上げます。
すべての `--shadow-*` トークンをテーマ selector で上書きするため、component ごとに
shadow 色を直接書かないでください。
