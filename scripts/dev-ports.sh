#!/bin/sh
set -eu

compose="${COMPOSE:-docker compose}"

print_port() {
	label="$1"
	service="$2"
	container_port="$3"
	spacing="$4"
	port="$($compose port "$service" "$container_port" 2>/dev/null | sed 's/.*://')"
	if [ -n "$port" ]; then
		printf "%s%shttp://localhost:%s\n" "$label" "$spacing" "$port"
	else
		printf "%s%snot running\n" "$label" "$spacing"
	fi
}

print_port "issue-tracker:" "dev" "8080" " "
print_port "orchestrator:" "dev" "8081" "  "
print_port "openapi:" "openapi" "8080" "       "
print_port "web:" "dev" "3000" "           "
