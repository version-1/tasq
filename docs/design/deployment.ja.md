# Deployment Flow

English counterpart: [deployment.md](deployment.md).

Tasq は version tag の push を起点に release artifacts を deploy します。Operator が `make prerelease` または `make release` で tag を作成し、GitHub Actions が `v*` tag 用の Release workflow を実行し、GoReleaser が GitHub Release artifacts を公開します。

## Flow Overview

1. Operator が clean working tree から release Make target を実行します。
2. `Makefile` は tag validation と作成処理を `scripts/release.sh` に委譲します。
3. Script が local Git tag を作成し、`origin` に push します。
4. `.github/workflows/release.yml` が `v*` に一致する pushed tag で起動します。
5. Workflow が Go tests を実行し、その後 GoReleaser を実行します。
6. GoReleaser が archives と checksums を作成し、GitHub Release を作成または更新します。

## Prerelease

Prerelease は、formal stable release として扱わない validation build に使います。

```sh
make prerelease
```

`prerelease` target は次を行います。

- clean working tree を要求します。
- `v0.1.0` のような reachable な最新 formal SemVer tag を探します。
- formal tag がない場合は `v0.0.0` に fallback します。
- `v0.1.0-dev.YYYYMMDDTHHmm.<short-sha>` 形式の tag を作成します。
- tag を `origin` に push します。

GoReleaser には `release.prerelease: auto` を設定しています。そのため、`-dev...` のような SemVer prerelease segment を含む tag は GitHub prerelease になります。

## Formal Release

Formal release は、通常の GitHub Release として提示する stable version に使います。

```sh
make release version=v0.1.1
```

`release` target は次を行います。

- `version` の指定を要求します。
- prerelease suffix を含まない `vX.Y.Z` tag のみ受け付けます。
- current branch が `main` であることを要求します。
- clean working tree を要求します。
- 既存 tag と衝突する場合は拒否します。
- tag を作成して `origin` に push します。

## Configuration

Makefile は release 用に次の variables を公開します。

| Variable | Default | Purpose |
|---|---|---|
| `RELEASE_BRANCH` | `main` | Formal release で要求する branch。 |
| `RELEASE_REMOTE` | `origin` | Release tag を push する remote。 |

通常の repository flow 以外で release behavior を検証する場合だけ override してください。

## Failure Handling

Tag 作成前の validation で失敗した場合は、表示された問題を直して同じ command を再実行します。

Local tag は作成されたが push に失敗した場合は、まず tag を確認します。

```sh
git show <tag>
git push origin <tag>
```

誤った tag が local にだけ作成され、まだ push されていない場合は、その local tag だけを削除します。

```sh
git tag -d <tag>
```

誤った tag がすでに push されている場合は、repository maintainers と調整せずに削除や rewrite をしないでください。Pushed tag はすでに GitHub Actions を起動し、release artifacts を作成している可能性があります。

## Verification Before Handoff

Release flow を変更した場合は、まず破壊的でない経路を検証します。

```sh
sh -n scripts/release.sh
make help
make release
make release version=0.1.1
make release version=v0.1.1-beta
make -n prerelease
make -n release version=v0.1.1
```

実際の `make prerelease` または `make release version=...` は、release tag を作成して push する意図がある場合だけ実行します。
