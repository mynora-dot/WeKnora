#!/bin/sh
# Emit only public filing values. jq safely encodes arbitrary env strings as JSON.
set -eu
if [ -n "${PUBLIC_SECURITY_BEIAN_NUMBER:-}${PUBLIC_SECURITY_BEIAN_URL:-}${PUBLIC_SECURITY_BEIAN_ICON_URL:-}" ]; then
  if [ -z "${PUBLIC_SECURITY_BEIAN_NUMBER:-}" ] || [ -z "${PUBLIC_SECURITY_BEIAN_URL:-}" ] || [ -z "${PUBLIC_SECURITY_BEIAN_ICON_URL:-}" ]; then
    echo 'Warning: incomplete public security filing configuration; footer entry will be hidden.' >&2
  fi
fi
printf 'Object.assign(window.__RUNTIME_CONFIG__, '
jq -cn --arg icp "${ICP_BEIAN_NUMBER:-}" \
  --arg number "${PUBLIC_SECURITY_BEIAN_NUMBER:-}" \
  --arg url "${PUBLIC_SECURITY_BEIAN_URL:-}" \
  --arg icon "${PUBLIC_SECURITY_BEIAN_ICON_URL:-}" \
  '{ICP_BEIAN_NUMBER:$icp,PUBLIC_SECURITY_BEIAN_NUMBER:$number,PUBLIC_SECURITY_BEIAN_URL:$url,PUBLIC_SECURITY_BEIAN_ICON_URL:$icon}'
printf ');\n'
