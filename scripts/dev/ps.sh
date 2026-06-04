#!/bin/sh
set -eu

ps -ef | grep -E "air|issue-tracker|orchestrator|cmd/web|air-web" | grep -v grep || true
