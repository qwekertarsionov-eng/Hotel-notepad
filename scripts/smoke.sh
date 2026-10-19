#!/usr/bin/env bash
set -uo pipefail

# TST-INFRA-001: docker compose up разворачивает стек; эндпоинт отвечает
if ! command -v docker >/dev/null 2>&1; then
	echo "skip: docker is not available"
	exit 0
fi
if ! docker compose version >/dev/null 2>&1; then
	echo "skip: docker compose is not available"
	exit 0
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="$ROOT_DIR/docker-compose.yml"
if [ ! -f "$COMPOSE_FILE" ]; then
	echo "skip: docker-compose.yml not found"
	exit 0
fi

APP_PORT="${APP_PORT:-8080}"

docker compose -f "$COMPOSE_FILE" up -d

up=0
for i in $(seq 1 30); do
	if curl -s -o /dev/null "http://127.0.0.1:${APP_PORT}/" 2>/dev/null; then
		up=1
		break
	fi
	sleep 1
done

if [ "$up" -eq 0 ]; then
	echo "fail: stack endpoint did not respond on port ${APP_PORT}"
	docker compose -f "$COMPOSE_FILE" logs --tail=50 2>/dev/null || true
	docker compose -f "$COMPOSE_FILE" down 2>/dev/null || true
	exit 1
fi

echo "ok: stack is up and responding"
docker compose -f "$COMPOSE_FILE" down 2>/dev/null || true
