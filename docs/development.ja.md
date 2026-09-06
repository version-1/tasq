# Development Workflow

## Local Development Quick Start

Local development では `make` 経由で Docker Compose を使います。

```sh
make dev-up
```

このコマンドは `dev` container と OpenAPI UI を起動し、`dev` container 内で issue-tracker、orchestrator、Web UI を起動します。Docker Compose が host ports を自動割り当てし、割り当てられた URLs を表示します。

URLs を再表示します。

```sh
make dev-ports
```

環境を停止します。

```sh
make dev-down
```

利用可能な development commands を一覧します。

```sh
make help
```

### Linux/WSL2 Sandbox Prerequisite

Codex は Linux sandboxing に Bubblewrap を使います。Dev image は `bubblewrap` を install しますが、Codex の sandboxed command を安定して動かすには Linux / WSL2 host 側でも unprivileged user namespace creation が許可されている必要があります。Codex が `bwrap: No permissions to create a new namespace` を報告する場合は、image に package がないだけではなく、host または Docker runtime の capability issue として扱います。

## Worktree 運用

作業はリポジトリ直下の `.worktrees/1` から `.worktrees/n` までの連番ディレクトリに worktree を作成し、そこを作業ディレクトリとして使います。

同時に複数の作業を行う場合は、作業単位ごとに未使用の番号を割り当て、既存の worktree と混在させないでください。

例:

```sh
git worktree add .worktrees/1 <branch>
git worktree add .worktrees/2 <branch>
```

作業開始前に `.worktrees/` 配下の既存番号を確認し、次に空いている番号を使います。

## Task Flow

タスクの開始から引き継ぎまでは、この流れを使います。

1. タスクの範囲、期待される成果物、影響しそうなファイルまたはコンポーネントを確認します。
2. 新しい task branch を作成する作業では、`cmd-start-branch` で作業を開始します。
3. 編集前に current branch と working tree を確認します。

   ```sh
   git status --short --branch
   ```

4. code または documentation を変更する前に、関連する設計ドキュメントと workflow document を読みます。

   - [docs/design.md](design.md)
   - 変更対象 area の component-level workflow document

5. 既存のコンポーネント境界と所有関係に合わせて、焦点を絞った変更を行います。
6. contract、setup、developer workflow に影響する場合は、関連 documentation と generated artifacts も更新します。
7. まず最小限で有用な検証を実行し、shared behavior、contract、persistence、user-facing flow に影響する変更では検証範囲を広げます。
8. Pull Request を作成する前に final diff を確認します。

   ```sh
   git diff
   git status --short
   ```

9. `cmd-create-pr` を使って、タスクの Pull Request を作成または更新します。
10. Pull Request URL、変更したファイル、実行した検証、残っているリスクまたはスキップした確認を簡潔にまとめて引き継ぎます。

## GitHub 操作

Pull Request の確認、作成、状態確認などの GitHub 操作には GitHub CLI (`gh`) を使います。

## Verification

標準の Compose-backed checks を実行します。

```sh
make dev-test
```

Go services と Web UI の両方に影響する変更を handoff する前は、broader build check を実行します。

```sh
make dev-build
```

## API 生成

Frontend API client の生成には `generate:api` を使います。

Generated Web UI API clients は、次の service を対象にします。

- Issue Tracker: `docs/openapi/issue-tracker.yml` から `cmd/web/frontend/src/lib/generated/issue-tracker.ts` を生成します。
- Orchestrator: `docs/openapi/orchestrator.yml` から `cmd/web/frontend/src/lib/generated/orchestrator.ts` を生成します。

API contract を変更するときは、同じ変更内で次の artifacts を更新します。

1. `docs/openapi` 配下の owning service OpenAPI document。
2. `cmd/web/frontend` で `npm run generate:api` を実行して生成する Web UI API clients。

変更した endpoint が standalone frontend development で使われる場合は、`cmd/web/frontend/src/mocks` 配下の MSW handlers と fixtures も更新します。

## Documentation Updates

Documentation を更新するときは、英語版の `.md` と日本語版の `*.ja.md` を同期させます。

- 同じ content change に対して両方のファイルを更新します。
- 片方の language file しかない場合は、対応するもう片方を追加します。
- 英語版と日本語版の links を揃えます。
- 日本語版の `*.ja.md` は `AGENTS.md` から link しなくてかまいません。`AGENTS.md` では英語版の `.md` を link します。
- `AGENTS.md`、`WORKFLOW.md`、`cmd/web/frontend/AGENTS.md` のようなエージェント指示ファイルは、この英日ペア規約の対象外とします。単一言語の運用ファイルとして扱います。
- ADR は historical decision record として扱います。typo や broken link のような明らかな mechanical fix を除き、後続 decision に合わせて過去 ADR を書き換えません。新しい decision が過去 ADR を変更または制約する場合は、新しい ADR 側にその変更と関係を書きます。

## Repository Documentation

このセクションは、repository documentation の唯一の索引です。この一覧を他の場所に複製せず、このセクションへリンクしてください。

- [WORKFLOW.md](../WORKFLOW.md): orchestrator が使う Symphony runtime workflow contract。
- [docs/design.md](design.md): system architecture と service boundaries。
- [docs/design/deployment.ja.md](design/deployment.ja.md): release tag 作成、GitHub Actions、GoReleaser の deployment flow。
- [docs/references/makefile.ja.md](references/makefile.ja.md): Makefile targets、variables、local development command reference。
- [cmd/issue-tracker/WORKFLOW.md](../cmd/issue-tracker/WORKFLOW.md): issue-tracker development workflow。
- [cmd/orchestrator/WORKFLOW.md](../cmd/orchestrator/WORKFLOW.md): orchestrator development workflow。
- [cmd/web/WORKFLOW.md](../cmd/web/WORKFLOW.md): Web UI development workflow。
- [docs/design/web.md](design/web.md): Web UI structure と styling conventions。
- [docs/openapi/issue-tracker.yml](openapi/issue-tracker.yml): issue-tracker OpenAPI contract。
- [docs/symphony/README.md](symphony/README.md): Symphony documentation index。
- [docs/symphony/SPEC.md](symphony/SPEC.md): Symphony orchestration and runner specification。
- [docs/symphony/DEVIATIONS.md](symphony/DEVIATIONS.md): Symphony specification からの intentional deviations。

## Operational Notes

- Runtime state と SQLite files は repository の `.tasq/` 配下に作成され、git からは無視されます。
- Compose は Go module/build caches、`cmd/web/frontend/node_modules`、Codex login state、GitHub CLI login state を named Docker volumes に保存します。
- Orchestrator は各 project の `WORKFLOW.md` から Symphony-oriented runtime settings と issue ごとの agent prompt を解決し、fallback として `$TQ_HOME/WORKFLOW.md` を使います。
- Web UI は Go server の proxy paths `/tracker/*` と `/orchestrator/*` 経由で local backends を呼び出します。
- Codex を device auth で認証し、authentication を `codex-home` Docker volume に永続化するため、初回に `make dev-codex-login` を実行します。
- GitHub CLI を認証し、Git が `gh` を HTTPS credential helper として使うよう設定し、credential を `gh-config` Docker volume に永続化するため、初回に `make dev-gh-login` を実行します。Dev container から push する場合は HTTPS Git remote を使います。
- Codex または GitHub access が必要な agent workflow を実行する前に、`make dev-codex-status` と `make dev-gh-status` で dev container が認証済みであることを確認します。

## tq CLI

利用者向けの [CLI リファレンス](site/i18n/ja/docusaurus-plugin-content-docs/current/reference/cli-reference.md)を `tq` の正典とします。Compose 固有の 1 つの実行方法と必要条件は、[Compose 開発環境での `tq`](references/tq.ja.md)を参照してください。
