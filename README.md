# provider-gitea

Gitea provider plugin for Semantic Release.

Provides Gitea repository, release, and metadata integration for Semantic Release.

## Documentation

- Docs (coming soon): <https://github.com/SemRels/semrel/tree/main/docs/plugins/provider-gitea>
- Template source: <https://github.com/SemRels/plugin-template>

## Repository Layout

`	ext
cmd/plugin/              Plugin entry point
internal/plugin/         Business logic scaffold
internal/grpc/           gRPC transport scaffold
proto/v1                 Symlink to the SemRel protobuf contract
.github/workflows/       CI, release, and security automation
`

## Development

`ash
go build ./cmd/plugin
go test ./...
`

## Configuration Example

`yaml
plugins:
  - name: provider-gitea
    type: provider
    config:
      api_url: https://gitea.example.com/api/v1
      owner: semrels
      repository: example-repo
      token: ${GITEA_TOKEN}
`

## Status

This repository is bootstrapped from SemRels/plugin-template and is ready for implementation.