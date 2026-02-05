# Dokploy CLI (Go)

A minimal Go-based CLI to manage Dokploy projects, compose apps, and domains (using each project's default environment) via the Dokploy HTTP API.

The binary exposes a single top-level command, `dokploy`, with subcommands for each resource.

For full CLI usage, flags, and examples, see [USAGE.md](USAGE.md).

## Build

You can build the CLI locally with Go into a build directory. The binary name follows the convention `dokploy_${platform}_${arch}` (with `.exe` on Windows):

```bash
# Example: build for your current platform
go build -o build/dokploy_$(go env GOOS)_$(go env GOARCH) .

# Example: explicitly build a Windows amd64 binary (from any OS)
GOOS=windows GOARCH=amd64 go build -o build/dokploy_windows_amd64.exe .

# Example: explicitly build a Linux amd64 binary
GOOS=linux GOARCH=amd64 go build -o build/dokploy_linux_amd64 .
```

This produces the binary in the build directory using the `dokploy_${platform}_${arch}` naming scheme (with an added `.exe` suffix on Windows).

### Version information

Release builds inject the Git tag and commit into the CLI. You can see this with:

```bash
dokploy --version
```

This prints the version in the form `vX.Y.Z (commit-sha)` for binaries built by the GitHub Actions workflow.

## Usage and examples

See here [USAGE.md](USAGE.md)

---

## Testing

This repository includes unit tests, integration tests, and optional end-to-end (e2e) tests for the Dokploy client and helpers.

- Run all tests (root + subpackages):

  ```bash
  go test -v ./...
  ```

- Run integration tests (fake Dokploy HTTP server, end-to-end project + default environment flow):

  ```bash
  go test -v -run Integration
  ```

- Run end-to-end (e2e) Dokploy API tests (against a real Dokploy instance):

  ```bash
  export DOKPLOY_URL="https://your-dokploy-instance.com"
  export DOKPLOY_API_KEY="YOUR-GENERATED-API-KEY"

  # Run only the e2e tests (requires -tags e2e)
  go test -v -tags e2e . -run E2E
  ```

  These e2e tests perform a small create/delete flow (project + its default environment) against the configured Dokploy instance. They are automatically skipped if `DOKPLOY_URL` or `DOKPLOY_API_KEY` is not set.
