#!/bin/sh
set -eu

script_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
. "$script_dir/dev-lib.sh"

service="${1:-all}"
log_dir="$(dev_log_dir)"

mkdir -p "$log_dir"

case "$service" in
	all)
		for service_name in $(dev_each_service); do
			touch "$log_dir/$(dev_service_log_file "$service_name")"
		done
		tail -f "$log_dir"/*.log
		;;
	*)
		if ! service_name="$(dev_service_name "$service")"; then
			echo "Usage: scripts/dev/logs.sh [issue-tracker|tracker|orchestrator|web]" >&2
			exit 2
		fi
		log_file="$log_dir/$(dev_service_log_file "$service_name")"
		touch "$log_file"
		tail -f "$log_file"
		;;
esac
