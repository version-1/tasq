# Deployment Flow

Japanese counterpart: [deployment.ja.md](deployment.ja.md).

Tasq deploys release artifacts by pushing version tags. The local operator creates a tag with `make prerelease` or `make release`, GitHub Actions runs the Release workflow for `v*` tags, and GoReleaser publishes the GitHub Release artifacts.

## Flow Overview

1. The operator runs a release Make target from a clean working tree.
2. `Makefile` delegates tag validation and creation to `scripts/release.sh`.
3. The script creates a local Git tag and pushes it to `origin`.
4. `.github/workflows/release.yml` runs on pushed tags matching `v*`.
5. The workflow builds the Web frontend, runs Go tests, and then runs GoReleaser.
6. GoReleaser builds `tq` plus the managed service executables, packages them into archives, checksums them, and creates or updates the GitHub Release.

## Prerelease

Use prereleases for validation builds that should not be treated as formal stable releases.

```sh
make prerelease
make prerelease version=v0.3.0
```

The prerelease target:

- requires a clean working tree;
- accepts an optional `version` value in `vX.Y.Z` format to use as the prerelease base;
- rejects suffixed or invalid `version` values;
- when `version` is omitted, finds the latest reachable formal SemVer tag such as `v0.1.0`;
- when `version` is omitted and no formal tag exists, falls back to `v0.0.0`;
- creates a tag in the form `v0.3.0-dev.YYYYMMDDTHHmm.<short-sha>`;
- pushes the tag to `origin`.

GoReleaser has `release.prerelease: auto`, so tags with a SemVer prerelease segment such as `-dev...` become GitHub prereleases.

## Formal Release

Use formal releases for stable versions that should be presented as normal GitHub Releases.

```sh
make release version=v0.1.1
```

The release target:

- requires `version` to be set;
- accepts only `vX.Y.Z` tags with no prerelease suffix;
- requires the current branch to be `main`;
- requires a clean working tree;
- rejects an already existing tag;
- creates and pushes the tag to `origin`.

## Configuration

The Makefile exposes these release variables:

| Variable | Default | Purpose |
|---|---|---|
| `RELEASE_BRANCH` | `main` | Branch required for formal releases. |
| `RELEASE_REMOTE` | `origin` | Remote that receives release tags. |

Override them only when validating release behavior outside the normal repository flow.

## Release Artifacts

Each platform archive contains `tq`, `issue-tracker`, `orchestrator`, and `web`. The `web`
binary embeds `cmd/web/frontend/dist`, so the Release workflow must run the frontend production
build before GoReleaser.

`make install-tq` and `make install-tq-prerelease` install `tq` using `TQ_INSTALL_NAME`, and install
the service executables next to it with their fixed names. `tq service start` looks for those sibling
executables before falling back to source-based `go run` service startup.

Installed users can run `tq update` to install the latest formal release into the fixed user install
location, apply local migrations, and restart services. `tq update --tag <tag>` accepts both release
and prerelease tags for targeted installs.

## Failure Handling

If validation fails before the tag is created, fix the reported issue and rerun the same command.

If the local tag is created but the push fails, inspect the tag first:

```sh
git show <tag>
git push origin <tag>
```

If the wrong tag was created locally and was not pushed, delete only that local tag:

```sh
git tag -d <tag>
```

If the wrong tag was already pushed, do not delete or rewrite it without coordinating with repository maintainers. A pushed tag may already have triggered GitHub Actions and created release artifacts.

## Verification Before Handoff

For release-flow changes, verify the non-destructive paths first:

```sh
sh -n scripts/release.sh
make help
make release
make release version=0.1.1
make release version=v0.1.1-beta
make prerelease version=0.3.0
make prerelease version=v0.3.0-dev.20260608T0000.abc1234
make -n prerelease
make -n prerelease version=v0.3.0
make -n release version=v0.1.1
```

Run real `make prerelease` or `make release version=...` only when intentionally creating and pushing a release tag.
