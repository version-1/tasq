#!/bin/sh
set -eu

release_branch="${RELEASE_BRANCH:-main}"
release_remote="${RELEASE_REMOTE:-origin}"

usage() {
	echo "Usage:"
	echo "  sh scripts/release.sh prerelease [v0.1.1]"
	echo "  sh scripts/release.sh release v0.1.1"
}

ensure_clean_working_tree() {
	context="$1"
	if [ -n "$(git status --porcelain)" ]; then
		echo "working tree must be clean before creating a $context tag"
		exit 1
	fi
}

ensure_tag_absent() {
	tag="$1"
	if git rev-parse -q --verify "refs/tags/$tag" >/dev/null; then
		echo "tag already exists: $tag"
		exit 1
	fi
}

latest_formal_release_tag() {
	git describe --tags --abbrev=0 --match 'v[0-9]*.[0-9]*.[0-9]*' --exclude '*-*' 2>/dev/null || true
}

validate_formal_version() {
	version="$1"
	if ! printf '%s\n' "$version" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$'; then
		echo "version must match vX.Y.Z without a prerelease suffix: $version"
		exit 1
	fi
}

create_prerelease() {
	base="${1:-}"
	if [ -n "$base" ]; then
		validate_formal_version "$base"
	fi

	if [ -z "$base" ]; then
		base="$(latest_formal_release_tag)"
	fi
	if [ -z "$base" ]; then
		base="v0.0.0"
	fi

	ensure_clean_working_tree "prerelease"

	timestamp="$(date -u '+%Y%m%dT%H%M')"
	short_sha="$(git rev-parse --short HEAD)"
	tag="$base-dev.$timestamp.$short_sha"

	ensure_tag_absent "$tag"

	echo "creating prerelease tag $tag"
	git tag "$tag"
	git push "$release_remote" "$tag"
}

create_release() {
	version="${1:-}"
	if [ -z "$version" ]; then
		echo "version is required. Usage: make release version=v0.1.1"
		exit 1
	fi

	validate_formal_version "$version"

	if [ "$(git branch --show-current)" != "$release_branch" ]; then
		echo "release tags must be created from $release_branch"
		exit 1
	fi

	ensure_clean_working_tree "release"
	ensure_tag_absent "$version"

	git tag "$version"
	git push "$release_remote" "$version"
}

command="${1:-}"
case "$command" in
	prerelease)
		shift
		create_prerelease "${1:-}"
		;;
	release)
		shift
		create_release "${1:-}"
		;;
	*)
		usage
		exit 1
		;;
esac
