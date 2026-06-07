#!/bin/sh
set -eu

script_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
. "$script_dir/lib.sh"

service="${1:-}"
mode="${2:-wait}"

if ! name="$(dev_service_name "$service")" || ! url="$(dev_service_health_url "$service")"; then
	echo "Usage: scripts/dev/ready.sh <issue-tracker|web> [--check]" >&2
	exit 2
fi

if [ "$mode" = "--check" ]; then
	curl -fsS "$url" >/dev/null 2>&1
	exit $?
fi

attempt=1
while [ "$attempt" -le 30 ]; do
	if curl -fsS "$url" >/dev/null 2>&1; then
		exit 0
	fi
	attempt=$((attempt + 1))
	sleep 1
done

echo "$name is not ready"
exit 1
