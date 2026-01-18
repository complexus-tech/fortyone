#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="${FORTYONE_HOME:-$(pwd)/fortyone-selfhost}"
DOCKER_DIR="$ROOT_DIR/docker"
ENV_FILE="$DOCKER_DIR/fortyone.env"
COMPOSE_FILE="$DOCKER_DIR/compose.yml"
COMPOSE_URL="https://fortyone.app/docker/compose.yml"
ENV_URL="https://fortyone.app/docker/fortyone.env"

print_header() {
  printf "\nFortyOne Docker Installer\n\n"
  printf "Install directory: %s\n" "$ROOT_DIR"
}

ensure_dir() {
  mkdir -p "$DOCKER_DIR"
}

require_docker() {
  if ! command -v docker >/dev/null 2>&1; then
    printf "Docker is required but not installed.\n"
    exit 1
  fi
  if ! docker info >/dev/null 2>&1; then
    printf "Docker is installed but not running. Start Docker and retry.\n"
    exit 1
  fi
}

fetch_artifacts() {
  if [ ! -f "$COMPOSE_FILE" ]; then
    printf "Downloading compose file...\n"
    curl -fsSL "$COMPOSE_URL" -o "$COMPOSE_FILE"
  fi

  if [ ! -f "$ENV_FILE" ]; then
    printf "Downloading environment template...\n"
    curl -fsSL "$ENV_URL" -o "$ENV_FILE"
  fi
}

prompt_value() {
  local label="$1"
  local key="$2"
  local default="$3"
  local value
  read -r -p "$label [$default]: " value
  value=${value:-$default}
  if grep -q "^${key}=" "$ENV_FILE"; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
      sed -i '' "s#^${key}=.*#${key}=${value}#" "$ENV_FILE"
    else
      sed -i "s#^${key}=.*#${key}=${value}#" "$ENV_FILE"
    fi
  else
    printf "\n%s=%s\n" "$key" "$value" >> "$ENV_FILE"
  fi
}

warn_placeholder() {
  local key="$1"
  local default="$2"
  if grep -q "^${key}=${default}$" "$ENV_FILE"; then
    printf "Warning: %s is still set to the default placeholder.\n" "$key"
  fi
}

configure_env() {
  printf "\nConfigure runtime values (press Enter to accept defaults).\n"
  prompt_value "Web URL" "AUTH_URL" "http://localhost:3000"
  prompt_value "API URL" "NEXT_PUBLIC_API_URL" "http://localhost:8000"
  prompt_value "Website URL" "APP_WEBSITE_URL" "http://localhost:3000"
  prompt_value "API secret (APP_AUTH_SECRET_KEY)" "APP_AUTH_SECRET_KEY" "change-me"
  prompt_value "Projects auth secret (AUTH_SECRET)" "AUTH_SECRET" "change-me"
  warn_placeholder "APP_AUTH_SECRET_KEY" "change-me"
  warn_placeholder "AUTH_SECRET" "change-me"
}

compose_cmd() {
  docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" "$@"
}

compose_cmd_with_proxy() {
  docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" -f "$DOCKER_DIR/compose.proxy.yml" "$@"
}

ensure_env() {
  if [ ! -f "$ENV_FILE" ]; then
    printf "Environment file not found. Run Install first.\n"
    exit 1
  fi
  if [ ! -f "$COMPOSE_FILE" ]; then
    printf "Compose file not found. Run Install first.\n"
    exit 1
  fi
}

install_stack() {
  require_docker
  ensure_dir
  fetch_artifacts
  configure_env
  compose_cmd up -d
  printf "\nFortyOne is starting.\n"
}

install_proxy() {
  require_docker
  ensure_dir
  fetch_artifacts
  configure_env
  if [ ! -f "$DOCKER_DIR/compose.proxy.yml" ]; then
    printf "Downloading proxy compose file...\n"
    curl -fsSL "https://fortyone.app/docker/compose.proxy.yml" -o "$DOCKER_DIR/compose.proxy.yml"
  fi
  if [ ! -f "$DOCKER_DIR/nginx.conf" ]; then
    printf "Downloading proxy config...\n"
    curl -fsSL "https://fortyone.app/docker/nginx.conf" -o "$DOCKER_DIR/nginx.conf"
  fi
  docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" -f "$DOCKER_DIR/compose.proxy.yml" up -d
  printf "\nFortyOne is starting with the proxy enabled.\n"
}

start_stack() {
  require_docker
  ensure_env
  if [ -f "$DOCKER_DIR/compose.proxy.yml" ]; then
    compose_cmd_with_proxy up -d
  else
    compose_cmd up -d
  fi
}

stop_stack() {
  require_docker
  ensure_env
  if [ -f "$DOCKER_DIR/compose.proxy.yml" ]; then
    compose_cmd_with_proxy stop
  else
    compose_cmd stop
  fi
}

restart_stack() {
  require_docker
  ensure_env
  if [ -f "$DOCKER_DIR/compose.proxy.yml" ]; then
    compose_cmd_with_proxy restart
  else
    compose_cmd restart
  fi
}

view_logs() {
  require_docker
  ensure_env
  if [ -f "$DOCKER_DIR/compose.proxy.yml" ]; then
    compose_cmd_with_proxy logs -f
  else
    compose_cmd logs -f
  fi
}

down_stack() {
  require_docker
  ensure_env
  if [ -f "$DOCKER_DIR/compose.proxy.yml" ]; then
    compose_cmd_with_proxy down
  else
    compose_cmd down
  fi
}

dev_build() {
  require_docker
  ensure_dir
  if [ ! -f "$ENV_FILE" ]; then
    if [ ! -f "$(pwd)/docker/fortyone.env" ]; then
      printf "Local docker/fortyone.env not found.\n"
      exit 1
    fi
    printf "Using local template at docker/fortyone.env\n"
    cp "$(pwd)/docker/fortyone.env" "$ENV_FILE"
    configure_env
  fi
  if [ ! -f "$COMPOSE_FILE" ]; then
    if [ ! -f "$(pwd)/docker/compose.yml" ]; then
      printf "Local docker/compose.yml not found.\n"
      exit 1
    fi
    cp "$(pwd)/docker/compose.yml" "$COMPOSE_FILE"
  fi
  docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" -f "$(pwd)/docker/compose.dev.yml" up -d --build
}

show_menu() {
  printf "\nSelect an action:\n"
  printf "1) Install (no proxy)\n"
  printf "2) Install (with proxy)\n"
  printf "3) Start\n"
  printf "4) Stop\n"
  printf "5) Restart\n"
  printf "6) View Logs\n"
  printf "7) Down (remove containers)\n"
  printf "8) Dev Build (local source)\n"
  printf "9) Exit\n"
  read -r -p "Action [1]: " choice
  choice=${choice:-1}
  case "$choice" in
    1) install_stack ;;
    2) install_proxy ;;
    3) start_stack ;;
    4) stop_stack ;;
    5) restart_stack ;;
    6) view_logs ;;
    7) down_stack ;;
    8) dev_build ;;
    9) exit 0 ;;
    *) printf "Invalid option.\n" ;;
  esac
}

print_header
show_menu
