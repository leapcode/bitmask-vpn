#!/usr/bin/env bash
set -euo pipefail

EXPECTED_QT_VERSION='6.6'  # major.minor only
FOUND_QT_VERSION=
QT_VERSION_OUTPUT=$(qmake6 --version)

# split input on newline into lines array
mapfile -t lines <<< "$QT_VERSION_OUTPUT"

for i in "${!lines[@]}"; do 
  line="${lines[i]}"
  # extract version from line
  QT_VER=$(awk -F'version[[:space:]]+' '{
    if ($2 ~ /^[0-9]+\.[0-9]+(\.[0-9]+)?/) {
      match($2, /^[0-9]+\.[0-9]+(\.[0-9]+)?/, m)
      print m[0]
    }
  }' <<<"$line")

  # check if QT_VER starts with EXPECTED_QT_VERSION
  if [[ -n "$QT_VER" && "$QT_VER" == "$EXPECTED_QT_VERSION"* ]]; then
    FOUND_QT_VERSION=$QT_VER
    break
  fi
done

if [[ -n "$FOUND_QT_VERSION" ]]; then
  echo "QT version match: $FOUND_QT_VERSION matches version requirement $EXPECTED_QT_VERSION" 
  exit 0
fi

echo "Expected QT version $EXPECTED_QT_VERSION is not in path, please update. Installed: ${QT_VER:-<none>}" >&2
exit 1