# Release Process

This document describes the process for creating a new release of StreamV3.

## Pre-Release Checklist

Before creating a new release tag, ensure:

1. ✅ All code changes are complete and tested
2. ✅ Documentation is updated (especially CLI codelab if CLI changed)
3. ✅ CHANGELOG is updated with new features/fixes
4. ✅ **Version string is updated in `cmd/streamv3/main.go`**

## Release Steps

### 1. Update Version String

**IMPORTANT**: This must be done BEFORE creating the tag!

```bash
# Edit cmd/streamv3/main.go
# Change line 27 from:
#   fmt.Println("streamv3 version X.Y.Z")
# To:
#   fmt.Println("streamv3 version X.Y.NEW")
```

### 2. Commit Version Update

```bash
git add cmd/streamv3/main.go
git commit -m "Bump version to vX.Y.NEW"
```

### 3. Create and Push Tag

```bash
# Create annotated tag
git tag -a vX.Y.NEW -m "vX.Y.NEW: Brief description

## What's New
- Feature 1
- Feature 2

## Bug Fixes
- Fix 1

## Examples
..."

# Push commit and tag together
git push && git push --tags
```

### 4. Verify Release

```bash
# Verify tag is correct
git describe --tags

# Test installation
GOPROXY=direct go install github.com/rosscartlidge/streamv3/cmd/streamv3@latest

# Verify version
streamv3 -version
# Should show: streamv3 version X.Y.NEW
```

## Version Numbering

We follow semantic versioning (semver):

- **Major (X.0.0)**: Breaking changes to API or CLI
- **Minor (0.X.0)**: New features, backward compatible
- **Patch (0.0.X)**: Bug fixes, backward compatible

Examples:
- `v0.7.0` - Added Phase 1 SQL operations (JOIN, DISTINCT, OFFSET, UNION)
- `v0.7.1` - Complete code generation with Chain() pattern
- `v0.7.2` - Added regexp/regex aliases (minor improvement)
- `v0.7.3` - Fixed version string display (patch fix)

## Common Mistakes to Avoid

### ❌ Creating tag before updating version string
This causes the installed binary to show the old version number.

**Wrong order:**
1. git tag vX.Y.Z
2. git push --tags
3. Update main.go (too late!)

**Correct order:**
1. Update main.go version string
2. git commit
3. git tag vX.Y.Z
4. git push && git push --tags

### ❌ Forgetting to push tags
Tags are not pushed by default with `git push`.

**Solution:** Always use `git push --tags` or `git push && git push --tags`

### ❌ Using lightweight tags instead of annotated tags
Lightweight tags don't include metadata.

**Solution:** Always use `git tag -a` for releases

## Troubleshooting

### If you forgot to update the version string

If you've already created and pushed a tag:

```bash
# 1. Update version string in main.go
vim cmd/streamv3/main.go

# 2. Commit
git add cmd/streamv3/main.go
git commit -m "Bump version to vX.Y.Z"

# 3. Delete old tag locally
git tag -d vX.Y.Z

# 4. Create new tag
git tag -a vX.Y.Z -m "..."

# 5. Push commit and force-push tag
git push && git push --force origin vX.Y.Z
```

**Better:** Just increment to the next patch version and do it correctly.

## Release Announcement

After release:
1. Update GitHub releases page with tag notes
2. Announce in relevant channels
3. Update README if needed

---

**Remember**: The version string in `cmd/streamv3/main.go` must match the git tag!
