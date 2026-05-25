# provider-gitea

`provider-gitea` is a SemRels subprocess plugin that creates Gitea releases over the REST API. It does not use gRPC or `hashicorp/go-plugin`.

## Environment

Required:

- `SEMREL_PLUGIN_BASE_URL`
- `SEMREL_PLUGIN_TOKEN`
- `SEMREL_PLUGIN_OWNER`
- `SEMREL_PLUGIN_REPO`
- `SEMREL_PLUGIN_TAG_NAME`, `SEMREL_TAG_NAME`, `SEMREL_VERSION`, or `SEMREL_NEXT_VERSION`

Optional:

- `SEMREL_PLUGIN_NAME` (defaults to the resolved tag name)
- `SEMREL_PLUGIN_BODY` (defaults to `SEMREL_CHANGELOG`)
- `SEMREL_PLUGIN_DRAFT`
- `SEMREL_PLUGIN_PRERELEASE` (falls back to `SEMREL_IS_PRERELEASE`)
- `SEMREL_DRY_RUN`

## Example `.semrel.yaml`

```yaml
plugins:
  - uses: gitea
    path: ./bin/semrel-plugin-gitea
    args:
      base-url: https://gitea.example.com
      token: ${GITEA_TOKEN}
      owner: SemRels
      repo: provider-gitea
      draft: false
```

## Development

```powershell
go build ./cmd/plugin
go test -cover ./...
```

## Repository layout

```text
cmd/plugin/        Subprocess entrypoint
internal/plugin/   Gitea release client and env parsing
.github/workflows/ CI, release, and security automation
```
