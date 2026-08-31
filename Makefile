BINARY_NAME := mcp-gitee-ent
GO ?= go
OSES := darwin linux windows
ARCHS := amd64 arm64
RELEASE_TARGETS := darwin-amd64 darwin-arm64 linux-amd64 linux-arm64 linux-arm windows-amd64 windows-arm64
BUILD_OUTPUT_DIR ?= bin/platforms
RELEASE_DIR ?= release
VERSION ?=
NPM_VERSION ?= 0.0.0

GITEE ?= gitee
GITEE_HOSTNAME ?= gitee.com
GITEE_REPO ?= oschina/mcp-gitee-ent
GITEE_API_PREFIX ?= https://$(GITEE_HOSTNAME)/api/v5
RELEASE_NOTES_DIR ?= release-notes
PREVIOUS_TAG ?=
RELEASE_NOTE_LANGUAGE ?= Simplified Chinese
RELEASE_NOTE_AI ?= 1
RELEASE_NOTE_FORCE ?= 0

RELEASE_VERSION := $(patsubst v%,%,$(VERSION))
RELEASE_TAG := v$(RELEASE_VERSION)
RELEASE_NAME := $(RELEASE_TAG) 版本发布
RELEASE_NOTES_FILE ?= $(RELEASE_NOTES_DIR)/$(RELEASE_TAG).md
RELEASE_DIR_ABS := $(abspath $(RELEASE_DIR))

GIT_TAG := $(shell git describe --tags --exact-match 2>/dev/null)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || printf unknown)
GIT_COMMIT_FULL := $(shell git rev-parse HEAD 2>/dev/null || printf unknown)
GIT_VERSION := $(if $(GIT_TAG),$(patsubst v%,%,$(GIT_TAG)),dev+$(GIT_COMMIT))
BUILD_VERSION ?= $(GIT_VERSION)
BUILD_FLAGS = -trimpath -ldflags "-s -w -X gitee.com/oschina/mcp-gitee-ent/utils.Version=$(BUILD_VERSION)"
NPM_BUILD_DIR := $(abspath tmp/npm-binaries)

.PHONY: build clean test tidy build-all-platforms npm-copy-binaries npm-publish \
	clean-release release-version-check release-check release release-note release-publish

build:
	@mkdir -p bin
	@$(GO) build $(BUILD_FLAGS) -o bin/$(BINARY_NAME) .
	@echo "Build complete: bin/$(BINARY_NAME) ($(BUILD_VERSION))"

clean:
	@rm -rf bin
	@echo "Clean complete."

test:
	@$(GO) test ./...

tidy:
	@$(GO) mod tidy

release-version-check:
	@test -n "$(strip $(RELEASE_VERSION))" || { echo "VERSION is required (for example: make release VERSION=0.2.0)" >&2; exit 2; }
	@printf '%s\n' "$(RELEASE_VERSION)" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$$' || { echo "VERSION must be valid SemVer without build metadata: $(VERSION)" >&2; exit 2; }

release-check: release-version-check
	@test -z "$$(git status --porcelain)" || { echo "Working tree must be clean before building a release" >&2; exit 2; }
	@git tag --points-at HEAD | grep -Fqx "$(RELEASE_TAG)" || { echo "HEAD must be tagged $(RELEASE_TAG) before building a release" >&2; exit 2; }
	@test "$$(git cat-file -t "$(RELEASE_TAG)")" = "tag" || { echo "$(RELEASE_TAG) must be an annotated tag" >&2; exit 2; }
	@unformatted="$$(gofmt -l .)"; test -z "$$unformatted" || { echo "Go files are not formatted:" >&2; printf '%s\n' "$$unformatted" >&2; exit 2; }
	@git diff --check
	@$(GO) mod tidy -diff
	@$(GO) test ./...
	@$(GO) test -race ./...
	@$(GO) vet ./...

build-all-platforms:
	@set -eu; \
	mkdir -p "$(BUILD_OUTPUT_DIR)"; \
	for os in $(OSES); do \
		for arch in $(ARCHS); do \
			ext=""; \
			if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
			echo "Building $$os/$$arch..."; \
			CGO_ENABLED=0 GOOS="$$os" GOARCH="$$arch" $(GO) build $(BUILD_FLAGS) \
				-o "$(BUILD_OUTPUT_DIR)/$(BINARY_NAME)-$$os-$$arch$$ext" .; \
		done; \
	done

clean-release:
	@rm -rf "$(RELEASE_DIR)"
	@echo "Release directory cleaned."

release: release-check clean-release
	@mkdir -p "$(RELEASE_DIR)"
	@$(MAKE) build-all-platforms BUILD_OUTPUT_DIR="$(RELEASE_DIR_ABS)/binaries" BUILD_VERSION="$(RELEASE_VERSION)"
	@echo "Building linux/arm (GOARM=7)..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 $(GO) build \
		-trimpath -ldflags "-s -w -X gitee.com/oschina/mcp-gitee-ent/utils.Version=$(RELEASE_VERSION)" \
		-o "$(RELEASE_DIR_ABS)/binaries/$(BINARY_NAME)-linux-arm" .
	@set -eu; \
	stage_root="$(RELEASE_DIR_ABS)/stage"; \
	mkdir -p "$$stage_root"; \
	for target in $(RELEASE_TARGETS); do \
		os=$$(printf '%s' "$$target" | cut -d- -f1); \
		arch=$$(printf '%s' "$$target" | cut -d- -f2); \
		ext=""; binary="$(BINARY_NAME)"; \
		if [ "$$os" = "windows" ]; then ext=".exe"; binary="$(BINARY_NAME).exe"; fi; \
		source="$(RELEASE_DIR_ABS)/binaries/$(BINARY_NAME)-$$os-$$arch$$ext"; \
		stage="$$stage_root/$$target"; \
		mkdir -p "$$stage"; \
		cp "$$source" "$$stage/$$binary"; \
		cp README.md README_CN.md LICENSE "$$stage/"; \
		tar -czf "$(RELEASE_DIR_ABS)/$(BINARY_NAME)-$$target.tar.gz" \
			-C "$$stage" README.md README_CN.md LICENSE "$$binary"; \
	done; \
	rm -rf "$(RELEASE_DIR_ABS)/binaries" "$$stage_root"
	@cd "$(RELEASE_DIR)" && if command -v sha256sum >/dev/null 2>&1; then \
		sha256sum ./*.tar.gz > checksums.txt; \
	else \
		shasum -a 256 ./*.tar.gz > checksums.txt; \
	fi
	@echo "Release artifacts ready in $(RELEASE_DIR)/ for $(RELEASE_TAG)."

release-note: release-version-check
	@GITEE_BIN="$(GITEE)" \
		RELEASE_TAG="$(RELEASE_TAG)" \
		PREVIOUS_TAG="$(PREVIOUS_TAG)" \
		RELEASE_NOTES_FILE="$(RELEASE_NOTES_FILE)" \
		RELEASE_NOTE_LANGUAGE="$(RELEASE_NOTE_LANGUAGE)" \
		RELEASE_NOTE_AI="$(RELEASE_NOTE_AI)" \
		RELEASE_NOTE_FORCE="$(RELEASE_NOTE_FORCE)" \
		./scripts/generate-release-note.sh

release-publish: release-version-check
	@command -v curl >/dev/null 2>&1 || { echo "curl is required to publish release artifacts" >&2; exit 2; }
	@command -v jq >/dev/null 2>&1 || { echo "jq is required to publish release artifacts" >&2; exit 2; }
	@command -v "$(GITEE)" >/dev/null 2>&1 || { echo "$(GITEE) is required to publish the Gitee Release" >&2; exit 2; }
	@test -s "$(RELEASE_NOTES_FILE)" || { echo "Release notes not found: $(RELEASE_NOTES_FILE)" >&2; echo "Run make release-note VERSION=$(RELEASE_VERSION), review the draft, and commit it before publishing." >&2; exit 2; }
	@remote_target="$$(git ls-remote origin "refs/tags/$(RELEASE_TAG)^{}" | awk '{print $$1}')"; \
		test "$$remote_target" = "$(GIT_COMMIT_FULL)" || { echo "Push annotated tag $(RELEASE_TAG) for $(GIT_COMMIT_FULL) to origin before publishing" >&2; exit 2; }
	@$(MAKE) release VERSION="$(VERSION)" RELEASE_DIR="$(RELEASE_DIR)"
	@GITEE_BIN="$(GITEE)" \
		GITEE_HOSTNAME="$(GITEE_HOSTNAME)" \
		GITEE_REPO="$(GITEE_REPO)" \
		GITEE_API_PREFIX="$(GITEE_API_PREFIX)" \
		RELEASE_TAG="$(RELEASE_TAG)" \
		RELEASE_NAME="$(RELEASE_NAME)" \
		RELEASE_TARGET="$(GIT_COMMIT_FULL)" \
		RELEASE_NOTES_FILE="$(RELEASE_NOTES_FILE)" \
		RELEASE_DIR="$(RELEASE_DIR_ABS)" \
		./scripts/publish-release.sh

npm-copy-binaries:
	@rm -rf "$(NPM_BUILD_DIR)"
	@mkdir -p "$(NPM_BUILD_DIR)"
	@$(MAKE) build-all-platforms BUILD_OUTPUT_DIR="$(NPM_BUILD_DIR)"
	@set -eu; \
	for os in $(OSES); do \
		for arch in $(ARCHS); do \
			ext=""; \
			if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
			name="$(BINARY_NAME)-$$os-$$arch"; \
			mkdir -p "npm/$$name/bin"; \
			cp "$(NPM_BUILD_DIR)/$$name$$ext" "npm/$$name/bin/$$name$$ext"; \
		done; \
	done; \
	rm -rf "$(NPM_BUILD_DIR)"
	@echo "Binaries copied to npm platform packages."

npm-publish: npm-copy-binaries
	@NPM_VERSION="$(NPM_VERSION)" NPM_TOKEN="$(NPM_TOKEN)" NPM_TAG="$(NPM_TAG)" NPM_DRY_RUN="$(NPM_DRY_RUN)" ./scripts/publish-npm.sh
	@echo "npm publish complete."
