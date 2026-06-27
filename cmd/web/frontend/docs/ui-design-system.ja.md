# Web フロントエンド UI デザインシステム

このドキュメントは、`cmd/web/frontend` の React/Vite フロントエンドが
現在採用しているデザイントークン、レイアウト基盤、UI primitive、
feature レベルのパターンをまとめたものです。ルーティングおよび
コンポーネント配置ルールを定義した [docs/design.ja.md](design.ja.md) と
併せて参照してください。

目的は、新しい画面を実装するときに app shell・issue board・issue 詳細
ページ・table view・workflow settings・dashboard・dialog といった既存
画面と一貫した見た目と挙動を保つための、単一のリファレンスを提供する
ことです。

## 単一情報源

- トークン: `src/app/globals.css` が CSS custom property と global な
  要素デフォルトを定義します。
- ビジュアルリファレンス: `ui-design-system.html` は token と component
  pattern をブラウザで確認しやすい形で一覧します。
- UI primitive: `src/components/ui/<name>/index.module.css` が共有
  コンポーネントごとのスタイルを所有します。
- レイアウト: `src/components/layout/` が app shell、header、sidebar、
  breadcrumb、shell レベルの dialog を所有します。
- Feature パターン: `src/features/<feature>/components/<name>/index.module.css`
  が issue board、issue 詳細ページ、table view、dashboard panel、
  workflow settings panel などの domain を意識したセクションを所有します。

スタイリング機構は CSS Modules のみで、Tailwind、styled-components、
theme provider、ランタイム CSS-in-JS は使用していません。コンポーネント
スタイルは `index.module.css` にコンポーネントと一緒に置く必要があります。

## Foundations

Tasq は issue と agent orchestration を扱う、業務に集中するための UI です。
見た目は静かで、密度があり、読みやすく、運用作業に向いているべきです。
装飾や marketing 的な見せ方よりも、スキャン、比較、反復操作、handoff の
明確さを優先します。

Foundations は、token や component の上位にある設計判断の層です。個別の
UI 判断が component rule で明示されていない場合は、この原則を最もよく
保つ選択をしてください。

### Product Personality

- **静かで実用的**: product screen では、装飾的な gradient、過大な hero、
  ornamental illustration、一回限りの visual effect を避けます。
- **高密度だが整理されている**: 反復的な運用作業に必要な情報を出しつつ、
  予測しやすい grouping、alignment、table / card 構造で読みやすさを保ちます。
- **表現より system を優先**: 新しい visual treatment を作る前に、既存の
  token、surface、primitive を再利用します。
- **状態を前面に出す**: status、priority、project、run state、workflow state
  は、色だけに依存せず比較しやすくします。

### Information Hierarchy

- page-level の階層は、色よりも layout、spacing、type weight で作ります。
- table は比較とスキャンのために使います。控えめな row hover、明確な header、
  安定した column、compact な badge を使います。
- card は grouped work item、modal、本当に frame が必要な tool のために使います。
  full-width layout や table の方が直接的な page section を、装飾的な card stack
  で囲まないでください。
- primary action は目立たせすぎず、数を絞ります。多くの画面では dominant な
  action cluster は 1 つにし、それ以外は neutral / tertiary action にします。

### Interaction Principles

- control は hover 前から control として認識できる必要があります。filter chip、
  menu、button、tab trigger は、境界、affordance、active rule のいずれかを
  持ちます。
- 反復操作では layout shift を最小化します。table、badge、control、counter、
  toolbar などの fixed-format element は安定した寸法を持たせます。
- status と priority は、必要に応じて label と icon、dot、shape を併用します。
  色は意味を補強するもので、単独で意味を担わせません。
- keyboard と screen reader の挙動は component contract の一部であり、後から
  追加する装飾ではありません。

### Layout And Density

- app shell の rhythm から始めます。page padding は `--space-6`、control は
  compact にし、surface は `--radius-sm` / `--radius-md` を基本にします。
- panel、table、sidebar、tool の内部では compact な type を使います。大きな
  type は実際の page title に限定します。
- 新しい breakpoint を追加する前に、既存の `1060px`、`900px`、`860px`、
  `720px`、`640px` を優先します。
- nested card は避けます。すでに surface が content を frame している場合、
  内側の group は通常、border、divider、table、または frame しない layout で
  表現します。

### Accessibility Principles

- icon だけの interactive element には、wrapping control 側に accessible label
  を付けます。
- control を custom style する場合でも、focus-visible は見える状態で残します。
- state を持つ control は、その interaction を所有する component が
  `aria-expanded`、`aria-controls`、checked state、selected tab state、menu role
  を公開します。
- state を色だけで表現しません。text、icon、dot、border、position のいずれかと
  組み合わせます。

### Anti-Patterns

- 既存 token がある役割に対して、新しい hex 値を追加する。
- shared primitive を compose できるのに、feature-local な button、badge、menu、
  table を作る。
- 1 画面の微調整のために `13px` や `17px` のような spacing 値を導入する。
- operational product view に、装飾的な card、gradient、大きな hero、
  illustrative section を持ち込む。
- issue status label などの feature semantics を domain-independent な UI
  primitive に移す。

## デザイントークン

Design token は、上記 Foundations を実装へ落とすための contract です。
global token はすべて `src/app/globals.css` の `:root` で宣言されています。
コンポーネントは CSS variable 経由で token を参照し、対応する token が存在する
限り、surface、text、border、status tone、spacing、radius、shadow、z-index、
font stack に固定値を直接記述してはいけません。

### Token Model

Tasq は実務上の token layer を 3 つに分けます。

| Layer | 目的 | 例 |
| ----- | ---- | -- |
| Base token | system scale を定義する共有 primitive 値。 | `--space-4`, `--radius-sm`, `--font-mono` |
| Semantic token | component をまたいで使う product role。 | `--surface`, `--text`, `--danger`, `--surface-hover` |
| Component / feature token | global role を component や feature に適用する local alias。 | `--badge-bg`, `--status-color`, `--ledger-rule` |

component / feature token は、重複を減らす場合や feature palette を閉じ込める
場合に使えます。ただし、その値が shared system の一部である場合は、global token
を参照する形にしてください。

### Token Governance

- token は、値が繰り返し使われる role または意図的な system scale step を
  表す場合だけ追加します。
- token 名は見た目ではなく役割で付けます。interaction surface として使うなら
  `--gray-100` より `--surface-hover` を優先します。
- feature 固有 palette は、独立した 2 箇所以上で同じ role が必要になるまで local
  に保ちます。
- すべての数値を token 化しません。固定 component dimension、typography size、
  breakpoint は、繰り返し使われる contract になるまで literal のままで構いません。
- 新しい token を導入するときは、この document、英語版、該当 token family を表示する
  visual reference を更新します。
- token を置き換えるときは、既存 call site に migration path が必要な場合だけ古い
  token を残します。それ以外は同じ変更で call site を更新します。

### カラー

| トークン             | 値        | 役割 |
| -------------------- | --------- | --- |
| `--bg`               | `#FFFFFF` | ページ既定の背景 |
| `--surface`          | `#FFFFFF` | card、panel、modal、dropdown の背景 |
| `--surface-strong`   | `#0A0A0A` | 強い surface 用の予約色 (利用は稀) |
| `--border`           | `#E5E5E5` | panel・field・table の標準 border |
| `--text`             | `#0A0A0A` | 標準前景文字 |
| `--muted`            | `#666666` | 副次テキスト、metadata、補助文 |
| `--accent`           | `#0A0A0A` | primary accent (`--primary-black` と同色) |
| `--accent-color`     | `#7C3AED` | フォームコントロールのネイティブ `accent-color` |
| `--glow-color`       | `#A78BFA` | 紫の glow アクセント用の予約色 |
| `--accent-strong`    | `#1A1A1A` | accent の hover/pressed バリエーション |
| `--danger`           | `#B42318` | エラー、破壊的フィードバック |
| `--warning`          | `#A15C07` | 注意、承認待ち |
| `--ok`               | `#087443` | 成功、ready 状態 |
| `--primary-black`    | `#0A0A0A` | primary ボタン背景、見出し |
| `--dark-gray`        | `#1A1A1A` | アイコンの濃色、split ボタンの仕切り |
| `--medium-gray`      | `#666666` | 非アクティブな tab、ラベル色 |
| `--light-gray`       | `#E5E5E5` | border、switch track、ledger rule |
| `--extra-light-gray` | `#F5F5F5` | hover surface、code block 背景、chip 背景 |
| `--white`            | `#FFFFFF` | 純粋な白の予約 |
| `--surface-wash`     | `#FBFBFB` | app shell と薄い page wash |
| `--surface-hover`    | `#F0F0F0` | sidebar active と hover row surface |
| `--surface-row-hover`| `#FAFAFA` | table row hover surface |
| `--control-strong`   | `#111111` | 強い action hover / filter apply |
| `--control-strong-hover` | `#252525` | 強い action の hover variant |
| `--control-divider`  | `#303030` | split button 内部 divider |

status、priority、project、approval、toast、filter-chip の tone は global
token です。component は `--badge-bg` のような local variable へ代入できますが、
元の値は `:root` から参照します。

| 役割 | トークン |
| ---- | -------- |
| Priority high | `--priority-high-accent`, `--priority-high-bg`, `--priority-high-text` |
| Priority normal | `--priority-normal-accent`, `--priority-normal-bg`, `--priority-normal-text` |
| Priority low | `--priority-low-accent`, `--priority-low-bg`, `--priority-low-text` |
| Status backlog | `--status-backlog-accent`, `--status-backlog-bg`, `--status-backlog-text` |
| Status ready | `--status-ready-accent`, `--status-ready-bg`, `--status-ready-text` |
| Status in progress | `--status-in-progress-accent`, `--status-in-progress-bg`, `--status-in-progress-text` |
| Status review | `--status-review-accent`, `--status-review-bg`, `--status-review-text` |
| Status done | `--status-done-accent`, `--status-done-bg`, `--status-done-text` |
| Status blocked | `--status-blocked-accent`, `--status-blocked-bg`, `--status-blocked-text` |
| Status failed | `--status-failed-accent`, `--status-failed-bg`, `--status-failed-text` |
| Status muted | `--status-muted-accent`, `--status-muted-bg`, `--status-muted-text` |
| Project | `--project-bg`, `--project-text` |
| Approval | `--approval-bg`, `--warning` |
| Toast | `--toast-error-*`, `--toast-success-*` |
| Filter chip | `--filter-chip-bg`, `--filter-chip-border` |

table 行が利用する status tone (`statusToneClassName`) は、単一の
`--status-color` にマッピングされます。
`src/features/issues/components/status-badge/index.module.css` で宣言され、
`src/components/ui/table/index.module.css` から同じ status 対応の行 class を
通じて再利用されます。

table view は、issue table を会計台帳のように見せるためにスコープ付きの
"ledger" パレットを宣言します。

| トークン            | 値        |
| ------------------- | --------- |
| `--ledger-ink`      | `#17201B` |
| `--ledger-muted`    | `#5F6B64` |
| `--ledger-rule`     | `#E6E6E6` |
| `--ledger-surface`  | `#FFFFFF` |
| `--ledger-wash`     | `#FBFBFB` |

これらは `table-view` と `filter-options` のローカル変数です。
共通の table primitive は `var(--ledger-rule, var(--border))` のように
参照することで、他の場所ではグローバルトークンにフォールバックします。

### タイポグラフィ

- 基本フォントファミリー: `--font-sans` (`body` で設定)。
- ID やコード風の値に使う等幅フォントスタック: `--font-mono`。
- 本番で利用している見出しサイズ:
  - ページタイトル (`h1`、header): `30px / 800`。
  - issue 詳細タイトル (`h2`): `28px / 既定 weight`、`640px` 未満で
    `22px` に縮小。
  - dashboard やページのセクション (`h2`): `22px`。
  - section 見出し (`h3`、panel): `16px`。
  - sidebar ブランド: `34px / 800`。
- 本文テキスト: table セル、補助文、ラベルは `14px`。
- metadata / caption: `12〜13px` を `--muted` で。
- フォーム field ラベル: `13px / 700`。
- フォーム field 入力: 継承フォント、`400` weight。

ボタンとフォームコントロールは、グローバルリセット
`button, input, select { font: inherit }` を通じてページフォントを継承します。

### 余白

spacing token は `:root` に定義し、UI 内で繰り返し使われている値を
4px ベースの scale と half step で表します。

| Token | 値 |
| ----- | -- |
| `--space-0` | `0` |
| `--space-1` | `4px` |
| `--space-1-5` | `6px` |
| `--space-2` | `8px` |
| `--space-2-5` | `10px` |
| `--space-3` | `12px` |
| `--space-3-5` | `14px` |
| `--space-4` | `16px` |
| `--space-4-5` | `18px` |
| `--space-5` | `20px` |
| `--space-5-5` | `22px` |
| `--space-6` | `24px` |
| `--space-7` | `28px` |
| `--space-8` | `32px` |
| `--space-10` | `40px` |
| `--space-20` | `80px` |

margin、padding、gap、positioning offset がレイアウト上の余白を表す場合は
これらの token を使います。サイズ、角丸、typography、breakpoint は、
documented token になるまでは literal のまま扱います。

### 角丸

| Token | 値 | 役割 |
| ----- | -- | ---- |
| `--radius-xs` | `5px` | filter popover 内の chip outline |
| `--radius-sm` | `6px` | button、form field、小型 surface |
| `--radius-md` | `8px` | panel、card、modal dialog、table wrapper |
| `--radius-lg` | `10px` | toast container |
| `--radius-pill` | `999px` | pill badge、switch track と thumb、timestamp chip |

### シャドウ

| Token | 役割 |
| ----- | ---- |
| `--shadow-panel` | 標準 panel と section surface |
| `--shadow-card` | issue card surface |
| `--shadow-dropdown` | dropdown と context menu |
| `--shadow-dialog` | modal dialog |
| `--shadow-toast` | toast container |
| `--shadow-switch-thumb` | switch thumb |
| `--shadow-header-action` | header action cluster |
| `--shadow-filter-trigger` | filter trigger の通常状態 |
| `--shadow-filter-active` | filter trigger の展開 / focus 状態 |
| `--shadow-filter-popover` | filter popover |

### Z-Index

| Token | 値 | 役割 |
| ----- | -- | ---- |
| `--z-menu` | `10` | context menu |
| `--z-sticky` | `10` | sticky header |
| `--z-popover` | `20` | filter popover |
| `--z-dialog` | `30` | dialog backdrop |
| `--z-toast` | `1000` | toast stack |

役割を満たす最小の layer を使ってください。明文化された理由なしに、
これらの layer 以外の場当たり的な z-index を導入してはいけません。

## レイアウトシステム

アプリケーションシェルは `src/components/layout/index.module.css` で
宣言された 2 カラムグリッドです。

```text
.appFrame {
  display: grid;
  grid-template-columns: 260px minmax(0, 1fr);
  min-height: 100vh;
}
```

- 左に `260px` の sidebar、右に流動的なコンテンツ。
- コンテンツカラムは `min-width: 0` を持つため、幅広の table がレイアウト
  全体を押し広げずにスクロールできます。
- `≤ 900px` ではグリッドが 1 カラムに崩れ、sidebar が下 border を持つ
  上部ストリップになります。
- shell の背景は `var(--surface-wash)`。card、panel、dialog、dropdown は
  `var(--surface)` の上に置きます。

### Sidebar

- `gap: 22px`、`padding: 28px 18px 18px` の縦スタック。
- ブランドリンク: `34px / 800`、`var(--primary-black)`。
- primary navigation: bottom border による区切り、行 gap `4px`。
- project list: 独自の `8px` 行 gap と、
  `12px / 700 / uppercase / --medium-gray` のセクションヘッダー。
- nav 行: 最小高さ `44px`、icon gap `12px`、`--radius-sm`、hover で
  `--surface-hover`。
- アクティブな route は `--surface-hover` 背景を使用 (border アクセントはなし)。
- sidebar 下部: settings リンクが `border-top` と theme switch 行で
  仕切られた同じ separator を持ちます。

### Header

- `top: 0` に sticky、`z-index: 10` と `1px` の bottom border。
- 3 つの行:
  - utility row: 通知ボタン、検索入力 (`360px` 最小幅とキーボードヒント)、
    more ボタン。
  - title row: ページタイトル (`30px / 800`) と split create button クラスター。
  - view row: tab。
- tab: gap `36px`。アクティブな tab は
  `box-shadow: inset 0 -2px 0 var(--primary-black)` を使い、
  前景色を `--primary-black` から継承します。
- split create button: 高さ `44px`、`--primary-black` 背景、hover で
  `#111111`、内部の `#303030` separator。

### Breadcrumb

`src/components/layout/header/breadcrumb` は水平リストを描画します。

- `14px` テキスト、`--medium-gray` 色。
- 項目間の gap は `8px`。
- 現在セグメントは `--primary-black` + weight `600`。

### コンテンツ padding

shell の `.content` は、ルーティング先の feature の周りに `24px` の余白を
追加します。issue table view のように溝まで広げたい feature view は、
これを `margin: -24px` で打ち消して自分の padding を再適用します。

## UI Primitive

UI primitive は `src/components/ui/` 配下に置きます。domain に依存せず、
再利用可能でなければなりません。

### Button (`ui/button`)

- 既定の primitive: `Button` は `--primary-black` 背景、白文字、
  高さ `40px`、radius `6px`、横 padding `0 16px` の primary アクションを
  描画します。
- 同伴の `.splitButton` セレクターは、末尾の `40px` 正方形 chevron スロットで
  同じ見た目を再利用します。
- グローバルな `button` リセットは、それ以外のすべての native button に
  中立の surface、`1px` border、radius `6px`、padding `8px 10px`、
  `--muted` の disabled 状態を与えます。

カスタムアクションバリアント (filter の `Apply`、dialog の submit、
header の `Create`) は所有モジュール内で個別にスタイル付けされますが、
常に `--primary-black`、radius `6px`、`--white` テキスト、hover で
`--accent-strong` (`#1A1A1A`) を再利用します。

### Badge (`ui/badge`)

`Badge` は、ラベル付きの chip とオプションのアイコンを描画します。
コンポーネントは `variant` prop で見た目を選びます。

- `project` — issue table と card で使う柔らかいティール chip。
- `priority-high` / `priority-normal` / `priority-low` — 対応するアクセント
  ドット付きのカラフルな pill。
- `status-backlog` / `status-ready` / `status-in-progress` / `status-review` /
  `status-done` / `status-blocked` / `status-failed` / `status-muted` —
  同じ pill 形状で status 固有のアクセント。

すべてのバリアントは、padding `4px 10px` の内側に `12px / 600` のテキストを
持ち、`999px` の完全な角丸を共有します。アイコンは `--badge-accent` を継承し、
ドットアクセントも同じ custom property を継承します。

### Switch (`ui/switch`)

- track `42px × 24px`、thumb `20px`、radius `999px`。
- 非アクティブ track: `--light-gray`。アクティブ track: `--primary-black`。
- thumb は `160ms` ease で水平方向に `18px` 移動。
- フォーカスリング: `2px solid var(--primary-black)` + offset `2px`。

### Pagination (`ui/pagination`)

- 末尾揃え、gap `12px` のインラインクラスター。
- サマリーテキストは `14px`、`--ledger-muted` (`--muted-foreground` に
  フォールバック)。
- `860px` 未満では `justify-content: space-between` に変化。

### Table (`ui/table`)

- ラッパー: `1px` border (`--ledger-rule`)、radius `6px`、最小高さ `920px`
  で空状態でも意図的に見えます。
- ヘッダー: `12px / 600`、ledger-muted 色、sticky に見える whitespace。
- セル: `14px`、padding `14px 12px`、`--ledger-rule` の `1px` bottom border。
- hover 行: 背景 `#FAFAFA`。
- ID セルは等幅スタックを使用。
- ソートボタン: アクティブでなければ opacity `0.48` の chevron アイコン。

### Context Menu (`ui/context-menu`)

- トリガーに対して絶対配置され、`40px` 下に開きます。
- 最小幅 `188px`、padding `8px`、行 gap `4px`、radius `8px`、card shadow。
- アイテム: padding `8px 8px`、radius `6px`、hover で `--extra-light-gray`。
- グループラベル: `12px` muted、メニューグループのタイトル用。
- ヘルプフッター: キーボードヒント用の小さな muted 段落。

### Modal (`ui/modal`)

`ModalOutlet` はポータルスロットを提供します。dialog の表現は呼び出し側の
layout コンポーネントが所有します。
繰り返し登場するルール (`add-issue-dialog` と `add-project-dialog` 参照):

- backdrop: `rgb(20 20 20 / 36%)`、ビューポート全面、padding
  `80px 24px 24px`、`z-index: 30`。
- dialog surface: 白いカード、radius `8px`、shadow
  `0 20px 50px rgb(0 0 0 / 18%)`、幅 `min(100%, 560px)` (project dialog は
  `520px`)。
- フォームレイアウト: gap `16px` の grid、`≤ 900px` で field grid が
  `1fr` に縮小。
- 送信ボタンは primary-black の見た目を再利用。キャンセル/閉じるボタンは
  中立な既定 button を使用。

### Markdown (`ui/markdown`)

- `react-markdown`、`remark-gfm`、`rehype-sanitize` を利用した再利用可能な
  Markdown レンダラー。
- table は各セルに `1px` border、ヘッダーは `--extra-light-gray`、padding
  `8px 10px`。
- 本文は長い URL や ID 用に `overflow-wrap: anywhere` で折り返し。
- `attachment://<id>` 参照は
  `/tracker/api/v1/attachments/<id>/content` に書き換えられます。

### Panel Message (`ui/pannel-message`)

- 致命的エラーと empty state 用の wide panel surface (`max-width: 960px`)。
  見出しは `--primary-black`、本文は `--danger`。

### Toast (`ui/toast`)

- 右上の固定スタック (`top: 24px`、`right: 24px`)。
- toast コンテナ: radius `10px`、padding `14px 16px`、toast カラーの
  `6px` 左アクセントバー。
- `error` アクセント: `#DC2626`。`success` アクセント: `#16A34A`。
- アイコンバブル: 色付き背景とアクセント前景を持つ `40px` の円。
- `720px` 未満ではスタックが全幅になります。

### Icon Proxy (`ui/icon-proxy`)

- `lucide-react` の精選サブセットを集約するラッパー。
- 既定 `size={16}`、`strokeWidth={2}`。呼び出し側は `name` と任意の
  `className`、`size`、`strokeWidth` を渡します。
- 常に `aria-hidden="true"` と `focusable="false"` で描画されます。

新しいアイコンを追加するときは、`icons` map に登録して prop 型を
リテラルユニオンのまま保ってください。

## Feature パターン

Feature コンポーネントは上記の primitive を product 固有のセクションに
組み立てます。以下に挙げる形は既に存在するパターンであり、新しい
feature 作業は並行する別パターンを発明せずに、これらを再利用してください。

### Card surface

panel、section、dialog、dashboard で利用される共通形:

```css
background: var(--surface);
border: 1px solid var(--border);
border-radius: 8px;
box-shadow: 0 1px 2px rgb(0 0 0 / 8%);
padding: 16px; /* セクションがヘッダーを所有する場合は 18px */
display: grid;
gap: 14px;
```

これは `runs-section`、`comment-list`、`issue-description`、
`status-actions`、`issue-header`、`workflow-meta-panel`、
`workflow-body-section`、`frontmatter-section` で繰り返し登場する形です。

### Issue Card

`features/issues/components/card`:

- radius `8px` のカード、`0 12px 32px rgb(0 0 0 / 8%)` の濃いめの shadow。
- タイトル行: `16px / 500` リンク + `32px` の menu トリガー。
- メトリック行: icon + value ペア。各メトリックを `1px` 左 border で区切ります。
- クイックアクションボタンはスコープ付き変数 (`--quick-action-bg`、
  `--quick-action-border`、`--quick-action-text`) を使い、各バリアント
  (`quickAction-ready`、`quickAction-done`) は色だけを上書きします。

### Issue Board

`features/issues/components/board`:

- 4 カラム grid、最小カラム幅 `260px`、狭いビューポート向けに
  x 軸スクロール。
- カラム間は `1px` 左 rule で区切り、最初のカラムだけ rule を隠します。
- カラムヘッダー: タイトル + カウント chip + アクションボタン。
  カウント chip は `--extra-light-gray` 上の `999px` pill。
- タスクリストの gap は `10px`。

### Issue Table View

`features/issues/components/table-view` は ledger パレットで `ui/table` を
拡張します。

- table の周囲を wash 背景と ledger rule border で囲みます。
- ツールバーは flex 行で、`860px` 未満では縦スタックに切り替わります。
- フィルタークラスターは `<fieldset>` 内に置かれ、`legend` は
  スクリーンリーダー向けに視覚的に隠されています。
- リセットボタンは tertiary アクション (透明背景、ledger muted テキスト、
  hover で ink) として読めます。

### Filter Options

`features/issues/components/filter-options`:

- トリガーサマリー chip: 最小高さ `40px`、radius `6px`、ledger surface、
  `aria-expanded="true"` で回転する chevron。
- 選択値はトリガー内の `999px → 5px` chip として表示。
- popover: radius `8px` のカード、`0 22px 46px` の shadow、最小幅
  `420px`、固定ヘッダー (`58px`) と clear/cancel ボタン、スクロールする
  option list (`max-block-size: 420px`)、primary apply ボタン (`#111111`)
  を持つアクション行。
- `≤ 860px` ではトリガーと popover が全幅に広がります。

### Issue 詳細ページ

`features/issues/components/issue-detail-page`:

- gap `16px` のページ grid。各セクションは標準の card surface。
- ヘッダーセクションは `28px` タイトル、issue ID 行、badge 行、4 カラム
  meta grid を使い、`900px` 未満で 2 カラム、`640px` 未満で 1 カラムに
  崩れます。
- `meta-item` は各エントリを top rule、`12px` muted `dt`、`700` weight
  `dd`、ID のための `overflow-wrap: anywhere` で描画します。
- `runs-section` は run 行を `8px` border のグループで囲みます。
- `run-row` は 2 カラム grid (`1fr` リンク領域 + `40px` コピーボタン)。
  リンクは等幅 ID と ledger-muted タイムスタンプを持つ 3 カラム内 grid です。
- `comment-list` はカード形状を再利用し、各コメントの前に
  `border-top` separator を入れます。
- `status-actions` は同じカード内の wrap-flex アクション行です。
- `issue-description` は共有 markdown レンダラーを使い、line-height
  `1.65`、アンカー色 `--primary-black`、インライン画像に border を付けます。

### Conversation ページ

`features/issues/components/conversation-page`:

- gap `16px` のページ grid。
- カードの timeline リスト。承認アイテムは `var(--warning)` の border と
  `inset 3px 0 0 var(--warning)` の左 rule を追加します。
- コマンドとコードブロックは radius `6px`、背景 `--extra-light-gray`、
  等幅フォントスタックを使用。
- exit code の警告は `rgb(220 38 38 / 10%)` の wash、danger border、
  `--danger` テキストを使用します。

### Dashboard

`features/dashboard/components/dashboard-view`:

- gap `18px` の縦 grid。
- メトリック grid: 既定ブレークポイントでは 4 カード、`1060px` 未満で
  2 カード、`720px` 未満で 1 カード。
- 詳細 grid: 2 カード。`720px` 未満で 1 カード。
- compact メトリックカードは `--extra-light-gray` を使い、shadow なし、
  値は `24px` に小さく。
- 分布バー: `--extra-light-gray` 上の `10px` 高の `999px` track と
  `--primary-black` の塗り。
- 実行 table はカード形状と標準 table 慣習を再利用します。

### Workflow Settings

`features/projects/components/workflow-settings-view`:

- 最大 `960px` の単一カラム `panelGrid`。
- `workflow-meta-panel`: ラベル/値ペアの flex chip 行、padding `12px 16px`。
- `frontmatter-section` と `workflow-body-section`: `18px` 見出しを持つ
  標準カード。
- `frontmatter-table`: radius `8px` のラッパー内の `1px` border 付き table、
  等幅キーに `padding-inline-start: calc(var(--frontmatter-depth, 0) * 20px)`
  を適用してネストキーを視覚的にインデント。ヘッダーセルは
  `--extra-light-gray` を使用。

### Issue レイアウト

`src/components/layout/issue/index.module.css` は issue のサブページシェルを
所有します。

- `max-width: 1040px`、行 gap `14px`、戻りリンクは `--primary-black`。
- tab はグローバルヘッダーと同じアクティブ rule パターン
  (`box-shadow: inset 0 -2px 0 var(--primary-black)`) を使用。

## フォーム

フォーム field の慣習は dialog と settings フォームで確認できます。

- フィールドグループ: gap `6px` の grid。ラベルは `13px / 700`、
  `--dark-gray` (settings は `--muted` を使用)。
- input/select/textarea: `1px` `--border`、radius `6px`、`--primary-black`
  テキスト、padding `10px`、`400` weight、`font: inherit`。
- textarea は `resize: vertical`。
- バリデーションメッセージ: `13px`、`--danger`、margin なし。
- 2 カラムフォームレイアウト (add-issue dialog の `formGrid`) は `900px`
  未満で 1 カラムに崩れます。
- 送信行: gap `8px` で末尾揃えの flex 行。primary submit は黒、
  secondary ボタンは既定の中立な見た目を保ちます。

無効なコントロールは opacity を `0.6` に下げ、`cursor: not-allowed` を
使用します。

## ステータスと優先度の意味

- priority "high" は警告のアンバー色として読めます。
- priority "normal" は情報のスカイブルーとして読めます。
- priority "low" は中立のグレーとして読めます。
- status `backlog`、`cancelled`、`duplicate`、`muted` はすべて中立のグレー。
- status `ready` は成功のグリーン。
- status `in-progress` は情報のスカイブルー。
- status `review` は警告のアンバー。
- status `done` は確認のグリーン (ready より少し濃い)。
- status `blocked` は警報のオレンジ。
- status `failed` は危険のレッド。

issue table の実行行は `statusToneClassName(issue.status)` を適用し、
各行が自分の `--status-color` を持つようにします。パレットを複製する
のではなく、行に status tone を継承させたい場所では、このヘルパーを
再利用してください。

## レスポンシブブレークポイント

このコードベースは統一されたスケールではなく、場当たり的な
`max-width` ブレークポイントを使用しています。ルールを追加するときは、
新しいブレークポイントを発明せず、最も近い既存のブレークポイントを
選んでください。

| ブレークポイント | 利用箇所 |
| ---------------- | -------- |
| `1060px`         | dashboard のメトリック/詳細 grid が 2 カラムに崩れる。 |
| `900px`          | app shell が 1 カラムに崩れる。header 行が縦に積む。issue card と詳細 meta grid が崩れる。dialog padding が縮む。 |
| `860px`          | table view ツールバーが縦に積む。フィルタートリガーが幅広に。pagination が `space-between` に切り替わる。 |
| `720px`          | toast が全幅に。conversation ページのイベントヘッダーが縦に積む。dashboard メトリック grid が 1 カラムに。issue tab が水平スクロール可。run row 内 grid が崩れる。 |
| `640px`          | issue 詳細タイトルが `22px` に縮小。コメントヘッダーが崩れる。 |

## アクセシビリティのパターン

- `Switch`、ソートボタン、コピーボタン、context menu アイテム、toast には、
  多くの場合 i18n キー経由で意味のある `aria-label` を付けます。
- context menu には `aria-haspopup`、`aria-expanded`、`aria-controls`、
  `role="menu"` / `role="menuitem"` を適用します。
- toast は各アイテムを `role="status"` で、スタックを
  `aria-live="polite"` でラップします。
- `IconProxy` 経由で描画されるアイコンは `aria-hidden="true"` かつ
  `focusable="false"` なので、ラベルは外側のインタラクティブな要素に
  付ける必要があります。
- フォーカススタイルは `outline` を消すのではなく `outline` または
  `box-shadow` のリングを使います。`focus` スタイルをカスタマイズする
  ときは、必ず `focus-visible` のルールも維持してください。

## Storybook

`src/components/**/index.tsx` および `src/features/**/index.tsx` 配下の
すべてのコンポーネントは、`index.stories.tsx` を同梱しなければなりません。

所有関係に一致するタイトルを使ってください。

- `src/components/ui` は `UI/...`。
- `src/components/layout` は `Layout/...`。
- `src/features/<feature>` は `Features/<Feature>/...`。

Storybook 専用のフィクスチャは `src/stories` 配下に置きます。プロダクション
コードは Storybook ヘルパーを import してはいけません。

## 新しい UI を追加するとき

1. 新しい部品が UI primitive なのか、layout shell の関心なのか、
   feature を意識したコンポーネントなのかを判別します。配置ルールは
   [design.ja.md](design.ja.md) に従ってください。
2. 上記の既存トークン、surface、パターンを再利用します。新しいトークンや
   shadow を追加するのは、ドキュメント化された値のどれもその役割に
   合わないときだけです。
3. `index.tsx`、`index.module.css`、`index.stories.tsx` を併置します。
4. 正しい所有関係のタイトルで Storybook エントリを書き、新しい
   コンポーネントをカバーします。
5. 新しい primitive、新しい surface パターン、新しいトークンの役割を
   導入したら、このドキュメントを更新してください。
