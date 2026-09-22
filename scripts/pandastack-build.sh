#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

# PandaStack injects secrets at build time too. Do not pass these application
# credentials to npm/go subprocesses; this is not a full build isolation boundary.
unset DS2API_CONFIG_JSON CONFIG_JSON DS2API_ADMIN_KEY DS2API_JWT_SECRET

printf '\n[1/4] Checking toolchains\n'
node --version
npm --version
go version

printf '\n[2/4] Installing locked WebUI dependencies, including build tools\n'
npm ci --include=dev --prefix webui

printf '\n[3/4] Building the admin UI\n'
npm run build --prefix webui
test -s static/admin/index.html

printf '\n[4/4] Building the Go service\n'
mkdir -p bin
BUILD_VERSION="$(tr -d '[:space:]' < VERSION)"

CGO_ENABLED=0 go build -trimpath -buildvcs=false \
  -ldflags="-s -w -X ds2api/internal/version.BuildVersion=${BUILD_VERSION}" \
  -o bin/ds2api ./cmd/ds2api

test -x bin/ds2api
printf '\nBUILD_OK: bin/ds2api and static/admin/index.html are ready.\n'
