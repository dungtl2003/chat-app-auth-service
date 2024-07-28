#!/usr/bin/bash

ENV_FILE="./environments/dev/.env";

setup_colors() {
    # Only use colors if connected to a terminal
    if [[ ! is_tty ]]; then
        FMT_RED="";
        FMT_GREEN="";
        FMT_YELLOW="";
        FMT_BLUE="";
        FMT_BOLD="";
        FMT_RESET="";
        return;
    fi

    FMT_RED=$(printf '\033[31m');
    FMT_GREEN=$(printf '\033[1;32m');
    FMT_YELLOW=$(printf '\033[33m');
    FMT_BLUE=$(printf '\033[1;34m');
    FMT_BOLD=$(printf '\033[1m');
    FMT_RESET=$(printf '\033[0m');
}

export_envs() {
    for line in "${lines[@]}"; do
        printf "export %s\n" $line;
        export $line;
    done
}

clean_envs() {
    for line in "${lines[@]}"; do
        pair=(${line//=/ })
        printf "unset %s\n" ${pair[0]};
        unset ${pair[0]};
    done
}

read_file() {
    IFS=$'\n' read -d '' -r -a lines < ${ENV_FILE};
}

run_cmd_with_envs() {
    read_file;
    export_envs;
    exec_cmd "$@";
    clean_envs;
}

exec_cmd() {
    local args="$@";
    printf "%sexecuting command: ./mvnw %s %s\n" $FMT_GREEN "$args" $FMT_RESET;
    ./mvnw $args;
}

main() {
    setup_colors;

    if [ -f ${ENV_FILE} ]; then
        printf "%senv file found.%s\n" $FMT_BLUE $FMT_RESET;
        run_cmd_with_envs "$@";
    else
        printf "%senv file not found.%s\n" $FMT_BLUE $FMT_RESET;
        exec_cmd "$@";
    fi
}

main "$@";
