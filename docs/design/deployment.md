# Deployment Flow

Japanese counterpart: [deployment.ja.md](deployment.ja.md).

Tasq deploys release artifacts by pushing version tags. The local operator creates a tag with `make prerelease` or `make release`, GitHub Actions runs the Release workflow for `v*` tags, and GoReleaser publishes the GitHub Release artifacts.

## Flow Overview

1. The operator runs a release Make target from a clean working tree.
2. `Makefile` delegates tag validation and creation to `scripts/release.sh`.
3. The script creates a local Git tag and pushes it to `origin`.
4. `.github/workflows/release.yml` runs on pushed tags matching `v*`.
5. The workflow runs Go tests and then runs GoReleaser.
6. GoReleaser builds archives, checksums them, and creates or updates the GitHub Release.

## Prerelease

Use prereleases for validation builds that should not be treated as formal stable releases.

```sh
make prerelease
```

The prerelease target:

- requires a clean working tree;
- finds the latest reachable formal SemVer tag such as `v0.1.0`;
- falls back to `v0.0.0` when no formal tag exists;
- creates a tag in the form `v0.1.0-dev.YYYYMMDDTHHmm.<short-sha>`;
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
make -n prerelease
make -n release version=v0.1.1
```

Run real `make prerelease` or `make release version=...` only when intentionally creating and pushing a release tag.
