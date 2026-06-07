#!/bin/sh

dev_log_dir() {
	tq_home="${TQ_HOME:-./.tasq}"
	printf "%s/system/log\n" "$tq_home"
}

dev_run_dir() {
	tq_home="${TQ_HOME:-./.tasq}"
	printf "%s/system/run\n" "$tq_home"
}

dev_service_name() {
	case "$1" in
		issue-tracker|tracker) echo "issue-tracker" ;;
		orchestrator) echo "orchestrator" ;;
		web) echo "web" ;;
		*) return 1 ;;
	esac
}

dev_each_service() {
	printf "%s\n" issue-tracker orchestrator web
}

dev_service_air_config() {
	case "$(dev_service_name "$1")" in
		issue-tracker) echo ".air.issue-tracker.toml" ;;
		orchestrator) echo ".air.orchestrator.toml" ;;
		web) echo ".air.web.toml" ;;
	esac
}

dev_service_air_pattern() {
	case "$(dev_service_name "$1")" in
		issue-tracker) echo 'air.*\.air[.]issue-tracker[.]toml' ;;
		orchestrator) echo 'air.*\.air[.]orchestrator[.]toml' ;;
		web) echo 'air.*\.air[.]web[.]toml' ;;
	esac
}

dev_service_log_file() {
	case "$(dev_service_name "$1")" in
		issue-tracker) echo "issue-tracker.log" ;;
		orchestrator) echo "orchestrator.log" ;;
		web) echo "web.log" ;;
	esac
}

dev_service_pid_file() {
	case "$(dev_service_name "$1")" in
		issue-tracker) echo "$(dev_run_dir)/issue-tracker.pid" ;;
		orchestrator) echo "$(dev_run_dir)/orchestrator.pid" ;;
		web) echo "$(dev_run_dir)/web.pid" ;;
	esac
}

dev_service_health_url() {
	case "$(dev_service_name "$1")" in
		issue-tracker) echo "http://127.0.0.1:8080/api/v1/health" ;;
		web) echo "http://127.0.0.1:3000/health" ;;
		*) return 1 ;;
	esac
}

dev_service_start_delay() {
	case "$(dev_service_name "$1")" in
		orchestrator) echo "1" ;;
		*) echo "0" ;;
	esac
}

dev_service_needs_mod_download() {
	case "$(dev_service_name "$1")" in
		issue-tracker) return 0 ;;
		*) return 1 ;;
	esac
}
