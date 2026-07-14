---
name: fx-strategy-optimizer
description: 指定されたプロフィットファクター目標に対して、fx-autotrader の新しい strategy と通貨ペア別 profile を実装、登録、バックテスト評価する。strategy/profile をバックテストで改善したい、PF・勝率・ドローダウンに合わせて entry 条件を調整したい、パフォーマンスレポートを求められた場合に必ず使う。コード変更またはバックテストの前に、目標 PF、対象期間、最低取引数、最大ドローダウン条件、CurrencyPair を確認する。
---

# FX Strategy Optimizer

新しい `fx-autotrader` strategy と profile を保守可能な形で作成し、要求された性能条件を満たすまで反復評価する。ただし、バックテストは将来の収益性を保証しない研究結果として扱う。

## 必須入力

実装前に次を集める。不足・曖昧な値だけを確認する。

| 入力 | 受け付ける形式 | 補足 |
| --- | --- | --- |
| 目標 PF | 正の数 | holdout 期間で満たす閾値。 |
| 対象期間 | 両端を含む日付または RFC3339 timestamp | 暦日指定はリポジトリの JST FX 市場境界へ変換し、使用時刻を報告する。 |
| 最低取引数 | 正の整数 | PF と同じ評価範囲に適用する。 |
| 最大 DD 条件 | 金額、または基準値を伴う百分率 | 保存される値は損失額。百分率には backtest investment を基準として明示する。 |
| CurrencyPair | `usdjpy` などの canonical key | `USD_JPY` / `USD/JPY` は CLI・パスへ使う前に canonical key へ変換する。 |

名前の希望も確認する。なければ、説明的な kebab-case profile 名と対応する snake_case strategy 名を選ぶ。

## リポジトリ契約

`fx-autotrader` の worktree で作業する。現在の契約の正として、変更前に次を読む。

- `AGENTS.md`
- `docs/guide/strategy.md`
- `docs/guide/profile.md`
- `docs/guide/run-backtest.md`
- `internal/strategy/base.go`
- `internal/profile/loader.go`

責務を分ける。

- entry 判断、time/position/regime filter、signal 構築は `strategy.BaseStrategy` を使い `internal/strategy/...` に置く。
- sizing、利確・損切り、銘柄制約は `internal/profile/profiles/<currency-pair>/<profile-name>.json` に置く。
- 新しい strategy 名を `internal/profile/loader.go` に登録する。JSON profile だけでは新しい strategy は動かない。
- live runtime 設定、credential、historical price、既存 profile を、バックテストを良く見せる目的で変更しない。

## 手順

### 1. 再現可能なベースラインを作る

1. runbook の query で対象 pair・期間の historical rate を確認する。
2. 近い既存 strategy/profile と、そのコード、order parameter、結果を記録する。
3. 対象期間を時系列で最適化区間と最終 holdout 区間に分ける。既定は 70% / 30% とし、可能なら FX 市場週境界を保つ。両区間で最低取引数を満たせない短い期間では、信頼できる holdout 評価ができないことを説明し、期間拡大または in-sample のみの研究を明示的に選んでもらう。
4. investment、sizing method、strategy 以外の runtime 設定は比較中に固定し、レポートに記載する。

同じ過去の価格列に条件を繰り返し合わせ込むことを避けるため、目標達成は holdout が全条件を満たした場合に限る。

### 2. 焦点を絞った候補を実装する

1. 最も近い strategy の条件から始め、捉える・除外する相場挙動を短い仮説として書く。
2. `NewStrategy` が `*strategy.BaseStrategy` を返す小さな package を追加する。既存 filter interface を組み合わせ、reference price の挙動が異なるときだけ `CommandBuilder` を使う。
3. `internal/profile/loader.go` の `resolveStrategy` に登録する。
4. `internal/profile/profiles/<currency-pair>/<profile-name>.json` を追加し、正確な説明を付ける。
5. strategy の entry/no-entry、必要なら warmup、loader をテストし、バックテスト前に関連 Go test を実行する。

1 回の反復では 1 つの首尾一貫した仮説だけを変える。説明できない広範なパラメータ総当たりは行わない。

### 3. バックテストと指標収集

バックテストの前に、対象 worktree の Compose project 名を決める。別の作業スペースの network・container・volume と衝突させないため、リポジトリ全体に固定した値ではなく、その worktree に隔離された project 名を明示する。Makefile はカレントディレクトリ名から既定値を作るが、実際に使った値をレポートへ残す。

その project 名と独立した profile 名を指定して runbook のコマンドを実行する。

```bash
COMPOSE_PROJECT_NAME=<worktree-compose-project-name> make backtest \
  FROM=<from-rfc3339> \
  TO=<to-rfc3339> \
  PROFILE=<profile-name> \
  CURRENCY_PAIR=<canonical-currency-pair>
```

完了出力の `result_id` を保存し、`docs/guide/run-backtest.md` の手順で run / slice / trade を取得する。最低限、取引数、勝率、PF、gross profit/loss、期待値、最終損益、最大 DD、定義される場合の recovery factor、対象期間、pair、profile、strategy、result ID、コード revision を収集する。

損失取引がなく PF が欠損する場合は、合格とせず未定義と扱う。0 取引、データ欠損、実行失敗は不合格評価とし、条件変更前に原因を調べる。

### 4. 判定と反復

holdout を次の順で判定する。

1. 最低取引数
2. 最大 DD 条件
3. 目標 PF

条件未達時は filter trace、個別取引、ベースライン差分から、次の 1 つの仮説を選ぶ。各候補で関連テスト、最適化区間、holdout を順に再実行する。最良候補と根拠を残し、既存 profile を上書きしない。

目標が未達でも、根拠ある次の仮説があれば反復を続ける。改善が止まる、データ不足、条件間の衝突がある場合は、証拠と最も低リスクな次の実験を報告して止める。要求条件を黙って緩めて成功としない。

## 完了条件

- strategy が `strategy.BaseStrategy` で構築され、profile loader に登録されている。
- 新しい profile が指定 pair の canonical key で load できる。
- 関連テストと `go test ./...` が通る。無関係な既存失敗は根拠付きで分離する。
- holdout に有効データがあり、PF・取引数・最大 DD の全条件を満たす。
- live 設定、secret、historical price は変更していない。

## 最終レポート

以下の構造を使う。失敗した反復も再現できるよう短く残す。

```markdown
# <strategy/profile name> backtest report

## Decision
- Status: achieved | not achieved | blocked
- Requested criteria: PF >= <target>, trades >= <minimum>, max DD <condition>
- Holdout verdict: pass | fail

## Artifacts
- Strategy: <path>
- Profile: <path>
- Registration and tests: <paths>

## Reproducibility
- Currency pair: <canonical key>
- Optimization period: <from> to <to>
- Holdout period: <from> to <to>
- Command: `<exact command>`
- Compose project name: `<worktree-compose-project-name>`
- Result IDs: <optimization and holdout IDs>
- Code revision: <commit or dirty-worktree note>

## Metrics
| Scope | Trades | Win rate | PF | Max DD | Final profit | Expected value | Verdict |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| Baseline holdout | | | | | | | |
| Candidate holdout | | | | | | | |

## Iterations and rationale
1. <hypothesis, change, and result>

## Limitations and next steps
<data coverage, sample size, regime sensitivity, or the next safe experiment>
```
