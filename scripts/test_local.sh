#!/bin/bash -e

# This script is used to run any command with environment variables specific to
# the test environment. By default, it will also run all necessary services in
# docker-compose.test.yaml file. You can change the compose file by setting the
# COMPOSE_FILE environment variable. You can also set the DEBUG environment variable
# to "true" to enable debug mode.

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
ROOT_DIR="$SCRIPT_DIR/.."
TEST_DIR=${TEST_DIR:-"$ROOT_DIR/tests"}
DEBUG=${DEBUG:-"false"}
COMPOSE_FILE=${COMPOSE_FILE:-"docker-compose.test.yaml"}

PORT=${PORT:-8400}
API_VERSION=${API_VERSION:-"v1"}
LOG_LEVEL=${LOG_LEVEL:-"DEBUG"}
LOG_KIND=${LOG_KIND:-"TEXT"}
ENV=${ENV:-"test"}
USER_SERVICE_URL=${USER_SERVICE_URL:-"http://localhost:8600"}
JWT_SECRET=${JWT_SECRET:-"secret"}
# set low for testing expiration, but don't set too low
ACCESS_TOKEN_DURATION_MS=${ACCESS_TOKEN_DURATION_MS:-4000}
REFRESH_TOKEN_DURATION_MS=${REFRESH_TOKEN_DURATION_MS:-6000}
DOMAIN_NAME=${DOMAIN_NAME:-"localhost"}
COST=${COST:-12}
ORIGIN=${ORIGIN:-"http://localhost:5174"}

# Test's specific environment variables
ADMIN_DATABASE_URL=${ADMIN_DATABASE_URL:-"postgresql://admin:testpass123@localhost:6000/chat-app?sslmode=disable"}
AUTH_URL=${AUTH_URL:-"http://localhost:8400/auth"}

command="$1"
extraArgs="${@:2}"

if [ "$DEBUG" == "true" ]; then
    echo "Debug mode is on"
    set -x
fi

function export_envs() {
    printf "export COMPOSE_FILE=%s\n" $COMPOSE_FILE
    export COMPOSE_FILE

    printf "export PORT=%s\n" $PORT
    export PORT
    printf "export API_VERSION=%s\n" $API_VERSION
    export API_VERSION
    printf "export LOG_LEVEL=%s\n" $LOG_LEVEL
    export LOG_LEVEL
    printf "export LOG_KIND=%s\n" $LOG_KIND
    export LOG_KIND
    printf "export ENV=%s\n" $ENV
    export ENV
    printf "export USER_SERVICE_URL=%s\n" $USER_SERVICE_URL
    export USER_SERVICE_URL
    printf "export JWT_SECRET=%s\n" $JWT_SECRET
    export JWT_SECRET
    printf "export ACCESS_TOKEN_DURATION_MS=%s\n" $ACCESS_TOKEN_DURATION_MS
    export ACCESS_TOKEN_DURATION_MS
    printf "export REFRESH_TOKEN_DURATION_MS=%s\n" $REFRESH_TOKEN_DURATION_MS
    export REFRESH_TOKEN_DURATION_MS
    printf "export DOMAIN_NAME=%s\n" $DOMAIN_NAME
    export DOMAIN_NAME
    printf "export COST=%s\n" $COST
    export COST
    printf "export ORIGIN=%s\n" $ORIGIN
    export ORIGIN

    printf "export ADMIN_DATABASE_URL=%s\n" $ADMIN_DATABASE_URL
    export ADMIN_DATABASE_URL
    printf "export AUTH_URL=%s\n" $AUTH_URL
    export AUTH_URL
}

function main() {
    export_envs
    NO_ENV_FILE=1 $SCRIPT_DIR/run_with_services.sh "$command" $extraArgs
}

main
