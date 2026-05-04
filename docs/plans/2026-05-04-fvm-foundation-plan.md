# fvm Foundation Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Build the first real foundation of `fvm`, a Go-based Foundry version manager with hybrid shims + optional shell helpers, plus a release/install pipeline suitable for curl-based installation.

**Architecture:** `fvm` is a Go CLI organized around three separable layers: command/UI surface, version-resolution/install engine, and packaging/distribution. Runtime resolution is shim-first and shell-optional. Persistent state lives under `~/.fvm`, with project-local version selection via `.fvm-version` and user-global default via `~/.fvm/version`.

**Tech Stack:** Go 1.26.x, stdlib-first CLI implementation, YAML config, shell scripts for installer/bootstrap, GitHub Actions for CI and release automation.

---

## Locked Product Contract

### Switching model
- Hybrid model.
- Shims on PATH are the primary execution path.
- Optional shell integration exists for convenience.

### Command semantics
- `fvm local <version>` writes `.fvm-version` in the current project.
- `fvm global <version>` writes `~/.fvm/version`.
- `fvm use <version>` is ephemeral shell/helper activation, not persistent config.

### Version resolution order
1. explicit shell override
2. nearest `.fvm-version`
3. global `~/.fvm/version`
4. error if no version can be resolved

### Filesystem contract
- `~/.fvm/config.yaml`
- `~/.fvm/version`
- `~/.fvm/shims/`
- `~/.fvm/versions/<version>/`
- `~/.fvm/downloads/`
- `~/.fvm/tmp/`
- `~/.fvm/registry/`
- `~/.fvm/log/`
- project-local `./.fvm-version`

### Per-version layout
- `~/.fvm/versions/<version>/foundry/`
- `~/.fvm/versions/<version>/bin/`
- `~/.fvm/versions/<version>/manifest.json`
- `~/.fvm/versions/<version>/.complete`

### Minimum command set for this phase
- `fvm current`
- `fvm which`
- `fvm local <version>`
- `fvm global <version>`
- `fvm where <version>`
- `fvm install <version>`
- `fvm list`
- `fvm list-remote`
- `fvm shim regenerate`
- `fvm doctor`
- `fvm init <shell>`

---

## Workstream Split

### Stream A — CLI / UX surface
Responsible for:
- command parsing and help text
- implementing user-facing commands backed by interfaces
- `current`, `which`, `where`, `local`, `global`, `doctor`, `init`
- source-aware output such as `13.346 (.fvm-version)`

### Stream B — Version engine and filesystem model
Responsible for:
- config loading and path resolution
- `.fvm-version` discovery
- global version persistence
- version resolution logic
- install root layout and manifest model
- shim generation and install records
- stubbed remote-version and install plumbing where real upstream details are still unknown

### Stream C — Release / distribution pipeline
Responsible for:
- release artifact naming contract
- real installer script skeleton with OS/arch detection
- GitHub Actions for build/test/release artifacts/checksums
- documentation for curl install flow and PATH/init instructions

---

## Task 1: Build foundational domain and path contracts

**Objective:** Create the core types and path/config helpers that define the locked `fvm` filesystem contract.

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`
- Create: `internal/paths/paths.go`
- Create: `internal/paths/paths_test.go`
- Modify: `go.mod`

**Step 1: Write failing tests**
- Verify default home root resolves to `~/.fvm`.
- Verify config path is `~/.fvm/config.yaml`.
- Verify global version path is `~/.fvm/version`.
- Verify project version filename is `.fvm-version`.

**Step 2: Run test to verify failure**
Run: `go test ./...`
Expected: FAIL because packages/files do not exist yet.

**Step 3: Write minimal implementation**
- Add path helpers for root/config/shims/versions/downloads/tmp/registry/log.
- Add YAML config model with minimal fields:
  - `install_root`
  - `shim_names`
  - `release_channel`
  - `cache_ttl`
  - `auto_regenerate_shims`
  - `prefer_official_builds`

**Step 4: Run test to verify pass**
Run: `go test ./...`
Expected: PASS for new package tests.

**Step 5: Commit**
Commit message: `feat: add config and path contracts`

---

## Task 2: Implement version resolution contract

**Objective:** Encode the locked resolution order and make it testable independently of CLI concerns.

**Files:**
- Create: `internal/resolve/resolve.go`
- Create: `internal/resolve/resolve_test.go`
- Modify: `internal/config/config.go` if required
- Modify: `internal/paths/paths.go` if required

**Step 1: Write failing tests**
- shell override beats local
- local beats global
- global is used when local absent
- error is returned when nothing resolves
- current result includes source enum/string

**Step 2: Run test to verify failure**
Run: `go test ./...`
Expected: FAIL because resolver is not implemented.

**Step 3: Write minimal implementation**
- Add resolver inputs for cwd, shell override, local version file lookup, global version file lookup.
- Return a structured result with version, source, and resolved path metadata.

**Step 4: Run test to verify pass**
Run: `go test ./...`
Expected: PASS.

**Step 5: Commit**
Commit message: `feat: add version resolution engine`

---

## Task 3: Implement local/global version persistence

**Objective:** Support writing and reading durable local/global version selections cleanly.

**Files:**
- Create: `internal/state/version_files.go`
- Create: `internal/state/version_files_test.go`
- Modify: `internal/resolve/resolve.go` if required

**Step 1: Write failing tests**
- `WriteLocalVersion` writes `.fvm-version`
- `WriteGlobalVersion` writes `~/.fvm/version`
- values are trimmed and newline-normalized
- nearest ancestor lookup works for nested directories

**Step 2: Run test to verify failure**
Run: `go test ./...`
Expected: FAIL.

**Step 3: Write minimal implementation**
- Add safe read/write helpers.
- Add nearest-ancestor search.

**Step 4: Run test to verify pass**
Run: `go test ./...`
Expected: PASS.

**Step 5: Commit**
Commit message: `feat: add local and global version state handling`

---

## Task 4: Build manifest/install/shim domain model

**Objective:** Create the per-version manifest format and shim-generation primitives without requiring full upstream install integration yet.

**Files:**
- Create: `internal/install/manifest.go`
- Create: `internal/install/manifest_test.go`
- Create: `internal/shim/shim.go`
- Create: `internal/shim/shim_test.go`
- Create: `internal/install/layout.go`
- Create: `internal/install/layout_test.go`

**Step 1: Write failing tests**
- per-version paths resolve correctly
- manifest marshals/unmarshals with required fields
- shim generation produces executable wrapper content for configured shim names
- incomplete installs are ignored unless explicitly requested by future callers

**Step 2: Run test to verify failure**
Run: `go test ./...`
Expected: FAIL.

**Step 3: Write minimal implementation**
- Define manifest fields:
  - version
  - source_url
  - installed_at
  - platform
  - arch
  - checksum
  - executable_path
- Define `.complete` semantics.
- Generate POSIX shims for `foundry` and `foundryvtt` by default.

**Step 4: Run test to verify pass**
Run: `go test ./...`
Expected: PASS.

**Step 5: Commit**
Commit message: `feat: add manifest and shim primitives`

---

## Task 5: Build CLI command surface around interfaces

**Objective:** Replace the current tiny demo CLI with the locked command set and clean output behavior.

**Files:**
- Modify: `cmd/fvm/main.go`
- Create: `internal/cli/app.go`
- Create: `internal/cli/app_test.go`
- Create: `internal/cli/commands_current.go`
- Create: `internal/cli/commands_local.go`
- Create: `internal/cli/commands_global.go`
- Create: `internal/cli/commands_which.go`
- Create: `internal/cli/commands_where.go`
- Create: `internal/cli/commands_doctor.go`
- Create: `internal/cli/commands_init.go`
- Create: `internal/cli/commands_list.go`
- Create: `internal/cli/commands_install.go`
- Create: `internal/cli/commands_shim.go`

**Step 1: Write failing tests**
- `fvm current` prints resolved version and source
- `fvm local 13.346` writes `.fvm-version`
- `fvm global 13.346` writes global version file
- `fvm which` prints executable path or useful error
- `fvm doctor` reports missing PATH/root issues clearly
- `fvm init zsh` and `fvm init bash` emit shell integration text

**Step 2: Run test to verify failure**
Run: `go test ./...`
Expected: FAIL.

**Step 3: Write minimal implementation**
- Keep stdlib-first parsing unless a library becomes clearly necessary.
- Use interfaces/services from lower layers instead of duplicating logic in command handlers.
- Preserve concise UX.

**Step 4: Run test to verify pass**
Run: `go test ./...`
Expected: PASS.

**Step 5: Commit**
Commit message: `feat: add core cli commands`

---

## Task 6: Add remote-list/install scaffolding

**Objective:** Add the install/list-remote/list plumbing and stable abstractions, even if upstream Foundry fetch details remain partially stubbed.

**Files:**
- Create: `internal/releases/releases.go`
- Create: `internal/releases/releases_test.go`
- Modify: `internal/install/` packages as needed
- Modify: `internal/cli/commands_install.go`
- Modify: `internal/cli/commands_list.go`

**Step 1: Write failing tests**
- installed versions can be listed from manifest roots
- remote versions provider interface can be mocked
- install flow creates temp dir, version dir, manifest, and `.complete` in correct order

**Step 2: Run test to verify failure**
Run: `go test ./...`
Expected: FAIL.

**Step 3: Write minimal implementation**
- Add provider interfaces and stub/default provider behavior.
- Add atomic install layout behavior suitable for later real downloads.

**Step 4: Run test to verify pass**
Run: `go test ./...`
Expected: PASS.

**Step 5: Commit**
Commit message: `feat: add install and release provider scaffolding`

---

## Task 7: Build installer and release automation

**Objective:** Replace the current installer stub with a real release-aware script and CI/release workflow skeleton.

**Files:**
- Modify: `scripts/install.sh`
- Modify: `.github/workflows/ci.yml`
- Create: `.github/workflows/release.yml`
- Create: `docs/install.md`
- Modify: `README.md`

**Step 1: Write failing validation expectation**
- installer script should support OS/arch detection and artifact URL construction
- release workflow should build darwin/linux/windows artifacts
- README should document PATH and shell init steps

**Step 2: Run validation to verify gaps**
Run: `shellcheck scripts/install.sh` if available, otherwise manual review plus script execution with dry-run env.
Expected: initial gaps from stub behavior.

**Step 3: Write minimal implementation**
- Detect platform/arch.
- Resolve release URL pattern.
- Download tarball/zip.
- Install binary into `${INSTALL_DIR:-$HOME/.local/bin}` or tool-specific location.
- Print post-install PATH/init guidance.
- Add release workflow with archives + checksums.

**Step 4: Run validation to verify pass**
Run:
- `bash -n scripts/install.sh`
- `go test ./...`
Expected: PASS.

**Step 5: Commit**
Commit message: `feat: add installer and release automation skeleton`

---

## Task 8: Final integration verification

**Objective:** Prove the three streams work together well enough to continue implementation safely.

**Files:**
- Modify: any touched files as needed from review feedback

**Step 1: Run full verification**
Run:
- `go fmt ./...`
- `go test ./...`
- `go build ./cmd/fvm`
- `./fvm current` (expect useful error or resolved version)
- `./fvm init zsh | head` or equivalent safe preview

**Step 2: Review contract coverage**
Checklist:
- [ ] hybrid model represented
- [ ] local/global/use semantics preserved
- [ ] resolution order implemented
- [ ] YAML config path locked
- [ ] global version stored in `~/.fvm/version`
- [ ] shims modeled under `~/.fvm/shims`
- [ ] installer/release skeleton no longer placeholder-only

**Step 3: Commit**
Commit message: `feat: complete fvm foundation phase`

---

## Parallel execution notes

Recommended stream grouping for immediate delegation:

- Stream A
  - Task 5
  - small CLI-facing portions of Task 8

- Stream B
  - Tasks 1, 2, 3, 4, 6

- Stream C
  - Task 7
  - docs portions of Task 8

Coordination rules:
- Stream B defines the interfaces and path contracts first.
- Stream A may proceed in parallel if it consumes interfaces conservatively and rebases once Stream B lands.
- Stream C should avoid guessing command names beyond the locked set above.
- If parallel edits collide, prefer Stream B contracts as authoritative.

## Verification standard

Before claiming success on any stream, run fresh evidence:
- `go fmt ./...`
- `go test ./...`
- `go build ./cmd/fvm`
- relevant script syntax validation for shell work

## Remember

- Keep the product contract locked.
- No overloading `use` with persistent writes.
- Keep shell integration optional.
- Preserve shim-first execution.
- Keep config minimal.
- Prefer boring, inspectable state over cleverness.
