#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

: "${DS2API_ADMIN_KEY:?Set DS2API_ADMIN_KEY as a PandaStack Secret before deployment}"
: "${DS2API_CONFIG_JSON:?Set DS2API_CONFIG_JSON as a PandaStack Secret before deployment}"

if [[ ${#DS2API_ADMIN_KEY} -lt 32 ]]; then
  printf 'DS2API_ADMIN_KEY must contain at least 32 random characters. Value omitted.\n' >&2
  exit 1
fi

if [[ ! -x bin/ds2api || ! -s static/admin/index.html ]]; then
  printf 'Build artifacts are missing. Re-run the PandaStack build step.\n' >&2
  exit 1
fi

# The durable source of config is the platform Secret, not the ephemeral VM disk.
export DS2API_ENV_WRITEBACK=0
export DS2API_AUTO_BUILD_WEBUI=false
export DS2API_STATIC_ADMIN_DIR="$PWD/static/admin"
export DS2API_CHAT_HISTORY_PATH="${DS2API_CHAT_HISTORY_PATH:-/tmp/ds2api-chat-history.json}"
export PORT="${PORT:-5001}"

exec ./bin/ds2api
