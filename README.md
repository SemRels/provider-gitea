# provider-gitea

[![Latest Release](https://img.shields.io/github/v/release/SemRels/provider-gitea?label=version&color=blue)](https://github.com/SemRels/provider-gitea/releases/latest)
[![CI](https://github.com/SemRels/provider-gitea/actions/workflows/ci.yml/badge.svg)](https://github.com/SemRels/provider-gitea/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/SemRels/provider-gitea)](LICENSE)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/SemRels/provider-gitea/badge)](https://scorecard.dev/viewer/?uri=github.com/SemRels/provider-gitea)
[![REUSE status](https://api.reuse.software/badge/github.com/SemRels/provider-gitea)](https://api.reuse.software/info/github.com/SemRels/provider-gitea)

Publishes the semrel release to a Gitea instance.

This plugin is distributed as the standalone Go binary `semrel-plugin-provider-gitea`. Semrel executes the binary as a subprocess, provides plugin configuration through `SEMREL_PLUGIN_*` environment variables, provides release context through `SEMREL_*` environment variables, reads standard output, and treats exit code `0` as success and any non-zero exit code as failure. Install the binary in `~/.semrel/plugins/` or anywhere on your `$PATH`.

## Installation

### Binary

```bash
go install github.com/SemRels/provider-gitea/cmd/plugin@latest
```

### Docker

Pre-built, multi-platform images (linux/amd64, linux/arm64) are published to the GitHub Container Registry on every release:

```bash
docker pull ghcr.io/semrels/provider-gitea:latest
```

Images are signed with [cosign](https://github.com/sigstore/cosign) and include a full SBOM attestation. Verify the signature:

```bash
cosign verify ghcr.io/semrels/provider-gitea:latest \
  --certificate-identity-regexp 'https://github.com/SemRels/provider-gitea/.github/workflows/container.yml.*' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com
```


## Configuration

```yaml
plugins:
  - uses: "@semrel/provider-gitea"
    phase: release
    args:
      base_url: "https://gitea.example.com"
      token: "${GITEA_TOKEN}"
      owner: "acme"
      repo: "service-api"
      draft: "false"
      comment_on_prs: "true"
```

## Options

Every option can be set as an entry under `args` in `.semrel.yaml`. semrel passes each entry to the plugin as the environment variable shown in the table, so you can also set that variable directly (for example in your CI environment) instead of using `args`.

| Argument | Environment variable | Required | Description | Default |
| --- | --- | --- | --- | --- |
| `base_url` | `SEMREL_PLUGIN_BASE_URL` | Required | Base URL of the Gitea instance. | None |
| `token` | `SEMREL_PLUGIN_TOKEN` | Required | Gitea API token. | None |
| `owner` | `SEMREL_PLUGIN_OWNER` | Optional | Repository owner. Defaults from the git remote when available. | Derived from git remote |
| `repo` | `SEMREL_PLUGIN_REPO` | Optional | Repository name. Defaults from the git remote when available. | Derived from git remote |
| `draft` | `SEMREL_PLUGIN_DRAFT` | Optional | Create the release as a draft. | false |
| `comment_on_prs` | `SEMREL_PLUGIN_COMMENT_ON_PRS` | Optional | Comment on pull requests associated with commits in the release. Defaults to true; set to `false` to disable. The token needs repository read and issue write permissions. | true |

## `SEMREL_*` release context used

| Variable | Description |
| --- | --- |
| `SEMREL_VERSION` | Resolved release version for the current run. |
| `SEMREL_TAG_NAME` | Git tag name semrel will create or publish. |
| `SEMREL_NEXT_VERSION` | Next version computed by semrel for the release. |
| `SEMREL_CHANGELOG` | Generated changelog text for the release. |
| `SEMREL_DRY_RUN` | Whether semrel is running in dry-run mode. |

## Example behavior

The plugin creates a Gitea release for the current tag and uploads the generated release notes from semrel. By default it posts an idempotent release link comment to associated merged pull requests. Set `SEMREL_PLUGIN_COMMENT_ON_PRS` to `false` to disable comments.

## License

Apache-2.0
