---
name: tasq-permission-check
description: tq の自律実行に必要な Codex のコマンド許可、Rules、requestApproval、blocked 実行、execpolicy を確認するときに使う。Tasq のライフサイクルで使うコマンドを列挙し、Rules ファイルに対して順番に検査して、承認なしで実行できないコマンドを報告する。
compatibility: `codex execpolicy check`、`jq`、Codex Rules ファイルが必要。
---

# Tasq Permission Check

## 目的

Tasq の課題を自律実行する前に、必要なコマンドが指定した Codex Rules ファイルで許可されていることを確認します。通常の実行が `requestApproval` や `blocked` で止まる前に、足りない Rule を見つけるためのスキルです。

## 安全な境界

これはポリシーの検査であり、実際のワークフローを実行するものではありません。権限確認のためだけに、プロジェクト、課題、コメント、worktree、コミット、push、pull request を作成してはいけません。付属スクリプトは代表コマンドを `codex execpolicy check` へ渡すだけで、コマンド自体は実行しません。

ユーザーから別途依頼されない限り、このスキルで Rules の編集、Codex 設定の変更、コマンドの承認は行いません。

## 入力

Rules ファイルが明らかでない場合はユーザーに尋ねます。プロジェクトに `codex/rules/tasq-dev.rules` がある場合はそれを優先し、ない場合はユーザーが指定したパスを使います。結果は検査した Rules ファイルにだけ適用され、未知の managed policy やユーザー単位のポリシーまで保証するものではありません。

## 手順

1. [必要なコマンド一覧](references/required-commands.ja.md)を読みます。通常の Tasq ライフサイクルで、状態を確認するコマンドとローカルまたはリモートを変更するコマンドを分けています。
2. `codex` と `jq` が利用できることを確認します。どちらかがない場合は、不足している依存関係を報告して止めます。
3. リポジトリのルートで付属チェッカーを実行します。

   ```sh
   bash .agents/skills/tasq-permission-check/scripts/check-execpolicy.sh \
     --rules codex/rules/tasq-dev.rules
   ```

4. 結果を順番に読みます。`allow` は、検査した Rules ファイルではそのコマンドに `requestApproval` が想定されないことを示します。`prompt`、`forbidden`、一致する Rule がない場合は許可不足です。確認のために実コマンドを実行してはいけません。
5. 許可不足ごとに、代表コマンド、decision、ライフサイクル上の目的を報告します。ユーザーには `codex/rules/tasq-dev.rules` を確認し、ワークフローに必要な最小の Rule だけを追加するよう案内します。
6. ユーザーが実行確認を求めた場合は、`tq version`、`tq project list`、`git status --short`、`gh pr status` など、変更を加えないコマンドだけを実行します。先にポリシーを確認してください。プロジェクト作成、課題更新、コメント、worktree の作成・削除、コミット、push、PR 作成を権限確認として実行してはいけません。

## レポート形式

次の形式を使います。

```md
# Tasq Permission Check

- Rules file: `<path>`
- Result: PASS | GAPS FOUND

## Allowed Commands

| Lifecycle step | Representative command | Decision |
| --- | --- | --- |

## Permission Gaps

| Lifecycle step | Representative command | Decision | Required action |
| --- | --- | --- | --- |

## Notes

- `allow` は、検査したコマンドに `requestApproval` が想定されないというポリシー上の予測です。
- チェッカーは変更を加えるライフサイクルコマンドを実行していません。
```

すべてが許可されている場合は、検査したポリシーが一覧の Tasq ライフサイクルをカバーしていると報告します。許可不足がある場合は、自律実行の準備ができていないことを伝え、ユーザーから依頼されない限り Rules を変更しません。
