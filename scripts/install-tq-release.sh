#!/bin/sh
set -eu

kind="${1:-release}"
requested_tag="${2:-}"
repo="${TQ_RELEASE_REPO:-version-1/tasq}"
install_dir="${TQ_INSTALL_DIR:-$HOME/.local/bin}"
install_name="${TQ_INSTALL_NAME:-tq}"

# Release installation is non-interactive; fail instead of waiting for gh to prompt.
export GH_PROMPT_DISABLED="${GH_PROMPT_DISABLED:-1}"

require_command() {
	if ! command -v "$1" >/dev/null 2>&1; then
		echo "$1 is required" >&2
		exit 1
	fi
}

resolve_platform() {
	os="$(uname -s | tr '[:upper:]' '[:lower:]')"
	arch="$(uname -m)"

	case "$os" in
		darwin|linux) ;;
		*) echo "unsupported OS: $os" >&2; exit 1 ;;
	esac

	case "$arch" in
		amd64|x86_64) arch="amd64" ;;
		arm64|aarch64) arch="arm64" ;;
		*) echo "unsupported architecture: $arch" >&2; exit 1 ;;
	esac

	printf '%s_%s\n' "$os" "$arch"
}

resolve_tag() {
	if [ -n "$requested_tag" ]; then
		printf '%s\n' "$requested_tag"
		return
	fi

	case "$kind" in
		release)
			gh release list --repo "$repo" --exclude-drafts --exclude-pre-releases --limit 1 --json tagName --jq '.[0].tagName // ""'
			;;
		prerelease)
			gh release list --repo "$repo" --exclude-drafts --limit 100 --json tagName,isPrerelease --jq 'map(select(.isPrerelease)) | .[0].tagName // ""'
			;;
		*)
			echo "usage: $0 release|prerelease [tag]" >&2
			exit 1
			;;
	esac
}

sha256_file() {
	if command -v shasum >/dev/null 2>&1; then
		shasum -a 256 "$1" | awk '{print $1}'
		return
	fi
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$1" | awk '{print $1}'
		return
	fi
	echo "shasum or sha256sum is required" >&2
	exit 1
}

require_command gh
require_command tar
require_command awk
require_command find
require_command grep
require_command mktemp

platform="$(resolve_platform)"
tag="$(resolve_tag)"

if [ -z "$tag" ]; then
	echo "no $kind found for $repo" >&2
	exit 1
fi

tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT INT TERM

asset_pattern="tasq_*_${platform}.tar.gz"
gh release download "$tag" --repo "$repo" --pattern "$asset_pattern" --pattern checksums.txt --dir "$tmp_dir" --clobber

asset_path="$(find "$tmp_dir" -maxdepth 1 -type f -name "$asset_pattern" | head -n 1)"
if [ -z "$asset_path" ]; then
	echo "asset not found for pattern: $asset_pattern" >&2
	exit 1
fi

asset_name="$(basename "$asset_path")"
expected_sha="$(grep " $asset_name\$" "$tmp_dir/checksums.txt" | awk '{print $1}')"
actual_sha="$(sha256_file "$asset_path")"

if [ "$expected_sha" != "$actual_sha" ]; then
	echo "checksum mismatch for $asset_name" >&2
	echo "expected: $expected_sha" >&2
	echo "actual:   $actual_sha" >&2
	exit 1
fi

extract_dir="$tmp_dir/extracted"
mkdir -p "$extract_dir"
tar -xzf "$asset_path" -C "$extract_dir"

for executable in tq issue-tracker orchestrator web; do
	if [ ! -f "$extract_dir/$executable" ]; then
		echo "archive does not contain $executable" >&2
		exit 1
	fi
done

mkdir -p "$install_dir"
cp "$extract_dir/tq" "$install_dir/$install_name"
chmod 0755 "$install_dir/$install_name"
for executable in issue-tracker orchestrator web; do
	cp "$extract_dir/$executable" "$install_dir/$executable"
	chmod 0755 "$install_dir/$executable"
done

printf "installed %s from %s to %s\n" "$install_name" "$tag" "$install_dir/$install_name"
