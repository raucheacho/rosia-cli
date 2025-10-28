# Release Guide

This document describes how to create a new release of Rosia CLI.

## Prerequisites

1. **GoReleaser installed**: Install from https://goreleaser.com/install/
2. **GitHub Personal Access Tokens** (for automated tap/bucket updates):
   - `HOMEBREW_TAP_GITHUB_TOKEN`: Token with repo access for homebrew-rosia repository
   - `SCOOP_BUCKET_GITHUB_TOKEN`: Token with repo access for scoop-rosia repository
3. **Git configured**: Ensure you have commit access to the repository
4. **Clean working directory**: Commit or stash all changes

## Release Process

### 1. Prepare the Release

```bash
# Ensure you're on the main branch
git checkout main
git pull origin main

# Run tests to ensure everything works
go test ./...

# Update version in documentation if needed
# Update CHANGELOG.md with release notes
```

### 2. Create and Push Tag

```bash
# Create a new tag (use semantic versioning)
git tag -a v0.1.0 -m "Release v0.1.0"

# Push the tag to GitHub
git push origin v0.1.0
```

### 3. Automated Release (via GitHub Actions)

Once you push the tag, GitHub Actions will automatically:
- Run tests
- Build binaries for all platforms
- Create GitHub release with binaries
- Update Homebrew tap (if token is configured)
- Update Scoop bucket (if token is configured)
- Build and push Docker images

Monitor the release at: https://github.com/raucheacho/rosia-cli/actions

### 4. Manual Release (Local)

If you prefer to release manually or need to test locally:

```bash
# Ensure GoReleaser is installed
goreleaser --version

# Test the release process (dry run)
goreleaser release --snapshot --clean

# Create actual release
export GITHUB_TOKEN="your_github_token"
export HOMEBREW_TAP_GITHUB_TOKEN="your_homebrew_token"
export SCOOP_BUCKET_GITHUB_TOKEN="your_scoop_token"

goreleaser release --clean
```

## Post-Release Tasks

### 1. Verify Release

- Check GitHub releases page: https://github.com/raucheacho/rosia-cli/releases
- Verify all platform binaries are present
- Test download and installation on different platforms

### 2. Test Installation Methods

**Homebrew (macOS/Linux)**:
```bash
brew tap raucheacho/rosia
brew install rosia
rosia version
```

**Scoop (Windows)**:
```powershell
scoop bucket add rosia https://github.com/raucheacho/scoop-rosia
scoop install rosia
rosia version
```

**Direct Download**:
```bash
# Linux/macOS
curl -fsSL https://raw.githubusercontent.com/raucheacho/rosia-cli/main/install.sh | bash

# Windows (PowerShell)
iwr -useb https://raw.githubusercontent.com/raucheacho/rosia-cli/main/install.ps1 | iex
```

**Docker**:
```bash
docker pull ghcr.io/raucheacho/rosia:latest
docker run --rm ghcr.io/raucheacho/rosia:latest version
```

### 3. Update Documentation

- Update README.md with new version number if needed
- Update any version-specific documentation
- Announce the release (Twitter, blog, etc.)

### 4. Monitor Issues

- Watch for installation issues
- Monitor GitHub issues and discussions
- Be ready to create a patch release if critical bugs are found

## Versioning Strategy

Rosia CLI follows [Semantic Versioning](https://semver.org/):

- **MAJOR** version (v1.0.0): Incompatible API changes
- **MINOR** version (v0.1.0): New functionality in a backward compatible manner
- **PATCH** version (v0.1.1): Backward compatible bug fixes

### Pre-release Versions

For testing before official release:
- Alpha: `v0.1.0-alpha.1`
- Beta: `v0.1.0-beta.1`
- Release Candidate: `v0.1.0-rc.1`

## Rollback Procedure

If a release has critical issues:

1. **Delete the GitHub release** (if not yet widely distributed)
2. **Delete the tag**:
   ```bash
   git tag -d v0.1.0
   git push origin :refs/tags/v0.1.0
   ```
3. **Fix the issue** and create a new patch release
4. **Communicate** the issue and fix to users

## Troubleshooting

### GoReleaser Fails

- Check `.goreleaser.yml` syntax
- Ensure all required environment variables are set
- Verify GitHub token has correct permissions
- Check build logs for specific errors

### Homebrew Tap Not Updated

- Verify `HOMEBREW_TAP_GITHUB_TOKEN` is set and valid
- Check that homebrew-rosia repository exists
- Manually update the formula if needed

### Scoop Bucket Not Updated

- Verify `SCOOP_BUCKET_GITHUB_TOKEN` is set and valid
- Check that scoop-rosia repository exists
- Manually update the manifest if needed

### Docker Images Not Published

- Verify GitHub Container Registry permissions
- Check Docker build logs in GitHub Actions
- Ensure `GITHUB_TOKEN` has package write permissions

## Release Checklist

- [ ] All tests passing
- [ ] CHANGELOG.md updated
- [ ] Version bumped appropriately
- [ ] Tag created and pushed
- [ ] GitHub release created successfully
- [ ] All platform binaries present
- [ ] Homebrew formula updated (if applicable)
- [ ] Scoop manifest updated (if applicable)
- [ ] Docker images published (if applicable)
- [ ] Installation tested on multiple platforms
- [ ] Documentation updated
- [ ] Release announced

## Contact

For questions about the release process, open an issue or discussion on GitHub.
