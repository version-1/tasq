#!/bin/sh
set -eu

script_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
. "$script_dir/lib.sh"

service="${1:-all}"

stop_one() {
	if ! service_name="$(dev_service_name "$1")"; then
		echo "unknown service: $1" >&2
		return 2
	fi
	pkill -f "$(dev_service_air_pattern "$service_name")" 2>/dev/null || true
}

case "$service" in
	all)
		for service_name in $(dev_each_service); do
			stop_one "$service_name"
		done
		;;
	*)
		stop_one "$service"
		;;
esac
