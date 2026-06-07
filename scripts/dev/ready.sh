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

air_pattern="$(dev_service_air_pattern "$name")"
log_file="$(dev_log_dir)/$(dev_service_log_file "$name")"
pid_file="$(dev_service_pid_file "$name")"

health_ready() {
	curl -fsS --connect-timeout 1 --max-time 2 "$url" >/dev/null 2>&1
}

print_log_tail() {
	if [ -f "$log_file" ]; then
		echo "last $name log lines:" >&2
		tail -n 40 "$log_file" >&2
	fi
}

service_process_alive() {
	if [ -f "$pid_file" ]; then
		pid="$(cat "$pid_file")"
		if [ -n "$pid" ]; then
			kill -0 "$pid" 2>/dev/null
			return $?
		fi
	fi
	pgrep -f "$air_pattern" >/dev/null 2>&1
}

if [ "$mode" = "--check" ]; then
	health_ready
	exit $?
fi

attempt=1
max_attempts="${DEV_READY_ATTEMPTS:-30}"
while [ "$attempt" -le "$max_attempts" ]; do
	if health_ready; then
		exit 0
	fi
	if ! service_process_alive; then
		echo "$name exited before it became ready" >&2
		print_log_tail
		exit 1
	fi
	if [ "$attempt" -eq 1 ] || [ $((attempt % 10)) -eq 0 ]; then
		echo "waiting for $name at $url ($attempt/$max_attempts)" >&2
	fi
	attempt=$((attempt + 1))
	sleep 1
done

echo "$name is not ready" >&2
print_log_tail
exit 1
