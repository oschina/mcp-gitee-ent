#!/bin/sh

set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
NPM_VERSION=${NPM_VERSION:-}
NPM_TOKEN=${NPM_TOKEN:-}
NPM_DRY_RUN=${NPM_DRY_RUN:-0}

if [ -z "$NPM_VERSION" ] || [ "$NPM_VERSION" = "0.0.0" ]; then
  echo "NPM_VERSION must be set to a release version" >&2
  exit 2
fi
printf '%s\n' "$NPM_VERSION" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$' || {
  echo "NPM_VERSION must be valid SemVer without build metadata: $NPM_VERSION" >&2
  exit 2
}
if [ -z "$NPM_TOKEN" ]; then
  echo "NPM_TOKEN is required" >&2
  exit 2
fi
if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required" >&2
  exit 2
fi
case "$NPM_DRY_RUN" in
  0|1) ;;
  *)
    echo "NPM_DRY_RUN must be 0 or 1" >&2
    exit 2
    ;;
esac

if [ -n "${NPM_TAG:-}" ]; then
  DIST_TAG=$NPM_TAG
else
  case "$NPM_VERSION" in
    *-*) DIST_TAG=next ;;
    *) DIST_TAG=latest ;;
  esac
fi
case "$DIST_TAG" in
  *[!A-Za-z0-9._-]*|'')
    echo "NPM_TAG contains unsupported characters: $DIST_TAG" >&2
    exit 2
    ;;
esac

staging_dir=$(mktemp -d "${TMPDIR:-/tmp}/mcp-gitee-ent-npm.XXXXXX")
trap 'rm -rf "$staging_dir"' EXIT HUP INT TERM
npmrc=$staging_dir/npmrc
printf '%s\n' "//registry.npmjs.org/:_authToken=$NPM_TOKEN" >"$npmrc"
chmod 600 "$npmrc"

staged_packages=""

stage_package() {
  package_dir=$1
  is_main=${2:-false}
  package_name=$(basename "$package_dir")
  stage=$staging_dir/$package_name

  if [ ! -f "$package_dir/package.json" ] || [ ! -d "$package_dir/bin" ]; then
    echo "npm package is incomplete: $package_dir" >&2
    exit 2
  fi

  mkdir -p "$stage"
  cp "$package_dir/package.json" "$stage/package.json"
  cp -R "$package_dir/bin" "$stage/bin"
  cp "$ROOT_DIR/LICENSE" "$stage/LICENSE"

  if [ "$is_main" = "true" ]; then
    cp "$ROOT_DIR/README.md" "$stage/README.md"
    jq --arg version "$NPM_VERSION" \
      '.version = $version | .optionalDependencies |= with_entries(.value = $version)' \
      "$stage/package.json" >"$stage/package.json.tmp"
  else
    jq --arg version "$NPM_VERSION" '.version = $version' \
      "$stage/package.json" >"$stage/package.json.tmp"
  fi
  mv "$stage/package.json.tmp" "$stage/package.json"

  staged_packages="$staged_packages $stage"
}

for platform in \
  darwin-amd64 darwin-arm64 \
  linux-amd64 linux-arm64 \
  windows-amd64 windows-arm64
do
  stage_package "$ROOT_DIR/npm/mcp-gitee-ent-$platform"
done
stage_package "$ROOT_DIR/npm/mcp-gitee-ent" true

for stage in $staged_packages; do
  npm publish "$stage" --dry-run --tag "$DIST_TAG" --userconfig "$npmrc"
done

if [ "$NPM_DRY_RUN" = "1" ]; then
  echo "npm publish dry-run complete (dist-tag: $DIST_TAG)"
  exit 0
fi

for stage in $staged_packages; do
  package_name=$(jq -r '.name' "$stage/package.json")
  published_version=$(npm view "$package_name@$NPM_VERSION" version --userconfig "$npmrc" 2>/dev/null || true)
  if [ "$published_version" = "$NPM_VERSION" ]; then
    echo "Skipping $package_name@$NPM_VERSION (already published)"
    continue
  fi
  npm publish "$stage" --tag "$DIST_TAG" --userconfig "$npmrc"
done
