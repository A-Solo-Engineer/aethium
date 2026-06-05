# Aethium — Contributing

Governance, PR requirements, benchmarks, branching, conduct, and license/CLA policy. Consistent with `VISION.md` (AGPL-3.0) and `BUILD_SYSTEM.md` (resource budgets as exit criteria).

---

## Governance Model

Aethium is a solo-maintained project. All architectural, API, and policy decisions are made exclusively by the sole maintainer. There is no maintainer council, no elected members, and no voting process.

### Decision authority

| Change type | Authority |
|-------------|-----------|
| Patch (bugfix, docs, perf, no API change) | Sole maintainer — merge after CI green |
| Minor API (additive) | Sole maintainer — no RFC required; changelog entry required |
| Breaking API | Sole maintainer — RFC document required in `docs/rfcs/` before implementation |
| Resource budget exception | Sole maintainer — `docs/exceptions/<issue>.md` required before merge |
| License change | Not permitted under any circumstance in this repository |

### RFC process (breaking changes only)

1. Sole maintainer opens `docs/rfcs/NNNN-title.md` with status `Draft`.
2. RFC must include: motivation, Go API signatures, migration path, binary/size impact estimate, TinyGo compatibility notes.
3. RFC is self-approved by sole maintainer after a personal review period of at least **48 hours** — this enforced pause exists to prevent hasty breaking changes, not to simulate a committee.
4. Status transitions: `Draft` → `Accepted` (sole maintainer, after 48h) → implementation PR linked.
5. Rejected RFCs remain in the repo with a `Rejected` header and documented reason — prevents re-litigation of closed decisions.

### External contributions

Pull requests from external contributors are welcome but governed by the same checklist in the PR Requirements section. The sole maintainer has final merge authority on all PRs with no appeals process. Contributors who disagree with a decision may fork under AGPL-3.0.

### Succession

This project has no formal succession plan at Stage 1. If the sole maintainer becomes unavailable, the AGPL-3.0 license ensures the codebase remains forkable and the community may continue development independently.

---

## PR Requirements

Every PR must satisfy **all** applicable items (checklist in PR template):

### Required (all PRs)

- [ ] `go test ./...` passes on **Go 1.22.7+** (native packages)
- [ ] `tinygo build` passes for packages tagged `//go:build tinygo` (CI)
- [ ] `go vet ./...` clean
- [ ] **AGPL-3.0** license header on new files with module line, for example:

```
// Copyright YEAR Name
// SPDX-License-Identifier: AGPL-3.0-only
// Module: github.com/A-Solo-Engineer/aethium
```
- [ ] CHANGELOG.md entry under `Unreleased` (user-facing) or `Internal` (refactor)
- [ ] No new core dependencies without RFC reference number

### API changes

- [ ] RFC `Accepted` linked
- [ ] `docs/` updated if behavior, flags, or lifecycle change
- [ ] Example app updated if public API surface changes

### Rendering pipeline or state engine (`canvas`, `scene`, `reactive`, `runtime`, `platform`)

- [ ] `go test -bench=. -benchmem` before/after attached in PR description
- [ ] No regression > **5%** in `BenchmarkTick` or `BenchmarkSignalNotify` without justification
- [ ] If PR touches binary size: `aethium build` size report (desktop + wasm gzip) in PR description

### Budget-related

- [ ] Meets Resource Efficiency Budget in `BUILD_SYSTEM.md` **or** links approved `docs/exceptions/<issue>.md`

---

## Benchmark Contracts

PRs modifying **rendering pipeline** or **state engine** must include:

```text
go test -bench='Benchmark(Tick|SignalNotify|DrawList)' -benchmem -count=5 ./...
```

Report format in PR:

| Benchmark | Old ns/op | New ns/op | Old B/op | New B/op |
|-----------|-----------|-----------|----------|----------|
| ... | | | | |

CI will run benches on `ubuntu-latest` and fail if regression exceeds threshold without `benchmark-approved` label (maintainer only).

Note: `BenchmarkTick` and `BenchmarkSignalNotify` are exit-criteria benchmarks for Stage 2 and must be added as tests in `runtime/runtime_test.go` and `reactive/reactive_test.go` before the `v0.1` tag; until then, PRs must include manual timing results in the PR description.

---

## Branch Strategy

**Trunk-based development** on `main`.

| Branch | Purpose | Lifetime |
|--------|---------|----------|
| `main` | Always releasable; protected | Permanent |
| `feat/<name>` | Features | Delete after merge |
| `fix/<issue>-<name>` | Bugfixes | Delete after merge |
| `release/vX.Y` | Cut only for patch releases | Tag then merge back |

**No long-lived Gitflow `develop` branch.**

### Release cadence

| Version | Cadence | Content |
|---------|---------|---------|
| `v0.x` | Bi-weekly while pre-1.0 | Breaking allowed with RFC |
| `v1.x+` | Monthly minor, weekly patch as needed | SemVer; breaking only major |

Tags signed by maintainer; GitHub Releases with `aethium` binaries for win/mac/linux optional after v0.3.

---

## Code of Conduct

Aethium adopts the **[Contributor Covenant v2.1](https://www.contributor-covenant.org/version/2/1/code_of_conduct/)**.

### Enforcement

| Step | Action |
|------|--------|
| Report | Email `conduct@aethium.dev` (or GitHub private security advisory for safety issues) |
| Triage | Sole maintainer responds within **72 hours** |
| Outcomes | Warning, PR block, temporary ban, permanent ban per Covenant enforcement guidelines |
| Appeals | Not available — the sole maintainer is the final authority. Disagreeing parties may fork under AGPL-3.0. |

The sole maintainer is bound by the same Contributor Covenant standards as all contributors. Harassment, discrimination, or sustained bad-faith argumentation from any party, including the maintainer, are violations of the project's stated values.

---

## License Consistency

### Project license

All repository code is **GNU Affero General Public License v3.0 (AGPL-3.0)** per `VISION.md`.

Contributions must be licensed under AGPL-3.0. You retain copyright; you grant the project perpetual rights to distribute under AGPL-3.0.

### CLA policy

**Contributor License Agreement (CLA): Not required.**

**Rationale:**

- AGPL-3.0 already requires derivative works and network-deployed modified versions to share source; inbound licensing is uniform via license header on each commit.
- CLAs that grant patent or relicensing rights to a corporate entity conflict with Aethium’s **anti-capture** goal (see cloud-capture discussion in `VISION.md`).
- Developers contribute via **DCO (Developer Certificate of Origin)** sign-off instead:

```text
Signed-off-by: Name <email@example.com>
```

Use `git commit -s`. CI rejects commits to `main` without `Signed-off-by` on non-trivial changes.

### What contributors must not submit

- Code incompatible with AGPL-3.0 (e.g. CC-BY-NC, proprietary licensed snippets)
- Dependencies with licenses that forbid AGPL combination (default deny; RFC for exceptions)

### Network deployment reminder

If you operate Aethium-as-a-service with modified core, AGPL-3.0 **requires** offering corresponding source to users interacting with your service—see `VISION.md`.

---

## Getting started (documentation only)

Stage 1 is documentation-only. After **`APPROVED — proceed to Stage 2`**, read forthcoming `docs/IMPLEMENTATION_PLAN.md`. After second approval **`APPROVED — proceed to implementation`**, clone and run tests.

**Do not open implementation PRs before Stage 2 plan approval.**
