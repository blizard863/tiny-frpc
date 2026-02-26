# AGENTS.md

## Cursor Cloud specific instructions

This is **tiny-frpc**, a lightweight reverse proxy client that communicates with frps via SSH tunnel gateway. It is a single Go module (not a monorepo) producing two binary variants.

### Go toolchain

Requires Go 1.23+. The VM's update script installs Go 1.23.6 to `/usr/local/go` and symlinks `/usr/bin/go` to it. If `go version` shows an older version or errors about "toolchain not available", verify `/usr/local/go/bin/go` exists and `/usr/bin/go` points to it.

### Key commands (all from repo root)

| Task | Command |
|------|---------|
| Lint | `go vet ./...` |
| Format | `go fmt ./...` (or `make fmt`) |
| Test | `make test` (runs `go test -v --cover ./...`) |
| Build both variants | `make build` |
| Build standalone only | `make gssh` → `bin/tiny-frpc` |
| Build native-ssh only | `make nssh` → `bin/tiny-frpc-ssh` |
| Run | `./bin/tiny-frpc -c <config.toml>` |

### E2E testing notes

- tiny-frpc requires an **frps** server (from [fatedier/frp](https://github.com/fatedier/frp)) with `sshTunnelGateway.bindPort` configured. For local testing, download a frps release binary and run with a minimal config.
- The standalone variant (`tiny-frpc`) reads `~/.ssh/id_rsa` for SSH authentication. Generate one with `ssh-keygen -t rsa -f ~/.ssh/id_rsa -N ""` if missing.
- No database or external services are needed beyond frps and a local backend to proxy.
