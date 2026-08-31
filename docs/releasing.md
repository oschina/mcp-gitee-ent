# 发布流程

本仓库使用本地、可重复执行的发布命令。它不会在推送 tag 后由 CI 自动触发；发布者需要在已认证 Gitee CLI 的环境中主动运行 `make release-publish`。

Gitee Release 和 npm 发布分为两个步骤。Gitee Release 会自动创建或复用并上传缺失附件；npm 包会逐个发布，因此也设计为可以在中断后安全重跑。

## 前置条件

- Go 1.23 或更高版本
- Make、`curl`、`jq`
- 已执行 `gitee auth login`，或者在发布环境中设置 `GITEE_TOKEN`
- 发布 npm 时需要 Node.js、npm 和 `NPM_TOKEN`
- `origin` 指向 `oschina/mcp-gitee-ent`

发布版本使用 Semantic Versioning，命令中的 `VERSION` 不带 `v`，Git tag 会自动使用 `v` 前缀。新增大块向后兼容能力建议提升次版本，例如 `0.2.0`；修复版本使用 `0.1.12` 形式。

## 准备发布

1. 同步远端标签：

```bash
git fetch origin --tags
```

2. 提交准备发布的源码，确认工作区干净，并运行基础测试：

```bash
git status --short
go test ./...
```

3. 生成可编辑的发布说明：

```bash
make release-note VERSION=0.2.0
```

默认输出到 `release-notes/v0.2.0.md`。命令会寻找最近的版本 tag，并根据提交记录、改动统计、文件列表和截断后的补丁生成草稿。

默认会调用当前 Gitee CLI 配置的 AI 服务，相关源码差异会被发送给该服务。需要完全在本地生成基于提交记录的草稿时使用：

```bash
make release-note VERSION=0.2.0 RELEASE_NOTE_AI=0
```

特殊提交历史可以显式指定比较基准：

```bash
make release-note VERSION=0.2.0 PREVIOUS_TAG=v0.1.11
```

命令不会覆盖已有草稿。确需重新生成时设置 `RELEASE_NOTE_FORCE=1`。发布前必须人工复核并提交发布说明。

4. 在最终提交上创建 annotated tag：

```bash
git tag -a v0.2.0 -m "mcp-gitee-ent v0.2.0"
```

## 本地验证

在 tag 对应的干净工作区执行：

```bash
make release VERSION=0.2.0
```

发布门禁会检查：

- `VERSION` 是否为合法 SemVer
- 工作区是否干净
- HEAD 是否有名称匹配的 annotated tag
- `gofmt` 和 `git diff --check`
- `go mod tidy -diff`
- `go test ./...`
- `go test -race ./...`
- `go vet ./...`

构建产物位于 `release/`：

- macOS: amd64、arm64
- Linux: amd64、arm64，以及兼容已有下载的 32 位 ARMv7
- Windows: amd64、arm64
- `checksums.txt`

压缩包名称保持 `mcp-gitee-ent-<os>-<arch>.tar.gz`，以兼容已有 Release 附件命名。

## 发布 Gitee Release

先推送最终提交和 tag：

```bash
git push origin master
git push origin v0.2.0
```

再执行完整发布：

```bash
make release-publish VERSION=0.2.0
```

该命令会重新执行全部发布门禁和构建，然后：

1. 按 tag 查找现有 Gitee Release；不存在时自动创建。
2. 从 `release-notes/v0.2.0.md` 同步名称和正文。
3. 版本包含预发布后缀时自动设置 prerelease，例如 `0.2.0-rc.1`。
4. 查询已有附件并按文件名跳过，上传其余压缩包和 `checksums.txt`。

发布中断后可直接重新运行同一个命令。脚本不会删除或覆盖已有附件；同名附件会被跳过。

## 发布 npm

先执行不会发布的完整 dry-run：

```bash
make npm-copy-binaries
NPM_VERSION=0.2.0 NPM_TOKEN=validation-only NPM_DRY_RUN=1 ./scripts/publish-npm.sh
```

确认 Gitee Release 成功后再发布：

```bash
make npm-publish NPM_VERSION=0.2.0 NPM_TOKEN=...
```

npm 脚本在临时目录中写入版本号，不修改仓库内的 `package.json`，也不会复制本地 `.npmrc`。它先验证所有平台包，再发布平台包，最后发布主包；已存在的相同版本会被跳过。预发布版本默认使用 npm `next` dist-tag，可通过 `NPM_TAG` 覆盖。

## 发布后验证

1. 检查 Gitee Release 的 tag、正文和附件数量。
2. 下载附件并使用 `checksums.txt` 校验 SHA-256。
3. 运行一个下载的二进制：

```bash
./mcp-gitee-ent --version
```

4. 从 npm 安装并再次检查版本和至少一个只读 MCP 工具。
