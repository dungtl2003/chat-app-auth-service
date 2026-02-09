#!/bin/bash

# Exit immediately if a command exits with a non-zero status
# Treat unset variables as an error
# Return the exit status of the last command in the pipe that failed
set -euo pipefail

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
ROOT_DIR="${SCRIPT_DIR}/.."

# --- Environment Configuration ---

# Mark all subsequently defined variables for export automatically
set -a

# Infrastructure & Paths
DEBUG="${DEBUG:-false}"
TEST_DIR="${TEST_DIR:-"$ROOT_DIR/tests"}"
COMPOSE_FILE="${COMPOSE_FILE:-"docker-compose.test.yaml"}"
ENVIRONMENT="${ENVIRONMENT:-test}"
AUTH_URL=${AUTH_URL:-"http://localhost:8400"}

# Service Configuration
PORT="${PORT:-8400}"
LOG_LEVEL="${LOG_LEVEL:-DEBUG}"
LOG_KIND="${LOG_KIND:-TEXT}"
JWT_SECRET=${JWT_SECRET:-"secret"}
ACCESS_TOKEN_DURATION_MS=${ACCESS_TOKEN_DURATION_MS:-4000}
REFRESH_TOKEN_DURATION_MS=${REFRESH_TOKEN_DURATION_MS:-6000}
DOMAIN_NAME=${DOMAIN_NAME:-"localhost"}
PASSWORD_HASH_COST=${PASSWORD_HASH_COST:-"10"}
ID_GENERATOR_SERVICE_ADDR=${ID_GENERATOR_SERVICE_ADDR:-"localhost:9000"}
ID_GENERATOR_SERVICE_EPOCH=${ID_GENERATOR_SERVICE_EPOCH:-1654041600000}
ID_GENERATOR_SERVICE_CERT_DIR=${ID_GENERATOR_SERVICE_CERT_DIR:-"$ROOT_DIR/environments/test/snowflake/ssl/certs"}
SMTP_HOST="${SMTP_HOST:-smtp.example.com}"
SMTP_PORT="${SMTP_PORT:-587}"
SMTP_EMAIL_FROM_NAME_DISPLAY="${SMTP_EMAIL_FROM_NAME_DISPLAY:-Chat App}"
PASSWORD_RESET_RATE_LIMIT_TTL="${PASSWORD_RESET_RATE_LIMIT_TTL:-1h}"
PASSWORD_RESET_RATE_LIMIT_MAX="${PASSWORD_RESET_RATE_LIMIT_MAX:-5}"
PASSWORD_RESET_CODE_TTL="${PASSWORD_RESET_CODE_TTL:-15m}"
REDIS_ADDRESSES="${REDIS_ADDRESSES:-localhost:6379}"
REDIS_PASSWORD="${REDIS_PASSWORD:-}"

# External Services
USER_SERVICE_URL="${USER_SERVICE_URL:-http://localhost:8600}"

# Test-Specific Variables
ADMIN_DATABASE_URL="${ADMIN_DATABASE_URL:-postgresql://admin:testpass123@localhost:6000/chat-app?sslmode=disable}"
ID_GENERATOR_TLS_ADDR="${ID_GENERATOR_TLS_ADDR:-localhost:9000}"
ID_GENERATOR_NON_TLS_ADDR="${ID_GENERATOR_NON_TLS_ADDR:-localhost:9001}"
ID_GENERATOR_FAKE_CERT_DIR="${ID_GENERATOR_FAKE_CERT_DIR:-$ROOT_DIR/environments/test/conversation/services/snowflake/fake_ssl}"
DATA_FILE_DIR="${DATA_FILE_DIR:-$ROOT_DIR/tests/data}"

# Meta Configuration (Logs & Topics)
# Note: Keeping the newlines helps readability, just ensure the consuming script handles them.
LOG_META="
$ROOT_DIR/tests/logs/user_service.log=chat-app-user-service;
$ROOT_DIR/tests/logs/database_service.log=chat-app-db-service;
$ROOT_DIR/tests/logs/snowflake_tls_service.log=chat-app-snowflake-tls-service;
$ROOT_DIR/tests/logs/snowflake_non_tls_service.log=chat-app-snowflake-non-tls-service;
$ROOT_DIR/tests/logs/meili.log=chat-app-meilisearch-service;
$ROOT_DIR/tests/logs/controller_1.log=chat-app-kafka-controller-1;
$ROOT_DIR/tests/logs/controller_2.log=chat-app-kafka-controller-2;
$ROOT_DIR/tests/logs/controller_3.log=chat-app-kafka-controller-3;
$ROOT_DIR/tests/logs/broker_1.log=chat-app-kafka-broker-1;
$ROOT_DIR/tests/logs/broker_2.log=chat-app-kafka-broker-2;
$ROOT_DIR/tests/logs/broker_3.log=chat-app-kafka-broker-3;
$ROOT_DIR/tests/logs/init.log=chat-app-init-service
"

# This is used to recreate topics each test
TOPICS="
asset-resource-delete:2:2;
asset-resource-delete-dlq:1:1;
participant-resource-update:2:2;
participant-resource-update-dlq:1:1;
asset-resource-confirm:2:2;
message-resource-created:3:3;
message-resource-created-dlq:2:2;
conversation-event-resource-created:3:3;
event-resource-created:3:3;
user-resource-updated:3:3
"

# Turn off auto-export
set +a

# --- Execution ---

if [ "$DEBUG" == "true" ]; then
    echo "Debug mode is on"
    set -x
    # Optional: Print env vars to verify they are set correctly
    env | grep -E "DATABASE_|PORT|KAFKA|SERVICE_URL"
fi

COMMAND="$1"
shift # Remove the first argument (command)
EXTRA_ARGS=("$@") # Capture the rest as an array to preserve spaces/structure

# Execute the service runner
"$SCRIPT_DIR/__test_with_services.sh" "$COMMAND" "${EXTRA_ARGS[@]}"
