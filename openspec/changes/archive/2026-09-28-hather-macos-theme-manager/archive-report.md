```yaml
schema: gentle-ai.archive-result/v1
status: pass
change: hather-macos-theme-manager
archived_path: openspec/changes/archive/2026-09-28-hather-macos-theme-manager
requirements: 14/14
scenarios: 37/37
tasks: 54/54
blockers: 0
verdict: pass_with_warnings
```

# Archive report

## Artifacts read

- `proposal.md`
- `specs/theme-manager/spec.md`
- `design.md`
- `tasks.md`
- `apply-progress.md`
- `verify-report.md`
- `sync-report.md`
- `openspec/config.yaml`

## Canonical sync

- Domain: `theme-manager`
- Canonical path: `openspec/specs/theme-manager/spec.md`
- Operation: created new canonical full-domain specification
- Added requirements: all 14 verified requirements
- Modified requirements: none
- Removed requirements: none
- Same-domain active-change warnings: none
- Destructive merge: none
- Source/canonical SHA-256: `b52d8332fe07309b15c4b3cee38572ffc75dbf0a1d93c12be31284d092d53b20`

## Completion and evidence

No unchecked implementation tasks remain. Final verification is PASS WITH WARNINGS at 14/14 requirements, 37/37 scenarios, and zero blockers. Exact-current local Homebrew acceptance used archive SHA-256 `541d001ab20b7c14f209da9d25bb92971cc604efe261cd517908bb2f8b57bd11`; vendored tests, temporary install, formula test, installed help/version/preview/apply, and cleanup passed.

The historical `apply-progress.md` overwrite incident was recovered exactly from immutable begin candidate tree `95dd8b9f38d9d33ca7b6487cbd36e3541ec618c4` before final evidence was appended.

## Remaining warnings

- Both Git repositories had no tracked `HEAD` when verified, so diff checks could not inspect untracked content or independently reconstruct historical slices.
- Live TCC denial remains fixture-tested only.
- No production tag, GitHub release, push, public tap publication, or public installation was performed before archive.
- Archive readiness does not authorize delivery by itself; delivery remains a separate user decision.

Action context was repo-local at `/Users/facundogayoso/projects/Hather`; all archive paths are within that workspace.
