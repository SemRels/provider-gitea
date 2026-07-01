# provider-gitea

[![Latest Release](https://img.shields.io/github/v/release/SemRels/provider-gitea?label=version&color=blue)](https://github.com/SemRels/provider-gitea/releases/latest)

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
  --certificate-identity-regexp 'https://github.com/SemRels/provider-gitea/.github/workflows/release.yml.*' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com
```


## Configuration

```yaml
plugins:
  - name: provider-gitea
    path: ~/.semrel/plugins/semrel-plugin-provider-gitea
    env:
      SEMREL_PLUGIN_BASE_URL: "https://gitea.example.com"
      SEMREL_PLUGIN_TOKEN: "${GITEA_TOKEN}"
      SEMREL_PLUGIN_OWNER: "acme"
      SEMREL_PLUGIN_REPO: "service-api"
      SEMREL_PLUGIN_DRAFT: "false"
```

## `SEMREL_PLUGIN_*` variables

| Name | Required | Description | Default |
| --- | --- | --- | --- |
| `SEMREL_PLUGIN_BASE_URL` | Required | Base URL of the Gitea instance. | None |
| `SEMREL_PLUGIN_TOKEN` | Required | Gitea API token. | None |
| `SEMREL_PLUGIN_OWNER` | Optional | Repository owner. Defaults from the git remote when available. | Derived from git remote |
| `SEMREL_PLUGIN_REPO` | Optional | Repository name. Defaults from the git remote when available. | Derived from git remote |
| `SEMREL_PLUGIN_DRAFT` | Optional | Create the release as a draft. | false |

## `SEMREL_*` release context used

| Variable | Description |
| --- | --- |
| `SEMREL_VERSION` | Resolved release version for the current run. |
| `SEMREL_TAG_NAME` | Git tag name semrel will create or publish. |
| `SEMREL_NEXT_VERSION` | Next version computed by semrel for the release. |
| `SEMREL_CHANGELOG` | Generated changelog text for the release. |
| `SEMREL_DRY_RUN` | Whether semrel is running in dry-run mode. |

## Example behavior

The plugin creates a Gitea release for the current tag and uploads the generated release notes from semrel.

## License

Apache-2.0
