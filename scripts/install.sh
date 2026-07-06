#!/bin/sh
set -eu

repo="${TASQ_REPO:-version-1/tasq}"
version="${TASQ_VERSION:-}"
install_dir="${TASQ_INSTALL_DIR:-$HOME/.local/bin}"
required_bins="tq issue-tracker orchestrator web"

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

resolve_version() {
	if [ -n "$version" ]; then
		printf '%s\n' "$version"
		return
	fi

	curl -fsSLI -o /dev/null -w '%{url_effective}' "https://github.com/$repo/releases/latest" | sed 's#.*/##'
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

verify_checksum() {
	asset_name="$1"
	asset_path="$2"
	checksums_path="$3"

	expected_sha="$(grep " $asset_name\$" "$checksums_path" | awk '{print $1}')"
	if [ -z "$expected_sha" ]; then
		echo "checksum entry not found for $asset_name" >&2
		exit 1
	fi

	actual_sha="$(sha256_file "$asset_path")"
	if [ "$expected_sha" != "$actual_sha" ]; then
		echo "checksum mismatch for $asset_name" >&2
		echo "expected: $expected_sha" >&2
		echo "actual:   $actual_sha" >&2
		exit 1
	fi
}

verify_installed_binary() {
	bin="$1"
	source_path="$extract_dir/$bin"
	installed_path="$install_dir/$bin"
	expected_sha="$(sha256_file "$source_path")"
	actual_sha="$(sha256_file "$installed_path")"

	if [ "$expected_sha" != "$actual_sha" ]; then
		echo "installed binary checksum mismatch for $bin" >&2
		echo "expected: $expected_sha" >&2
		echo "actual:   $actual_sha" >&2
		exit 1
	fi
}

require_command awk
require_command chmod
require_command cp
require_command curl
require_command grep
require_command mkdir
require_command mktemp
require_command sed
require_command tar
require_command tr
require_command uname

platform="$(resolve_platform)"
tag="$(resolve_version)"

if [ -z "$tag" ]; then
	echo "could not resolve latest release for $repo" >&2
	exit 1
fi

archive="tasq_${tag#v}_${platform}.tar.gz"
base_url="https://github.com/$repo/releases/download/$tag"

tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT INT TERM

curl -fsSL "$base_url/$archive" -o "$tmp_dir/$archive"
curl -fsSL "$base_url/checksums.txt" -o "$tmp_dir/checksums.txt"
verify_checksum "$archive" "$tmp_dir/$archive" "$tmp_dir/checksums.txt"

extract_dir="$tmp_dir/extracted"
mkdir -p "$extract_dir"
tar -xzf "$tmp_dir/$archive" -C "$extract_dir"

for bin in $required_bins; do
	if [ ! -f "$extract_dir/$bin" ]; then
		echo "archive does not contain $bin" >&2
		exit 1
	fi
done

mkdir -p "$install_dir"
for bin in $required_bins; do
	cp "$extract_dir/$bin" "$install_dir/$bin"
	chmod 0755 "$install_dir/$bin"
	verify_installed_binary "$bin"
done

echo "installed tasq $tag binaries to $install_dir"
printf "verified installed tq sha256: %s\n" "$(sha256_file "$install_dir/tq")"
