# ADR 0127: the distribution names the official project

## Status

Accepted. Supersedes ADR 0126 and the repository decision of ADR 0118; what
ADR 0118 says about reporting nothing still holds.

## Context

ADR 0118 and ADR 0126 pointed `kivgraph update`, the installers, the build
scripts and the documentation at the repository the work was developed in. That
made every page, command and manifest in this tree name a repository other than
the project's own, and a pull request carrying the work back had to undo it in a
separate commit.

## Decision

`internal/update.Repository`, `scripts/install.sh`, `scripts/install.ps1`,
`scripts/downloads.sh`, `scripts/build-mcpb.sh`, `scripts/build-server-json.sh`,
`server.json` and its test, `glama.json`, the README, the installation docs,
the landing site and the publishing skill name `Luqueee/kivgraph`. The MCP
Registry name is `io.github.Luqueee/kivgraph`.

What does not change:

- **Telemetry.** Nothing is sent. `telemetry.Endpoint` stays empty and the
  installers make no request (ADR 0118).
- **The MCP Registry job** stays behind `PUBLISH_MCP_REGISTRY`. The registry
  grants a namespace to the repository that authenticates, so only the
  official repository can publish under that name.
- **The install commands** keep the GitHub release URLs; no domain is
  introduced.
- **The Go module path** was `github.com/Luqueee/kivgraph` all along.

## Consequences

- The documented install commands download the official project's latest
  release. That is a build of the official repository: it is the version
  published there, not this tree, and it follows that project's telemetry
  policy, not ADR 0118.
- `kivgraph update` in a binary built from this tree reads the official
  releases, and offers one only when its version is higher.
- A release cut from a repository other than the official one would publish
  installers that download from the official one. Releases are cut there.
- Earlier documents no longer name the repositories the work passed through;
  they say "the development repository" and "the fork".
