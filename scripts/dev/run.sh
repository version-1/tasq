#!/bin/sh
set -eu

script_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
. "$script_dir/lib.sh"

service="${1:-}"
air_version="${AIR_VERSION:-v1.52.3}"
log_dir="$(dev_log_dir)"
download_log="${log_dir}/go-mod-download.log"

usage() {
	echo "Usage: scripts/dev/run.sh <issue-tracker|orchestrator|web>" >&2
}

if ! service_name="$(dev_service_name "$service")"; then
	usage
	exit 2
fi
config="$(dev_service_air_config "$service_name")"
log_file="${log_dir}/$(dev_service_log_file "$service_name")"
run_dir="$(dev_run_dir)"
pid_file="$(dev_service_pid_file "$service_name")"
delay="$(dev_service_start_delay "$service_name")"

mkdir -p "$log_dir" "$run_dir" /workspace/.tmp
pattern="$(dev_service_air_pattern "$service_name")"
pkill -f "$pattern" 2>/dev/null || true
rm -f "$pid_file"

if dev_service_needs_mod_download "$service_name"; then
	go mod download >>"$download_log" 2>&1
fi

nohup sh -c '
	delay="$1"
	air_version="$2"
	config="$3"
	log_file="$4"
	if [ "$delay" != "0" ]; then
		sleep "$delay"
	fi
	exec go run github.com/air-verse/air@"$air_version" -c "$config" >>"$log_file" 2>&1
' sh "$delay" "$air_version" "$config" "$log_file" >/dev/null 2>&1 &
echo "$!" >"$pid_file"
