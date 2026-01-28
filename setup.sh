#!/bin/sh
set -eu

ENV_FILE="fortyone.env"
EXAMPLE_FILE="fortyone.env.example"
COMPOSE_FILE="docker-compose.yml"
NON_INTERACTIVE=false
REPO="${FORTYONE_REPO:-complexus-tech/fortyone}"
RELEASE_TAG="${FORTYONE_RELEASE_TAG:-latest}"
ASSET_BASE_URL="${FORTYONE_ASSET_BASE_URL:-}"

for arg in "$@"; do
  case "$arg" in
    --non-interactive) NON_INTERACTIVE=true ;;
  esac
done

prompt() {
  var_name="$1"
  prompt_text="$2"
  default_value="$3"

  printf "%s" "$prompt_text"
  if [ -n "$default_value" ]; then
    printf " [%s]" "$default_value"
  fi
  printf ": "
  read -r input
  if [ -z "$input" ]; then
    input="$default_value"
  fi
  eval "$var_name=\"$input\""
}

prompt_yes_no() {
  var_name="$1"
  prompt_text="$2"
  default_value="$3"
  default_hint="y"

  if [ "$default_value" = "false" ]; then
    default_hint="n"
  fi

  printf "%s [%s]: " "$prompt_text" "$default_hint"
  read -r input
  if [ -z "$input" ]; then
    input="$default_hint"
  fi
  case "$input" in
    y|Y|yes|YES) eval "$var_name=true" ;;
    *) eval "$var_name=false" ;;
  esac
}

asset_base_url() {
  if [ -n "$ASSET_BASE_URL" ]; then
    printf "%s" "$ASSET_BASE_URL"
    return
  fi

  if [ "$RELEASE_TAG" = "latest" ]; then
    printf "https://github.com/%s/releases/latest/download" "$REPO"
    return
  fi

  printf "https://github.com/%s/releases/download/%s" "$REPO" "$RELEASE_TAG"
}

download_asset() {
  asset_name="$1"
  destination="$2"
  url="$(asset_base_url)/$asset_name"

  echo "Downloading $asset_name..."
  if ! curl -fsSL -o "$destination" "$url"; then
    echo "ERROR: Failed to download $asset_name from $url"
    echo "Check FORTYONE_REPO or FORTYONE_ASSET_BASE_URL."
    exit 1
  fi
}

ensure_assets() {
  if [ ! -f "$COMPOSE_FILE" ]; then
    download_asset "$COMPOSE_FILE" "$COMPOSE_FILE"
  fi

  if [ ! -f "$EXAMPLE_FILE" ]; then
    download_asset "$EXAMPLE_FILE" "$EXAMPLE_FILE"
  fi
}

load_env() {
  if [ ! -f "$ENV_FILE" ]; then
    echo "ERROR: $ENV_FILE not found. Run install first."
    exit 1
  fi
  set -a
  . "$ENV_FILE"
  set +a
}

profiles_args() {
  args=""
  if [ -n "${COMPOSE_PROFILES:-}" ]; then
    old_ifs="$IFS"
    IFS=','
    for profile in $COMPOSE_PROFILES; do
      args="$args --profile $profile"
    done
    IFS="$old_ifs"
  fi
  printf "%s" "$args"
}

has_profile() {
  profile="$1"
  case ",${COMPOSE_PROFILES:-}," in
    *",${profile},"*) return 0 ;;
    *) return 1 ;;
  esac
}

compose() {
  if [ ! -f "$COMPOSE_FILE" ]; then
    echo "ERROR: $COMPOSE_FILE not found. Run install to download it."
    exit 1
  fi
  args=$(profiles_args)
  docker compose --env-file "$ENV_FILE" $args "$@"
}

run_migrations() {
  load_env
  sslmode="require"
  if [ "${APP_DB_DISABLE_TLS:-true}" = "true" ]; then
    sslmode="disable"
  fi

  db_url="postgresql://${APP_DB_USER}:${APP_DB_PASSWORD}@${APP_DB_HOST}:${APP_DB_PORT}/${APP_DB_NAME}?sslmode=${sslmode}"
  migrations_dir="$PWD/apps/server/internal/migrations"

  if [ ! -d "$migrations_dir" ]; then
    echo "WARNING: migrations directory not found: $migrations_dir"
    echo "Skipping migrations. Provide migrations or run them separately."
    return
  fi

  if [ "${APP_DB_HOST}" = "postgres" ]; then
    if ! has_profile "postgres"; then
      echo "ERROR: APP_DB_HOST is 'postgres' but the postgres profile is not enabled."
      echo "Add 'postgres' to COMPOSE_PROFILES in $ENV_FILE or use an external DB host."
      exit 1
    fi
    compose up -d postgres
    echo "Waiting for postgres..."
    attempts=30
    while [ $attempts -gt 0 ]; do
      if docker compose --env-file "$ENV_FILE" --profile postgres exec -T postgres pg_isready -U "$APP_DB_USER" -d "$APP_DB_NAME" >/dev/null 2>&1; then
        break
      fi
      attempts=$((attempts - 1))
      sleep 2
    done
    if [ $attempts -eq 0 ]; then
      echo "ERROR: Postgres is not ready."
      exit 1
    fi
  fi

  echo "Running migrations..."
  docker compose --env-file "$ENV_FILE" --profile migrations run --rm --no-deps migrations \
    -path /migrations \
    -database "$db_url" \
    up
}

install_env() {
  ensure_assets
  if [ "$NON_INTERACTIVE" = "true" ]; then
    if [ ! -f "$ENV_FILE" ]; then
      if [ -f "$EXAMPLE_FILE" ]; then
        cp "$EXAMPLE_FILE" "$ENV_FILE"
        echo "Created $ENV_FILE from $EXAMPLE_FILE"
      else
        echo "ERROR: $EXAMPLE_FILE not found"
        exit 1
      fi
    else
      echo "$ENV_FILE already exists, leaving it unchanged"
    fi
    return
  fi

  echo "FortyOne install"
  echo "This will generate $ENV_FILE. Press Enter to accept defaults."
  echo ""

  prompt "WEB_PORT" "Web host port" "3000"
  default_website_url="http://localhost:${WEB_PORT}"
  prompt "APP_WEBSITE_URL" "Website URL (public domain for links/emails)" "$default_website_url"
  SERVER_PORT="8000"
  WORKER_PORT="8080"
  INTERNAL_API_URL="http://server:8000"
  NEXT_PUBLIC_API_URL="http://localhost:8000"
  APP_API_HOST="0.0.0.0:8000"

  prompt_yes_no "USE_POSTGRES" "Use local Postgres container" "true"
  if [ "$USE_POSTGRES" = "true" ]; then
    APP_DB_HOST="postgres"
    APP_DB_PORT="5432"
    APP_DB_USER="postgres"
    APP_DB_PASSWORD="postgres"
    APP_DB_NAME="fortyone"
    APP_DB_DISABLE_TLS="true"
    POSTGRES_DB="fortyone"
    POSTGRES_USER="postgres"
    POSTGRES_PASSWORD="postgres"
    POSTGRES_HOST_PORT="5432"
  else
    prompt "APP_DB_HOST" "External DB host" ""
    prompt "APP_DB_PORT" "External DB port" "5432"
    prompt "APP_DB_USER" "External DB user" ""
    prompt "APP_DB_PASSWORD" "External DB password" ""
    prompt "APP_DB_NAME" "External DB name" "fortyone"
    prompt "APP_DB_DISABLE_TLS" "Disable DB TLS (true/false)" "false"
    POSTGRES_DB="fortyone"
    POSTGRES_USER="postgres"
    POSTGRES_PASSWORD="postgres"
    POSTGRES_HOST_PORT="5432"
  fi

  prompt_yes_no "USE_REDIS" "Use local Redis container" "true"
  if [ "$USE_REDIS" = "true" ]; then
    APP_REDIS_HOST="redis"
    APP_REDIS_PORT="6379"
    APP_REDIS_PASSWORD=""
    APP_REDIS_DISABLE_TLS="true"
    REDIS_HOST_PORT="6380"
  else
    prompt "APP_REDIS_HOST" "External Redis host" ""
    prompt "APP_REDIS_PORT" "External Redis port" "6379"
    prompt "APP_REDIS_PASSWORD" "External Redis password" ""
    prompt "APP_REDIS_DISABLE_TLS" "Disable Redis TLS (true/false)" "false"
    REDIS_HOST_PORT="6380"
  fi

  APP_TRACING_HOST="jaeger:4318"

  prompt_yes_no "USE_MAILPIT" "Use local Mailpit (SMTP)" "true"
  if [ "$USE_MAILPIT" = "true" ]; then
    APP_EMAIL_HOST="mailpit"
    APP_EMAIL_PORT="1025"
    APP_EMAIL_FROM_ADDRESS="notifications@fortyone.local"
    APP_EMAIL_FROM_NAME="FortyOne"
    APP_EMAIL_ENVIRONMENT="development"
    MAILPIT_WEB_PORT="8025"
    MAILPIT_SMTP_PORT="1025"
  else
    prompt "APP_EMAIL_HOST" "SMTP host" ""
    prompt "APP_EMAIL_PORT" "SMTP port" "587"
    prompt "APP_EMAIL_FROM_ADDRESS" "From address" ""
    prompt "APP_EMAIL_FROM_NAME" "From name" "FortyOne"
    prompt "APP_EMAIL_ENVIRONMENT" "Email environment" "production"
    MAILPIT_WEB_PORT="8025"
    MAILPIT_SMTP_PORT="1025"
  fi

  prompt "APP_STORAGE_PROVIDER" "Storage provider (aws/azure)" "aws"
  STORAGE_PROFILE_IMAGES_NAME="profile-images"
  STORAGE_WORKSPACE_LOGOS_NAME="workspace-logos"
  STORAGE_ATTACHMENTS_NAME="attachments"
  USE_JAEGER=true
  JAEGER_UI_PORT="16686"
  JAEGER_GRPC_PORT="4317"
  JAEGER_HTTP_PORT="4318"
  if [ "$APP_STORAGE_PROVIDER" = "azure" ]; then
    APP_AWS_ACCESS_KEY_ID=""
    APP_AWS_SECRET_ACCESS_KEY=""
    APP_AWS_REGION=""
    APP_AWS_ENDPOINT=""
    APP_AWS_PUBLIC_URL=""
    APP_AWS_FORCE_PATH_STYLE=""
    MINIO_ROOT_USER="minioadmin"
    MINIO_ROOT_PASSWORD="minioadmin"
    MINIO_PORT="9000"
    MINIO_CONSOLE_PORT="9001"
    MINIO_ENDPOINT="http://minio:9000"

    prompt "APP_AZURE_STORAGE_ACCOUNT_KEY" "Azure storage account key" ""
    prompt "APP_AZURE_STORAGE_CONNECTION_STRING" "Azure connection string" ""
    prompt "APP_AZURE_STORAGE_ACCOUNT_NAME" "Azure account name" ""
    USE_MINIO=false
  else
    prompt_yes_no "USE_MINIO" "Use local MinIO" "true"
    if [ "$USE_MINIO" = "true" ]; then
      APP_AWS_ACCESS_KEY_ID="minioadmin"
      APP_AWS_SECRET_ACCESS_KEY="minioadmin"
      APP_AWS_REGION="us-east-1"
      APP_AWS_ENDPOINT="http://minio:9000"
      APP_AWS_PUBLIC_URL="http://localhost:9000"
      APP_AWS_FORCE_PATH_STYLE="true"
      MINIO_ROOT_USER="minioadmin"
      MINIO_ROOT_PASSWORD="minioadmin"
      MINIO_PORT="9000"
      MINIO_CONSOLE_PORT="9001"
      MINIO_ENDPOINT="http://minio:9000"
    else
      prompt "APP_AWS_ACCESS_KEY_ID" "AWS access key" ""
      prompt "APP_AWS_SECRET_ACCESS_KEY" "AWS secret key" ""
      prompt "APP_AWS_REGION" "AWS region" ""
      prompt "APP_AWS_ENDPOINT" "AWS endpoint" ""
      prompt "APP_AWS_PUBLIC_URL" "AWS public URL" ""
      prompt "APP_AWS_FORCE_PATH_STYLE" "Force path style (true/false)" "false"
      MINIO_ROOT_USER="minioadmin"
      MINIO_ROOT_PASSWORD="minioadmin"
      MINIO_PORT="9000"
      MINIO_CONSOLE_PORT="9001"
      MINIO_ENDPOINT="http://minio:9000"
    fi

    APP_AZURE_STORAGE_ACCOUNT_KEY=""
    APP_AZURE_STORAGE_CONNECTION_STRING=""
    APP_AZURE_STORAGE_ACCOUNT_NAME=""
  fi

  prompt "AUTH_SECRET" "Auth secret (leave blank to generate)" ""
  if [ -z "$AUTH_SECRET" ]; then
    AUTH_SECRET=$(openssl rand -base64 32 2>/dev/null || true)
    if [ -z "$AUTH_SECRET" ]; then
      AUTH_SECRET="changeme"
      echo "WARNING: Could not generate AUTH_SECRET. Using 'changeme'."
    fi
  fi

  cat > "$ENV_FILE" <<EOF
NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL
INTERNAL_API_URL=$INTERNAL_API_URL
AUTH_SECRET=$AUTH_SECRET
WEB_PORT=$WEB_PORT

SERVER_PORT=$SERVER_PORT
WORKER_PORT=$WORKER_PORT
APP_API_HOST=$APP_API_HOST
APP_WEBSITE_URL=$APP_WEBSITE_URL
APP_TRACING_HOST=$APP_TRACING_HOST

APP_DB_HOST=$APP_DB_HOST
APP_DB_PORT=$APP_DB_PORT
APP_DB_USER=$APP_DB_USER
APP_DB_PASSWORD=$APP_DB_PASSWORD
APP_DB_NAME=$APP_DB_NAME
APP_DB_DISABLE_TLS=$APP_DB_DISABLE_TLS
POSTGRES_DB=$POSTGRES_DB
POSTGRES_USER=$POSTGRES_USER
POSTGRES_PASSWORD=$POSTGRES_PASSWORD
POSTGRES_HOST_PORT=$POSTGRES_HOST_PORT

APP_REDIS_HOST=$APP_REDIS_HOST
APP_REDIS_PORT=$APP_REDIS_PORT
APP_REDIS_PASSWORD=$APP_REDIS_PASSWORD
APP_REDIS_DISABLE_TLS=$APP_REDIS_DISABLE_TLS
REDIS_HOST_PORT=$REDIS_HOST_PORT

APP_EMAIL_HOST=$APP_EMAIL_HOST
APP_EMAIL_PORT=$APP_EMAIL_PORT
APP_EMAIL_FROM_ADDRESS=$APP_EMAIL_FROM_ADDRESS
APP_EMAIL_FROM_NAME=$APP_EMAIL_FROM_NAME
APP_EMAIL_ENVIRONMENT=$APP_EMAIL_ENVIRONMENT
MAILPIT_WEB_PORT=$MAILPIT_WEB_PORT
MAILPIT_SMTP_PORT=$MAILPIT_SMTP_PORT

JAEGER_UI_PORT=$JAEGER_UI_PORT
JAEGER_GRPC_PORT=$JAEGER_GRPC_PORT
JAEGER_HTTP_PORT=$JAEGER_HTTP_PORT

APP_STORAGE_PROVIDER=$APP_STORAGE_PROVIDER

APP_AWS_ACCESS_KEY_ID=$APP_AWS_ACCESS_KEY_ID
APP_AWS_SECRET_ACCESS_KEY=$APP_AWS_SECRET_ACCESS_KEY
APP_AWS_REGION=$APP_AWS_REGION
APP_AWS_ENDPOINT=$APP_AWS_ENDPOINT
APP_AWS_PUBLIC_URL=$APP_AWS_PUBLIC_URL
APP_AWS_FORCE_PATH_STYLE=$APP_AWS_FORCE_PATH_STYLE
MINIO_ROOT_USER=$MINIO_ROOT_USER
MINIO_ROOT_PASSWORD=$MINIO_ROOT_PASSWORD
MINIO_PORT=$MINIO_PORT
MINIO_CONSOLE_PORT=$MINIO_CONSOLE_PORT
MINIO_ENDPOINT=$MINIO_ENDPOINT

APP_AZURE_STORAGE_ACCOUNT_KEY=$APP_AZURE_STORAGE_ACCOUNT_KEY
APP_AZURE_STORAGE_CONNECTION_STRING=$APP_AZURE_STORAGE_CONNECTION_STRING
APP_AZURE_STORAGE_ACCOUNT_NAME=$APP_AZURE_STORAGE_ACCOUNT_NAME
EOF

  profiles=""
  append_profile() {
    if [ -z "$profiles" ]; then
      profiles="$1"
    else
      profiles="$profiles,$1"
    fi
  }

  if [ "$USE_POSTGRES" = "true" ]; then
    append_profile "postgres"
  fi
  if [ "$USE_REDIS" = "true" ]; then
    append_profile "redis"
  fi
  if [ "$USE_MAILPIT" = "true" ]; then
    append_profile "mailpit"
  fi
  if [ "$USE_JAEGER" = "true" ]; then
    append_profile "jaeger"
  fi
  if [ "${USE_MINIO:-false}" = "true" ]; then
    append_profile "minio"
  fi

  if [ -n "$profiles" ]; then
    printf "\nCOMPOSE_PROFILES=%s\n" "$profiles" >> "$ENV_FILE"
  fi
}

action="${1:-}"

if [ -z "$action" ]; then
  echo "Select action:"
  echo "  1) Install"
  echo "  2) Start"
  echo "  3) Stop"
  echo "  4) Upgrade"
  printf "Action: "
  read -r choice
  case "$choice" in
    1) action="install" ;;
    2) action="start" ;;
    3) action="stop" ;;
    4) action="upgrade" ;;
    *) exit 0 ;;
  esac
fi

case "$action" in
  install)
    install_env
    load_env
    compose up -d
    run_migrations
    ;;
  start)
    load_env
    compose up -d
    ;;
  stop)
    load_env
    compose down
    ;;
  upgrade)
    load_env
    compose pull
    compose up -d
    ;;
  *)
    echo "Usage: ./setup.sh [install|start|stop|upgrade] [--non-interactive]"
    exit 1
    ;;
esac
