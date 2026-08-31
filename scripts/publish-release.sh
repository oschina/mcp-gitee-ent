#!/bin/sh

set -eu

GITEE_BIN=${GITEE_BIN:-gitee}
GITEE_HOSTNAME=${GITEE_HOSTNAME:-gitee.com}
GITEE_REPO=${GITEE_REPO:-oschina/mcp-gitee-ent}
GITEE_API_PREFIX=${GITEE_API_PREFIX:-https://$GITEE_HOSTNAME/api/v5}
RELEASE_TAG=${RELEASE_TAG:-}
RELEASE_NAME=${RELEASE_NAME:-$RELEASE_TAG}
RELEASE_TARGET=${RELEASE_TARGET:-master}
RELEASE_NOTES_FILE=${RELEASE_NOTES_FILE:-}
RELEASE_DIR=${RELEASE_DIR:-release}

if [ -z "$RELEASE_TAG" ]; then
  echo "RELEASE_TAG is required" >&2
  exit 2
fi
case "$GITEE_REPO" in
  */*) ;;
  *)
    echo "GITEE_REPO must use owner/repo format" >&2
    exit 2
    ;;
esac
if [ ! -d "$RELEASE_DIR" ]; then
  echo "Release directory not found: $RELEASE_DIR" >&2
  exit 2
fi
if [ -z "$RELEASE_NOTES_FILE" ] || [ ! -s "$RELEASE_NOTES_FILE" ]; then
  echo "Release notes file not found: $RELEASE_NOTES_FILE" >&2
  exit 2
fi
for command_name in "$GITEE_BIN" curl jq; do
  if ! command -v "$command_name" >/dev/null 2>&1; then
    echo "Required command not found: $command_name" >&2
    exit 2
  fi
done

token=${GITEE_TOKEN:-}
if [ -z "$token" ]; then
  token=$("$GITEE_BIN" auth token --hostname "$GITEE_HOSTNAME")
fi
if [ -z "$token" ]; then
  echo "No Gitee token available; run gitee auth login or set GITEE_TOKEN" >&2
  exit 2
fi

release_json=$("$GITEE_BIN" release view "$RELEASE_TAG" \
  --hostname "$GITEE_HOSTNAME" -R "$GITEE_REPO" --json=id --no-tui 2>/dev/null || true)
release_id=$(printf '%s' "$release_json" | jq -r '(.id // 0) | select(. > 0)' 2>/dev/null || true)

if [ -z "$release_id" ]; then
  set -- release create --hostname "$GITEE_HOSTNAME" -R "$GITEE_REPO" \
    --tag "$RELEASE_TAG" --name "$RELEASE_NAME" --target "$RELEASE_TARGET" \
    --body "$(cat "$RELEASE_NOTES_FILE")" --no-tui
  case "$RELEASE_TAG" in
    *-*) set -- "$@" --prerelease ;;
  esac
  "$GITEE_BIN" "$@"

  release_json=$("$GITEE_BIN" release view "$RELEASE_TAG" \
    --hostname "$GITEE_HOSTNAME" -R "$GITEE_REPO" --json=id --no-tui)
  release_id=$(printf '%s' "$release_json" | jq -r '(.id // 0) | select(. > 0)')
fi

case "$release_id" in
  ''|*[!0-9]*)
    echo "Could not resolve numeric release ID for $RELEASE_TAG" >&2
    exit 1
    ;;
esac

owner=${GITEE_REPO%%/*}
repo=${GITEE_REPO#*/}
attachments_url="$GITEE_API_PREFIX/repos/$owner/$repo/releases/$release_id/attach_files"
release_url="$GITEE_API_PREFIX/repos/$owner/$repo/releases/$release_id"

# Keep the token out of process arguments and command output.
authenticated_curl() {
  printf 'header = "Authorization: Bearer %s"\n' "$token" | \
    curl --config - --fail --silent --show-error "$@"
}

prerelease=false
case "$RELEASE_TAG" in
  *-*) prerelease=true ;;
esac
updated_release=$(authenticated_curl --request PATCH \
  --data-urlencode "tag_name=$RELEASE_TAG" \
  --data-urlencode "name=$RELEASE_NAME" \
  --data-urlencode "body@$RELEASE_NOTES_FILE" \
  --data-urlencode "prerelease=$prerelease" \
  "$release_url")
updated_release_id=$(printf '%s' "$updated_release" | jq -r '(.id // 0) | select(. > 0)')
if [ "$updated_release_id" != "$release_id" ]; then
  echo "Unexpected response while updating release notes for $RELEASE_TAG" >&2
  exit 1
fi
echo "Updated release notes from $RELEASE_NOTES_FILE"

attachments=$(authenticated_curl "$attachments_url?per_page=100")
if ! printf '%s' "$attachments" | jq -e 'type == "array"' >/dev/null; then
  echo "Unexpected response while listing release attachments" >&2
  exit 1
fi

found_artifact=false
for artifact in "$RELEASE_DIR"/*.tar.gz "$RELEASE_DIR"/*.zip "$RELEASE_DIR"/checksums.txt; do
  if [ ! -f "$artifact" ]; then
    continue
  fi
  found_artifact=true
  name=$(basename "$artifact")
  if printf '%s' "$attachments" | jq -e --arg name "$name" 'any(.[]; .name == $name)' >/dev/null; then
    echo "Skipping $name (already uploaded)"
    continue
  fi

  response=$(authenticated_curl --form "file=@$artifact" "$attachments_url")
  uploaded_name=$(printf '%s' "$response" | jq -r '.name // empty')
  if [ "$uploaded_name" != "$name" ]; then
    echo "Unexpected response after uploading $name" >&2
    exit 1
  fi
  echo "Uploaded $name"
done

if [ "$found_artifact" != "true" ]; then
  echo "No release artifacts found in $RELEASE_DIR" >&2
  exit 2
fi

echo "Published $RELEASE_TAG to https://$GITEE_HOSTNAME/$GITEE_REPO/releases/tag/$RELEASE_TAG"
