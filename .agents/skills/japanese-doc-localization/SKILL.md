---
name: japanese-doc-localization
description: Improve Japanese technical documentation readability by translating general English terms into natural Japanese prose while preserving identifiers, config keys, protocol names, and normative keywords. Use when editing *.ja.md files, Japanese versions of English specs, localized developer docs, or reviews that complain Japanese docs are hard to read because they are just line-broken English.
---

# Japanese Doc Localization

## Purpose

Make Japanese technical documents read like Japanese, not like English text with Japanese particles.

Use this skill when editing Japanese documentation, especially `*.ja.md` files that mirror English source documents.

## Core Rule

Translate general terminology into natural Japanese prose.

Preserve only names that readers must match exactly in code, configuration, protocols, commands, APIs, states, domain terminology, or external products.

## Preserve As-Is

Keep these in English or code style:

- Configuration keys: `tracker.kind`, `workspace.root`, `codex.command`
- Domain terms that the project treats as names: Issue Tracker
- Enum values and state names: `Todo`, `Done`, `Human Review`
- Protocol and product names: Codex app-server, Linear GraphQL, RFC 2119
- Command names and flags: `codex app-server`, `--port`
- Field names when they are part of the contract: `thread_id`, `issue_id`, `turn_id`
- Normative keywords: `MUST`, `SHOULD`, `MAY`, `OPTIONAL`, `REQUIRED`

## Translate General Terms

Translate English words used as ordinary explanation:

| Avoid | Prefer |
| --- | --- |
| issue | 課題 |
| workspace | ワークスペース |
| workflow | ワークフロー |
| runtime | 実行時 |
| dispatch | 割り当てる / 割り当て |
| retry | 再試行 |
| cleanup | クリーンアップ |
| state | 状態 |
| terminal state | 終端状態 |
| active state | アクティブ状態 |
| agent | エージェント |
| coding agent | コーディングエージェント |
| prompt | プロンプト |
| observability | 可観測性 |
| dashboard | ダッシュボード |
| log | ログ |
| validation | 検証 |
| configuration | 設定 |
| default | 既定値 |
| failure | 失敗 |
| success | 成功 |

Prefer verbs that make the sentence Japanese:

- `dispatch issue` -> 課題を割り当てる
- `fetch candidate issues` -> 候補課題を取得する
- `validate config` -> 設定を検証する
- `emit log` -> ログを出力する
- `preserve workspace` -> ワークスペースを保持する

## Editing Workflow

1. Read the nearby English source when the Japanese file mirrors one.
2. Identify contract terms that must stay exact.
3. Translate ordinary explanatory terms into Japanese.
4. Rewrite fragments into complete Japanese sentences.
5. Keep headings and bullets scannable, but do not leave bullet items as English noun piles.
6. Verify that the meaning, requirements, and normative strength did not change.

## Style Guidance

Use Japanese sentence order instead of direct English calques.

Bad:

- `workflow config value に対する typed getter を提供する。`

Good:

- `ワークフロー設定値に対する型付き getter を提供する。`

Bad:

- `candidate issue を fetch し dispatch する。`

Good:

- `候補課題を取得し、割り当てる。`

Bad:

- `operator-visible observability を提供する。`

Good:

- `運用者が確認できる可観測性を提供する。`

## Review Checklist

- General terms are Japanese unless they are exact identifiers.
- Sentences are not just English clauses with Japanese endings.
- Code identifiers remain searchable and exact.
- Japanese and English versions still describe the same contract.
- No requirement strength changed: `MUST` is still `MUST`, `SHOULD` is still `SHOULD`.
