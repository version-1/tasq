---
id: recover-blocked-session
title: blocked になったセッションを復旧する
sidebar_position: 4
---

# blocked になったセッションを復旧する

Codex が blocked になった場合、ターミナル接続を失った場合、または別のシェルから
続けたい場合は、最初からやり直さず、保存済みセッションを resume します。

対象の課題に紐づいた正確な Codex thread を resume するため、まず Tasq Web UI
から確認します。

## thread ID を確認する

1. Web UI で対象の課題を開きます。
2. **Activity** タブを開きます。
3. 最新の実行行を見つけます。
4. その実行の thread ID をコピーします。

thread ID は、その課題に取り組んでいた Codex セッションを識別します。CLI から
resume するときにこの値を使います。

## CLI から resume する

repository checkout で `codex resume` を実行し、Activity タブでコピーした
thread ID を渡します。

```sh
codex resume <thread id>
```

Web UI に実行の thread ID がない場合は、Codex CLI のセッション選択画面に
fallback します。

```sh
codex resume
```

現在の作業ディレクトリの最新セッションを resume することもできます。

```sh
codex resume --last
```

## セッションを現在の状態に合わせる

resume 後は、blocked 中に変わったことを短く伝える復旧用プロンプトを送ります。

```text
Continue the previous task. First run git status, inspect the latest diff, and
then proceed from the current repository state.
```

resume したセッションは以前の transcript と plan history を保持するため、Codex は
以前のコンテキストを使えます。ただし、編集前には現在の working tree を再確認させて
ください。
