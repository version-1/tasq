# Orchestrator Workflow

orchestrator は agent assignment と run state を所有します。issue-tracker から executable work item を claim し、自身の SQLite database に run を記録し、durable outbox から run event を配送します。

`cmd/orchestrator`、`internal/orchestrator`、`db/schema/orchestrator.sql` を変更するときは、この workflow を使います。

## Scope

- Run record、run attempt、run に紐づく claim token、outbox delivery state は orchestrator が所有します。
- orchestrator が issue state を直接変更しないようにします。
- issue 側へ伝える変更は run event を issue-tracker へ push して表現します。
- Real agent や workspace execution を追加する前も、retry-safe な outbox behavior を保ちます。

Service boundary は [../../docs/design.md](../../docs/design.md) を参照してください。

## Local Run

Queue と event boundary を確認するときは、issue-tracker と orchestrator を Compose で一緒に起動します。

```sh
make orchestrator-up
make dev-ports
```

Host-only development の場合は、先に issue-tracker を起動してから実行します。

```sh
go run ./cmd/orchestrator \
  -db tasq-orchestrator.sqlite \
  -issue-tracker http://localhost:8080
```

## Change Flow

1. Polling、claim、run creation、outbox delivery の変更は、実装上の責務を分けて扱います。
2. SQLite schema 変更は `db/schema/orchestrator.sql` に置きます。
3. 現在の simulated lifecycle は一時的な verification behavior として扱います。
4. 将来 real runner behavior を追加するときも、意図的な API 変更でない限り issue-tracker contract を安定させます。

## Verification

開発中は focused orchestrator tests を実行します。

```sh
go test ./internal/orchestrator/...
```

Claim、run、outbox delivery を変える behavior を渡す前に、broader checks を実行します。

```sh
go test ./...
```

Compose toolchain で確認するときは `make dev-test` を使います。

## Operational Notes

- orchestrator は配送前に run event を outbox へ書き込む必要があります。
- Delivery retry が issue-tracker state transition を二重適用してはいけません。
- Claim token は run event を 1 つの work item claim generation に紐づけます。
- Lease behavior は parallel orchestrator instances に対して安全である必要があります。
