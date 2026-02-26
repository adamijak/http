# CI/CD Documentation

This repository uses GitHub Actions for automated testing and releasing.

## Workflows

### Test Workflow (`test.yml`)

**Trigger**: Runs on every push and pull request to `main`/`master` branches.

**What it does**:
- Tests the code on multiple platforms (Linux, macOS, Windows)
- Tests with multiple Go versions (1.23, 1.24)
- Runs the following checks:
  - Go vet (static analysis)
  - Go fmt (code formatting)
  - Executes `test.sh` script
  - Builds the binary
- Runs golangci-lint for comprehensive linting

**Matrix Testing**:
- **Go versions**: 1.23, 1.24
- **Operating Systems**: Ubuntu, macOS, Windows
- **Total combinations**: 6 test runs per push

### Release Workflow (`release.yml`)

**Trigger**: Runs when a tag matching `v*` is pushed (e.g., `v1.0.0`, `v1.2.3`).

**What it does**:
- Builds binaries for multiple platforms and architectures
- Creates checksums (SHA256) for each binary
- Creates a GitHub Release with all binaries attached
- Generates comprehensive release notes with installation instructions

**Build Matrix**:
- **Linux**: amd64, arm64, 386
- **macOS**: amd64 (Intel), arm64 (Apple Silicon)
- **Windows**: amd64, 386
- **FreeBSD**: amd64

**Features**:
- Version information is embedded in binaries from git tag
- SHA256 checksums for verification
- Automated release notes generation
- All binaries are stripped and optimized (`-ldflags="-s -w"`)

## Creating a Release

To create a new release:

1. **Update version** (if needed):
   ```bash
   # Update version.go if you want to update the version constant
   vim version.go
   git add version.go
   git commit -m "Bump version to X.Y.Z"
   git push
   ```

2. **Create and push a tag**:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

3. **Watch the workflow**:
   - Go to the Actions tab in GitHub
   - The Release workflow will automatically start
   - It will build binaries and create a release
   - The release will appear in the Releases section

## Downloading Releases

Users can download pre-built binaries from the [Releases page](https://github.com/adamijak/http/releases).

### Quick install example:

```bash
# Linux AMD64
wget https://github.com/adamijak/http/releases/latest/download/http-linux-amd64
chmod +x http-linux-amd64
sudo mv http-linux-amd64 /usr/local/bin/http

# macOS ARM64 (Apple Silicon)
wget https://github.com/adamijak/http/releases/latest/download/http-darwin-arm64
chmod +x http-darwin-arm64
sudo mv http-darwin-arm64 /usr/local/bin/http
```

## Verify Checksums

Each binary comes with a SHA256 checksum file:

```bash
# Download binary and checksum
wget https://github.com/adamijak/http/releases/latest/download/http-linux-amd64
wget https://github.com/adamijak/http/releases/latest/download/http-linux-amd64.sha256

# Verify
sha256sum -c http-linux-amd64.sha256
```

## Local Testing

Before pushing changes, you can test locally:

```bash
# Run tests
bash test.sh

# Check formatting
gofmt -s -l .

# Run vet
go vet ./...

# Build
go build -o http
```

## Troubleshooting

### Test failures

If tests fail in CI:
1. Check the Actions tab for detailed logs
2. Run the same test locally: `bash test.sh`
3. Fix issues and push again

### Release failures

If release workflow fails:
1. Check that the tag follows the `v*` format (e.g., `v1.0.0`)
2. Verify Go code compiles: `go build`
3. Check the Actions tab for detailed build logs
4. Delete the tag if needed and recreate: 
   ```bash
   git tag -d v1.0.0
   git push origin :refs/tags/v1.0.0
   ```

## Workflow Badges

Add these badges to your README to show build status:

```markdown
![Test](https://github.com/adamijak/http/workflows/Test/badge.svg)
![Release](https://github.com/adamijak/http/workflows/Release/badge.svg)
```
