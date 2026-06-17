---
description: >
  Read this when cutting a release of croni — whenever I say "cut a release",
  "ship it", "tag a version", or ask to update release notes. Covers the full
  release checklist, semver choice, and keeping docs in sync.
read_when: cutting a release, tagging vX.Y.Z, editing release notes, bumping the version
---

# Releasing

## Checklist

1. **Update the README** (and any other user-facing docs) so they describe the
   version being released — new commands, flags, behavior changes. The README is
   the primary user-facing documentation; it must match the release.

2. **Run checks locally**:

   ```bash
   make vet && make test
   ```

3. **Commit all changes and push to main**:

   ```bash
   git push origin main
   ```

4. **Pick the version** per [semver](https://semver.org/):
   - **Patch** (`v0.1.1`) — bug fixes, docs corrections
   - **Minor** (`v0.2.0`) — new commands, flags, or features
   - **Major** (`v1.0.0`) — breaking changes to CLI interface or job format

5. **Tag the release**:

   ```bash
   git tag vX.Y.Z
   git push origin vX.Y.Z
   ```

6. **Wait for the release workflow** to finish. It runs GoReleaser, which builds
   darwin binaries (amd64 + arm64), creates the GitHub release, and publishes the
   Homebrew formula to [`dsaiztc/homebrew-tap`](https://github.com/dsaiztc/homebrew-tap).

   ```bash
   gh run list --workflow=release.yml --limit 1
   gh run watch <run-id> --exit-status
   ```

7. **Edit the release notes** — GoReleaser generates a commit-based changelog, but
   replace it with hand-written notes following the format below:

   ```bash
   gh release edit vX.Y.Z --notes "..."
   ```

## Release notes format

Follow [Keep a Changelog](https://keepachangelog.com/) conventions. Start with a
one-line summary of the release, then use `###` sections as applicable:

- **Added** — new features
- **Changed** — changes to existing functionality
- **Deprecated** — features marked for removal
- **Removed** — features that were removed
- **Fixed** — bug fixes
- **Security** — vulnerability fixes

Only include sections that apply. See the
[v0.1.0 release](https://github.com/dsaiztc/croni/releases/tag/v0.1.0) for an
example.

## Gotchas

- **`HOMEBREW_TAP_TOKEN`**: the release workflow requires this GitHub secret — a
  PAT with write access to the `dsaiztc/homebrew-tap` repo. If the Homebrew step
  fails, check the token hasn't expired.
- **GoReleaser changelog filter**: commits prefixed with `docs:`, `test:`, or
  `chore:` are excluded from the auto-generated changelog. This doesn't matter
  if you replace the notes (step 7), but be aware if you skip that step.
- **darwin-only builds**: `.goreleaser.yaml` only targets macOS (amd64/arm64).
  croni depends on launchd, so Linux/Windows builds are intentionally excluded.
