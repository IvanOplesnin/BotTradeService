#!/usr/bin/env bash
set -euo pipefail

set -a
. "$ENV_FILE"
set +a


go run cmd/grpc_server/main.go