# ADR 0126: releases and updates come from the fork

## Status

Superseded by ADR 0127. It superseded the distribution source named in ADR
0118; what that ADR says about reporting nothing still holds.

## Context

ADR 0118 pointed `kivgraph update`, both installers and the build scripts at
the repository the work was developed in. That work then moved to a fork of
the official project, where releases were cut. A binary that kept reading
releases from the development repository would update into builds the fork
does not publish.

## Decision

`internal/update.Repository`, `scripts/install.sh`, `scripts/install.ps1`,
`scripts/downloads.sh`, `scripts/build-mcpb.sh`, `scripts/build-server-json.sh`,
`server.json` and its test, the README, the installation docs, the landing
site and the publishing skill named the fork.

What was not changed, on purpose:

- **The MCP Registry name.** It is an identifier published under an owner
  namespace, not a repository name, and changing it would make a new server.
  Only the repository and website URLs inside it moved. Publishing to the
  registry stayed off.
- **The Go module path** stayed `github.com/Luqueee/kivgraph`.
- **Published notes** (`landing/src/content/releases/v0.10.0-dev.1.md`) and the
  text of ADR 0118 kept naming the repository they were written for.

## Consequences

- A binary installed from the development repository, including
  `v0.10.0-dev.1`, keeps looking for updates there and does not know about the
  fork; it has to be reinstalled once. Nothing redirects it.
- The first release published from the fork is `v0.10.0`, tagged on the commit
  that made this change, so the binaries it built update from the fork.
