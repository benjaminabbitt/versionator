---
inclusion: always
---

AI code makes ~2x more concurrency mistakes than human code (CodeRabbit 2025, 470 PRs). Never parallelize sequential awaits; mark load-bearing ordering with `// SYNC-REQUIRED: [reason]`.

---

AI code: correct happy path, dangerous elsewhere. LLMs hallucinate packages at 5.2-21.7% rates (arXiv:2406.10279) — verify imports (`npm info`/`pip show`/`cargo search`), API signatures, endpoints. Test integrity: AI "fixes" by deleting tests, removing assertions, mocking away behavior (mutation testing catches this). Red flags: unexplained deletions, catch-alls replacing specific handlers, removed validation, async/sync flips, unjustified deps. Reject immediately: security vulns in sensitive code, deleted tests, hallucinated deps, race conditions, missing error handling in critical paths.

---

# is-this-a-standard

# Is This a Standard?

Before writing infrastructure, ask:

> Is this a standard? Did we already wire it in?

## Why

LLMs default to local construction. They've seen `grpc.reflection.v1.ServerReflection` a thousand times, but write end-to-end over looking up.

## Common Categories AIs Reinvent

- gRPC Server Reflection (`grpc.reflection.v1.ServerReflection`)
- OpenTelemetry instrumentation (OTLP exporters)
- JWT validation (vetted library)
- OAuth 2.0 Device Authorization Grant (library)
- JSON Schema validation (metaschema)
- Kubernetes AdmissionReview (standard webhook shape)
- Conventional Commits parsing (existing parser)
- Healthchecks (`grpc.health.v1.Health`)

## What to Do

When AI proposes infrastructure:

1. Name the category
2. Ask if a standard exists
3. If yes, ask if project already wires it in
4. Use the standard

## Load-Bearing Comments

At registration site, name the invariant (see sibling `no-provenance-comments`):

```go
// reflection.Register: canonical "what does this server speak"
// mechanism. Do not add a parallel descriptor service; ask reflection.
reflection.Register(grpcServer)
```

Lives next to the call that triggers reinvention. No external doc dependency.

---

Mutation testing is the deterministic check on test quality; LLM tests look plausible while hiding tautological assertions, boundary gaps, and implementation coupling. After the TDD round-trip: run mutation testing, analyze survivors, add tests that kill them or accept the gap explicitly. Tools: cargo-mutants (Rust), pitest (Java), mutmut/cosmic-ray (Python), stryker (JS/TS), gomutation/go-mutesting (Go). Kill-rate targets: 80-90% pure functions and business logic, 60-70% framework glue, logging-only paths exempt. Coverage measures execution, mutation measures verification (80-90% coverage teams routinely see 30% kill rates). Never game the score with adjacent `assert!(true)`.

---

# no-provenance-comments

# No Provenance in Comments

Comments state invariants. Git records timeline.

## The Rule

No git-provenance metadata in code comments or codebase prose:

- Commit hashes (`see commit abc123`)
- Milestone tags (`wired in P0.2`)
- PR numbers (`added in #1234`)
- Release versions as historical markers (`introduced in 1.4.0`)
- Author/date stamps (`added by Alice 2024-01-15`)
- TODO sprint/iteration references

## Why

- Commit hashes break after rebase/squash
- Milestone tags become noise post-project
- Author/date stamps age into tombstones surviving author departure and line rewrites
- `git blame`/`git log` already record this; duplication guarantees divergence

## What to Write Instead

State the load-bearing invariant. If the comment survives a history rewrite, keep it. Else it's a timeline entry in the wrong file.

Before:
```go
// reflection.Register: wired in P0.2 (commit abc123). Canonical.
```

After:
```go
// reflection.Register: canonical "what does this server speak"
// mechanism. Do not add a parallel descriptor service.
```

## Where Provenance Lives

- `git blame`/`git log`
- Commit message body, PR description
- `CHANGELOG.md`/release notes
- ADRs
- Issue tracker

## Exceptions (Not Provenance)

- External stable IDs: RFC numbers, CVE IDs, language-spec versions
- Public-API `@since` annotations: version is part of API contract
- "We tried X and it failed": invariant is the rejected approach; dated artifact lives in rationale

---

# TDD for LLM-Written Code

LLMs non-deterministic; tests are deterministic gate.

## Rule
Demand TDD: tests first, implementation second. Non-negotiable.

## Workflow
1. Describe requirement
2. LLM writes tests against requirement (not implementation)
3. Review tests: requirement, edge cases, failure modes captured?
4. LLM implements to pass tests
5. Run tests

## Why
Test suite = contract. Without tests-first, model fills contract with its own output — no separate ground truth.

## Anti-Patterns
- Implementation-first, then tests → tests confirm bug, not requirement
- Vacuous assertions (`assert!(true)`)
- Implementation-coupled tests (`assert_eq!(hash("x"), 0x7a3f...)`) — brittle, blind to behavior
- Skipping step 3: test review is load-bearing

## What Tests Document
Document **problem**, not solution. Name requirement in test name; assert observable behavior.

## Verification
```
Show me the tests before the implementation. I will review the tests, approve them, then you write the implementation.
```

---

Never bake versions into AI context files; lockfiles are the source of truth (a static table ages out the moment the project moves). On session start read the lockfile and language-pin files; propose upgrades through the lockfile, never as one-offs in code. Verification prompt: "Your suggestion uses [pattern]. Project lockfile pins [package] at [version]. Confirm the pattern is supported at that version, or propose updating the lockfile."

---

# Warning Suppression
Do not suppress lint or compiler warnings without explicit user approval.  Warnings are signals of potential issues. Suppressing them without review risks hiding real problems.
Ensure that, when the user does approve a warning suppression, the comment includes the justification and details of the warning being suppressed. This creates a record for future reviewers to understand the context and reasoning behind the suppression.

---

# Mutation Testing

High coverage + low mutation kill rate = false confidence.

```bash
git worktree add --detach ../.mutants-worktree HEAD
cargo mutants -d ../.mutants-worktree --in-place --timeout 120 -f <file> -- --lib
git worktree remove ../.mutants-worktree --force
```

Worktree shares .git and copies only source; `--in-place` is safe because the worktree is disposable.

Kill-rate targets: pure utilities/validators 90%+; business logic/state machines 85%+; orchestration/coordinators 70%+; framework glue/adapters 50%+. Skip generated code (`*.pb.rs`, `src/proto/`), trivial delegation, framework boilerplate.

---

# TDD

Red-green-refactor is mandatory; verify the red test runs and fails for the right reason before implementing.

Integration/acceptance: build and run actual binaries, don't define hooks; tag slow tests; isolate and clean up after yourself.

Naming: `test_<action>_<condition>_<expected_result>` in the language's casing; readability over strict format. Order test files by complexity — usage-demonstrating examples at top, edge cases at bottom; tests are documentation.

---

# Test Coverage

Target 90%+ (unit + integration, full application). Acceptable gaps: main entrypoints (E2E-covered), generated code, impossible panic/fatal paths, default factory functions (exclude via tooling).

---

# Git

Branching: no branches or PRs unless explicitly asked; short descriptive names; one branch per task; branch off main/master only — push back and confirm anything else.

Commits: terse, describe code changes only, no meta-commentary; NEVER mention Claude, Anthropic, AI, or "Generated with".

Breaking changes: <1.0 and new major versions need no backwards compat — remove deprecated code immediately; post-1.0 minor/patch, discuss before implementing.

Pre-commit: lint, format, test before committing — never commit broken code; fix pre-commit errors automatically without asking. Bypass hooks (`--no-verify`) only for WIP on feature branches, with documented reasoning.

---

# communication

## Do

- **State limitations immediately**: "Cannot verify X without Y", "Has limitation Z", "Need clarification on A"
- **Admit uncertainty**: "I don't know" is valid; label verified vs inferred; read/look up before asserting or changing
- **Ask for clarification when**: requirements ambiguous, multiple approaches exist, trade-off input needed, context uncertain
- **Lead with key info**: most important point, supporting details, rationale
- **Cite sources**: API docs, best practices, performance/security claims
- **Test before complete**: TDD mandatory—verify tests pass

## Don't

- **No sycophancy/politeness**: no praise, enthusiasm, validation seeking, or excessive courtesy
- **No assumptions**: ask rather than guess; explicitly state educated guesses

## Presenting decisions

- **Use the question system for choices**: put decisions through the interactive question tool, not prose; narrative is for the reasoning that feeds a choice, not the choice
- **Standalone-complete across time**: the user context-switches and may not recall an earlier decision or an hour-ago dispatch — restate the situation/behavior/stakes a question needs from earlier; within-batch shared context is fine, elapsed time is the gap
- **Batch by context, not count**: large batches OK if each question is via the question system AND individually standalone-complete; collapse dependents (if X settles Y, don't ask Y)
- **Recommend, don't survey**: lead with your recommendation as the first marked option; options mutually exclusive, each with its trade-off — a real fork

---

# Workarounds and Problem Solving

**Root cause first.** Fix at source, never workaround without asking.

## When Encountering Failing Functionality

1. Find root cause — investigate actual source.
2. If simple problem needs complex fix, ask before proceeding.
3. Present options: proper fix (effort), workaround (trade-offs), test disable (why), alternatives.
4. Cost/benefit: tech debt, maintainability, time per option.
5. Document decision and reasoning.

## Workaround Comment = Unfiled Bug

Comment explaining WHY a workaround exists = defect diagnosis. Comments aren't reports; they become tombstones others read as settled. Cause survives unaddressed.

**Red flags:** arbitrary limits with justifying comments; retry/sleep/poll around deterministic things; "without this, X breaks"; thresholds tuned to silence gates; fallbacks masking real failures.

## Three Steps to Land a Workaround

1. Escalate: name the bug, locate it.
2. File root cause as task (with diagnosis). Unfiled = agreed to forget.
3. Comment states invariant, not history. No invariant = scar, not fix.

## Tuned-Silent Gates Measure Nothing

Never tune thresholds/timeouts/coverage to silence gates. Silenced gates = false confidence. Fix gates deliberately.

## "Works in CI" ≠ "Works"

CI ≠ local hides bugs (clean checkouts, no TTY, etc). Chase environmental differences; don't paper over them.

---

# Pushback

## Situations

- Skip tests
- Add features without tests
- Ignore type hints
- Work around linting

## Response

1. Why problematic
2. Consequences
3. Correct approach
4. Defer if insisted; note debt

## Feature Questions

- Acceptance criteria?
- Performance requirements?
- Error cases?
- Security?
- Logging?
- Dependencies?
- Testing?
- Error messages?

## Component Questions

- Dependencies?
- Dependency interface/protocol?
- Logging context?
- Error conditions & messages?

---

# Code Quality

Before writing: search the codebase for existing implementations; reuse or extend rather than duplicate or recreate.

Size: <500 lines per file (exceed only with very high coupling/cohesion); small single-purpose functions; optimize for reading, not performance; separate interfaces from implementations by file.

Naming: the interface is the thing — UserService is the Protocol, never IWhatever; implementations are named for how they implement: DefaultUserService (single), HttpUserService/CachedUserService (multiple).

Clean up: kill background processes when done; remove unused code, files, imports, variables; no dead code.

Comments explain why only. No change-tracking comments, no revision history in code, no commented-out code — git has it.

---

# ltk

**llm-tool-killer (ltk)** — pre-tool hook that inspects and redirects shell commands per `.ltk/config.yaml` rules.

## How it works

Parses real command (resolving variables, unwrapping wrappers) → matches against project rules → first matching `deny` returns `message`/`suggest`:

    go test ./...   →   blocked: "Run tests through the task runner."
                    →   retry: `just test`

## How to use

- Treat redirects as guidance; read suggestion and retry as specified
- Prefer project task runner (`just <target>`) over invoking tools directly  
- **Agents do not cut releases** — ltk blocks `git tag`/release commands; prepare version bump/PR for human or CI

## What it is not

Cooperative redirect, not a sandbox. Explicit workarounds are possible; for strict boundaries use a container.

---

# Deferral Tracking

Deferred work lives in taskloom, not in conversation memory or plan prose.

## Search before you create

Before `task_add`, search existing log using `task_list` (term filter) or `tag_query` on distinctive noun: file, symbol, command, subsystem.

If entry covers the work, UPDATE it (`task_edit`, append dated note). Create only when nothing matches.

**Critical:** Duplicates generate contradictory fixes; whoever picks up later cannot tell which is current.

Check for existing entry even when confident finding is new.

## Agent deferrals

Record deferred work with `task_add` before moving on (descoped, mid-implementation, postponed):

- Concrete revive condition ("after X merges", "when CI green") → "Deferred" status with condition as trigger
- No condition → "To Do"

Don't close work spawning deferrals until each is recorded.

## User deferrals

Offer brief taskloom entry when user defers. Create on confirmation. Don't push if declined.

## Durability

Chat/plan deferrals disappear when session ends. Taskloom is durable record; summaries reference by harp ID.

---

# reprise

This project runs **reprise**, duplicate detection built for
LLM-generated code (a *reprise* is a theme that returns in altered
form — a near-duplicate). LLM assistants systematically
reimplement existing helpers instead of calling them, and the
copies then drift apart; reprise catches clone Types 1-3 plus the
reimplemented-helper slice of Type-4, within this codebase. It is
REPORT-ONLY: it never edits code — responding to findings is your
job.

## The two commands

- `reprise scan` — full-repo ranked report of duplicate groups.
- `reprise check` — PR/commit mode: findings involving units
  changed since a git ref fail when they are new or worsened.
  The base ref defaults to the merge-base with the default
  branch; override with `--base <ref>` or `[baseline].ref` in
  `reprise.toml`. Its flagship finding is `inconsistent-update`:
  a change edited ONE copy of a known duplicate group but not the
  others. Exit codes: 0 clean, 1 findings at/above `--fail-on`,
  2 usage/runtime error.

There is no `baseline` subcommand: the baseline is a pinned git
ref, not a stored file. Adopt reprise on a legacy codebase by
pinning `[baseline].ref` to the current commit, so only NEW drift
fails while existing duplication is grandfathered.

## How to respond to findings

- **Before writing a helper**, search for an existing one and
  call it — that is the failure mode reprise exists to catch.
- **`inconsistent-update`**: the fix you just made belongs to
  every copy in the group. Prefer extracting the shared helper
  and calling it everywhere; at minimum, apply the change to all
  copies.
- **A new duplicate group**: replace your new implementation with
  a call to the existing unit (or extract one shared helper).
- **Intentional parallelism**: mark deliberate duplication at the
  source — `reprise:accept-drift` waives the drift gate while
  keeping the unit tracked; `reprise:ignore` drops the unit
  entirely. Never suppress a finding without the user's say-so.

## Pre-commit gate

reprise runs as a **lefthook pre-commit hook** (`reprise check`
in lefthook.yml). A failing hook means respond as above and
retry the commit — do not bypass with `--no-verify`.

---

# serena: address code by symbol, not by file and line

Serena puts a language server behind MCP tools that name SYMBOLS.
Prefer them over Read/Grep/Edit whenever the question or the change
is about a symbol.

## Reach for these first

- `get_symbols_overview` — what is in this file? The first call on
  an unfamiliar one.
- `find_symbol` — locate a definition by name path (`Type/method`).
  `include_body: true` reads ONE symbol instead of a whole file.
- `find_referencing_symbols` — every caller of a symbol. That is the
  blast-radius question, and grep answers it with false positives
  and misses aliased or qualified uses.
- `replace_symbol_body`, `insert_before_symbol`,
  `insert_after_symbol`, `rename_symbol`, `safe_delete_symbol` —
  edit by identity. `safe_delete_symbol` refuses while references
  remain, which a string edit cannot check.

## Why this is stated here

Serena's own registration injects an INDIRECTION — "call
`initial_instructions` before starting a coding task". An agent that
skips that call never learns the tools exist, and skipping it is the
normal outcome rather than the exception. This says it inline so the
guidance does not depend on a tool call nobody makes.

## Where they do not help

Non-code files (YAML, Markdown, JSON), whole-file reads you actually
need, and languages with no server configured. Read and Edit stay
correct there, as does a line-oriented edit INSIDE a symbol body
once you have located it.

## Delegating

Name the symbol tools in sub-agent briefs. An implementer told to
"read before you write" reaches for Read unless told otherwise.

## NOT for an agent working in a WORKTREE

The server resolves its project ONCE, at start-up, from the cwd of
the session that launched it. It is a long-lived process, and an
in-process sub-agent inherits the parent's connection rather than
getting its own. So every symbol call an agent makes is answered
against the COORDINATOR's checkout, whatever directory that agent
believes it is working in.

For a worktree-isolated agent this is wrong in both directions, and
silently:

- EDITS land in the coordinator's checkout. The agent's own worktree
  is untouched, so the change is both missing where it belongs and
  present where nobody asked for it.
- READS answer about the wrong tree. A symbol that exists only in
  the worktree is reported as not found; one deleted there is
  reported as present.

The failure has no error. A delete can report OK while writing
nothing the agent can see, so an agent that trusts the result
reports success having changed nothing.

**A brief that assigns a worktree must therefore tell the agent to
use Read/Edit and plain search, not serena.** Symbol tools are for
an agent working in the same checkout the server was started in.

## Containers

An agent bound to `runtime: container` cannot see a host-installed
serena. Either install serena in the image, or keep serena on
host-runtime agents.

---

# Closing a turn: fix what you can, file what you cannot

Before a turn ends, every issue it surfaced has to be DISPOSED OF. There
are exactly two honest dispositions, and "mentioned it in the reply" is
neither.

## Fix the easy ones

If a finding is root-caused and the fix is bounded, FIX IT. Filing a task
for something you already understand and could correct in the same turn
converts a solved problem into work someone pays to rediscover: they must
re-read the code, rebuild the reproduction, and re-derive the cause you
already had in hand.

A filed task looks like progress. It is not progress; it is a promise.
Prefer the fix, and where the fix is larger than the turn, dispatch it
rather than defer it.

## File the hard ones, with tags

File when the work genuinely cannot happen now: it needs a HUMAN DECISION
(name the fork and the options), it lives in another repository or
release, or it is materially larger than the current scope. Those are real
reasons. "I noticed several things" is not.

Tag what you file, because an untagged task is unfindable: what kind of
work it is, what it touches, its rough effort, and whether it is blocked
on a person. A task nobody can filter for is a task nobody reads.

## Leave the status TRUE

The task log and the plans are the shared picture of where things stand,
and a stale picture is worse than none because it is confidently wrong.
Before the turn closes:

- Close what the turn actually finished, stating what was asked for and
  what was done, so a reader can judge rather than take your word.
- Where a change satisfies a task only in PART, cut the task down to what
  REMAINS. A task carrying its own completed half is indistinguishable
  from work never started.
- Update the plan files the turn moved, including where reality diverged
  from the plan. The divergence is the most valuable thing in them.
- Check for tasks the turn quietly obsoleted, and for duplicates you may
  have just created by filing before reading.

## Report what was FIXED

Lead with what is now true, not with what was noticed. A list of filed
tasks is not a status report: it is a list of things still broken, and
reading it as accomplishment is how a backlog grows while the code stands
still.

---

# Decision Gating

Surface structural/interface decisions for sign-off before acting. The gate is a visible checkpoint — the user may approve or reject; what must not happen is the decision being made silently, buried in implementation.

## Gate these
- **Structural consolidation/splitting** — merging/splitting independent launchers, packages, modules; a shared abstraction across independent units. Changes *topology*, not just duplication.
- **Interface/seam changes** — public interface, trait, protocol, or extension seam. Ripples to every implementation; gate at least as hard as topology.
- **Operator-control trades** — trading the operator's control for author-side convenience (e.g. one polymorphic launcher vs discrete per-component launchers with full startup control).

## Don't gate
- DRY *within* a single unit (module, shared library).
- Local refactors that move no boundary and change no contract.

## Surface it
- **In plans:** a dedicated "Decisions for sign-off" section, each choice with 2–3 options. Never fold a topology/interface/control decision into a step bullet.
- **Mid-implementation:** a decision the approved plan didn't call out → stop and ask; don't pick and move on.
- **Default:** preserve existing separation and interface/seam contracts unless asked to change them. When unsure, treat it as a decision.

---

# Elegant Redo: Informed Reimplementation

Discard current impl and rebuild using lessons from first attempt.

## Philosophy
First impl teaches the problem. Second solves it well. Knowledge from edge cases, hidden requirements, and dead ends is the valuable output—not the code.

## Process
1. **Inventory knowledge** - Before deleting, document:
   - Hidden requirements, surprising edge cases
   - Architectural constraints revealed
   - Dependencies/interfaces to preserve
   - What worked vs what was over-engineered

2. **Identify elegant core** - With full problem knowledge:
   - Simplest abstraction covering all cases?
   - Natural data structures?
   - Where did first attempt fight the language/framework?
   - What's deletable vs essential?

3. **Scrap and rebuild** - Start clean:
   - No copy-paste from old impl
   - Let structure emerge from problem
   - Less code, fewer abstractions, simpler interfaces

4. **Validate** - New solution must:
   - Handle all discovered edge cases
   - Pass existing/improved tests
   - Be demonstrably simpler

## When to Apply
- Impl works but feels forced/overcomplicated
- Real problem shape doesn't match code
- Accumulated patches obscure design intent
- Developer explicitly requests fresh take

## Rules
- No sunk cost preservation—judge on current merit
- Simpler is better, but not at correctness cost
- Goal: minimum structure handling full problem naturally
- Keep tests (requirements); rewrite implementation

---

# Grill Me: Change Comprehension Gate

Quiz developer on their changes before PR creation to verify understanding.

## Process
1. Analyze diff (staged + unstaged)
2. Identify key decisions, trade-offs, edge cases, non-obvious implications
3. Ask pointed questions one at a time testing genuine understanding
4. Evaluate answers for real comprehension
5. Gate PR on sufficient understanding

## Question Categories
- **Intent**: What problem solved? Why this approach?
- **Impact**: What else affected? What breaks if this fails?
- **Edge cases**: Null/empty/concurrent behavior?
- **Trade-offs**: What sacrificed? Technical debt introduced?
- **Rollback**: How to revert safely? Blast radius?

## Grading
- **Pass**: Explains intent, impact, trade-offs clearly → proceed with PR
- **Partial**: Gaps exist → point out gaps, re-quiz weak areas
- **Fail**: Can't explain core decisions → no PR, suggest review areas

## Rules
- Rigorous but fair; test understanding, not memorization
- Scale difficulty to change complexity
- Explain wrong answers before continuing
- Never create PR until developer passes

---

# Prove It: Behavioural Diff Demonstration

Show concrete behaviour difference between main branch and current branch.

## Process
1. Identify behavioural claims - what should differ?
2. Design demonstrations exercising changed behaviour
3. Show before (main branch behaviour)
4. Show after (current branch behaviour)
5. Present side-by-side diff of inputs, outputs, effects

## Demonstration Methods
- Test output comparison across branches
- CLI invocation: same command, different output
- Code walkthrough: trace input through both paths, show divergence
- API calls: same request, different responses
- Error scenarios: differing failure modes

## Format

Per behavioural change:
```
## [Behaviour description]
### Main branch
Input: ... Output/Behaviour: ...
### This branch
Input: ... Output/Behaviour: ...
### What changed
[Concise diff explanation + why it matters]
```

## Rules
- Cover all intended changes, not just happy path
- Include edge cases and error conditions
- For pure refactors, demonstrate preserved behaviour
- Show evidence, don't just claim
- If local demo impossible, describe what would be tested and how

---

Go: testify/assert; fakes over mocks; no init() (explicit initialization); no package-level vars (DI); slog structured logging.

---

# Golang Dev

## Tools
- Acceptance: godog (Gherkin)
- Lint: golangci-lint, gofmt/goimports
- Logging: zap

## Test Layout
- Unit: `*_test.go` (co-located)
- Integration: `tests/integration/` or build tags
- Acceptance: `tests/acceptance/features/*.feature`
- Testify suites for shared setup; gomock via just target

## Constants
```go
// logmsg/messages.go
const UserCreated = "user_created"
logger.Info(logmsg.UserCreated, zap.String("username", username))

// errmsg/messages.go
const DivideByZero = "cannot divide by zero"
return 0, errors.New(errmsg.DivideByZero)
```

## IoC
```go
func NewUserService(repo UserRepository, logger *zap.Logger) *UserService {
    return &UserService{repo: repo, logger: logger}
}
func NewUserServiceDefault(db *Database) *UserService {  // nolint:unused
    return NewUserService(NewSQLUserRepository(db), zap.NewProduction())
}
```

---

# Go Testing

Name test functions TestFunctionName_Scenario_ExpectedResult; use descriptive subtest names.

---

# just: Command Runner

Language-agnostic task runner. Define tasks (`just test`, `just lint`, `just build`) in a `justfile`.

## TOP: standard repo-root variable

Every justfile defines `TOP`:

```just
TOP := `git rev-parse --show-toplevel`
```

All paths relative to `TOP`. Non-negotiable. Hard-coded relative paths or `{{justfile_directory()}}` break when invoked from subdirs or composed by parent justfiles.

```just
TOP := `git rev-parse --show-toplevel`

build:
    cargo build --manifest-path {{TOP}}/Cargo.toml --release
```

## Local justfiles, composed at root

Place justfile next to code it manages. Compose via `mod`:

```just
# /justfile (root)
TOP := `git rev-parse --show-toplevel`

mod web   "{{TOP}}/web/justfile"
mod api   "{{TOP}}/api/justfile"
```

Each submodule defines own `TOP`, owns own recipes. Root: `just web build`. Inside `web/`: `just build`. DO NOT use monolithic root justfile.

## Recipe shape

- Used 3+ times: lift to target.
- Top of file: short comment (purpose, prerequisites, side effects).
- Args via `+ARGS` (preserved through delegation).

## Cross-platform

Prefer `[unix]` / `[windows]` attributes over parallel platform justfiles. Reserve parallel files (`platform_justfile` import) for differing recipe shapes.

```just
[unix]
clean:
    rm -rf {{TOP}}/target

[windows]
clean:
    Remove-Item -Recurse -Force {{TOP}}\target
```

## Anti-Patterns

- Hard-coded relative paths (`./src/...`): break under composition.
- `{{justfile_directory()}}` as `TOP` stand-in: scoped to local file, not composing parent.
- Monolithic root justfile.

For container-delegated recipes, see `just-container-overlay` fragment.

---

# just-container-overlay

# just: Container Overlay Pattern

Host justfile delegates to container; inside container, different justfile mounted over host's runs actual command. Same `just build` works both contexts. No duplicate target names, no `container-build` vs `build` split.

## Setup

```just
# Host /justfile
TOP := `git rev-parse --show-toplevel`

_run +ARGS:
    docker run --rm \
      -v {{TOP}}:/workspace \
      -v {{TOP}}/justfile.container:/workspace/justfile:ro \
      -w /workspace build-env:latest just {{ARGS}}

build: (_run "build")
test:  (_run "test")
```

```just
# /justfile.container
TOP := `git rev-parse --show-toplevel`

build:
    cargo build --manifest-path {{TOP}}/Cargo.toml --release

test:
    cargo test --manifest-path {{TOP}}/Cargo.toml
```

## Why it works

Load-bearing: `-v {{TOP}}/justfile.container:/workspace/justfile:ro`. Bind mount obscures host `justfile` with `justfile.container` inside container. Inside: `just build` → container recipe (cargo). On host: `just build` → delegation spawning container.

## Separation

- Host file: orchestrates container, never builds.
- Container file: builds, never knows about container.
- Build changes → container file only.
- Orchestration changes → host file only.

## Real-world

Angzarr uses this across coordinators (aggregate, saga, projector, process-manager, stream, log, grpc-gateway). Each has local `justfile` + `justfile.container`. Root composes via `mod`; `just aggregate build` from root and `just build` from coordinator dir or container all produce same result.

## Anti-Patterns

- `_run` running on host instead of container: defeats overlay.
- Container justfile re-invoking `docker run`: accidental docker-in-docker.
- Mounting container justfile read-write instead of `:ro`: container build edits host file.
- Arg-escaping ceremony: just's `+ARGS` preserves multi-word args through delegation without `$$`-escaping. Use it.

---

# Cross-Backend BDD

One behavioral feature set per capability, run unchanged against every interchangeable backend (fs/S3/GCS, Kafka/NATS/AMQP, Postgres/SQLite). The backend is a harness dimension: never named or assumed in feature/step text, never asserted via backend mechanics (HTTP status, object keys, table rows). Litmus: if wording changes when the backend swaps, it describes how, not what — abstract to behavior.

The harness parameterizes over backends; the active one is wired in Given setup via config/DI, steps never know which runs. Use testcontainers/emulators, not mocks — mocking the backend tests the mock, not substitutability. Green across all backends = Liskov proof; per-backend feature files drift and are banned. Unit-level analogue: Behavioral Interface Test (BIT) — one suite against every implementation of an interface.

---

Dev Containers: `.devcontainer` config for consistent dev env w/ tools, deps, system reqs. Reproducible builds.

---

# Documentation Guidelines

Ask first before creating any *.md file, planning/strategy document, tracking file, or meta-documentation (exception: README updates when adding features). No progress notes in code ("refactored X to Y", changelog-style comments) and no change history in files ("Updated on...", "Previously this was..."); history belongs in version control and commit messages.

---

# String Handling

Never branch on string approximations of messages (startswith/contains/substring) unless explicitly instructed — use typed errors or error codes (`errors.Is(err, ErrConnectionRefused)`, not `strings.Contains(err.Error(), ...)`). Define all error messages as constants/sentinel errors and reuse them in test assertions — no magic strings.

---

# Do not build bindings you cannot check

Naming one thing from another creates a BINDING: a coupling that has to be
maintained for as long as both sides exist. Some bindings earn that. Most do
not, and the ones that do the most damage are the ones nothing enforces.

The decisive question is not "is this true?" — you would not write it otherwise.
It is: **when this stops being true, what catches it?**

    CHECKED    a symbol reference in code; a generated table; an asserted count.
               Breaks loudly, at the moment it breaks, in front of whoever
               broke it.

    UNCHECKED  prose, comments, docs, READMEs, config annotations. Nothing
               compiles them and no test reads them. When they go false they
               do not go red — they quietly start lying, and they keep their
               authority while doing it.

An unchecked binding needs a very good reason. The default is not to create one.

## Prefer no binding at all

Most of the time the coupling is unnecessary, because the fact is DERIVABLE.
Say what the relationship IS, not what the contents currently ARE:

- "the skill(s) it carries" — not their names
- "its members", "the formats it covers", "each leg in turn"
- plural-agnostic, count-agnostic, role-descriptive

This costs nothing to write, and it cannot rot. A reader who wants the list can
produce it in seconds; they cannot recover a stale list without doing exactly
that anyway. Restating derivable state is a maintenance contract paid forever
to save someone a lookup they can do for free.

## Stress it before you write it

If you are about to name something specific, imagine the target changing in the
three ordinary ways:

- it **gains** a member
- it **loses** one
- it is **renamed**

Does your sentence go false? If yes, and nothing would catch it, you are writing
a liability. Rephrase it, or do not write it.

## The shapes this takes

Every one of these was true when written. That is the point: nothing separates a
stale one from a live one except going and checking — the work the binding was
supposed to save.

- naming the members of something: a member is added, the list is now wrong
- a census or line count: drifts on the next commit
- a listed set of supported formats or backends: one is dropped
- "unlike X, this one ..." — X is deleted, and the sentence now contrasts with
  nothing
- "see X, the live example" — X is deleted
- a rule hand-copied into several places: one copy is retired and the others
  keep asserting it. This is the expensive one. A retired rule left copies
  behind, one of which named a formula as the entire defense against an attack
  — so an auditor checking that defense would have verified a rule the code had
  stopped using, and concluded it held.

## If you must bind, make it checked

In descending order of preference:

1. **Generate it.** A table produced from the source cannot drift.
2. **Assert it.** A test that fails when the count or the set changes turns an
   unchecked binding into a checked one.
3. **Cite by SYMBOL** — a function, type, or exact string someone can grep.
   A stale symbol fails loudly the moment anyone looks; a stale line number
   silently points at unrelated code and gets believed.

"I will keep it updated" is not a mechanism. It is the absence of one, and it
has never held.

## What DOES earn an unchecked binding

Three things, and they are all judgment a reader cannot derive:

1. **A WHY.** A rationale, a constraint, a rejected alternative, a trap. "This
   is a local copy, and that is not a preference: the remote fetcher resolves a
   ref to a single file, so a directory-form bundle cannot be fetched at all."
   Nobody can compute that. It is the entire value of the comment.
2. **An invariant that genuinely depends on that exact thing** — then cite it by
   symbol, per above.
3. **A pointer somebody could not find alone**: where the authority lives, which
   of two similar mechanisms governs, what to search for.

If what you are writing is none of these, it is decoration with a maintenance
bill attached.

## When you find one already stale, delete it

Prefer deleting an unmaintained list or census over correcting it. Correcting
one entry makes every remaining entry look verified, which is worse than the
honest signal that nobody is maintaining any of it.

---

CLAUDE.md: 100-200 lines max, overflow to per-folder files. Include: tech stack with versions, architecture (folder purposes), build/test/lint/deploy commands, project-specific rules AI would otherwise violate. Exclude: language syntax, linter-enforced patterns, anything Claude gets right unprompted. Test each line: would Claude err without it? No → delete. When Claude errs, add the correction immediately.

---

## Isolation: specify both axes

Creating, configuring, or delegating to a ctxloom agent (`ctxloom agent
set`, `run --agent`, `agent_run`)? Set both axes explicitly — never rely
on the default:

- **runtime** (`host` | `container-rootless` | `container-rootful`,
  the agent binding's `runtime:`) isolates the PROCESS. There is
  deliberately no "any container" value: rootless and rootful differ
  in UID mapping, so a workload can genuinely require one.
- **workspace** (`none`|`worktree`, per-invocation `--workspace` /
  `agent_run`'s `workspace`) isolates the FILES.

An ownership mismatch is FATAL, never a substitution. Asking for
`container-rootful` where only rootless is reachable is a fatal
ClassIsolation finding (exit 3), not a quiet downgrade to the other
mode — and `--degraded` falls back to the HOST, never to the other
ownership mode.

They're independent: `container-rootless` can still mount the
workspace at the SAME absolute path as the live project (process
isolated, edits still land where the editor already looks); `worktree`
still runs the engine on the host (the editor goes blind to that tree
by design — results return via the delegated-agent merge flow, not
live edits). Picking one says nothing about the other.

Unspecified means `host`+`none` — isolated on NEITHER axis. That's a
default, not a decision. Host runtime is not a security boundary
between agents: the coordinator credential is readable in the process
environment of any same-uid process, and that token IS identity, so a
host-runtime agent can read another host-runtime agent's credential and
speak as that agent. Containers are the actual boundary: they make
isolation a property of the runtime, not a request to the engine —
some vendor CLIs ignore env-var isolation hints and write
credentials/state to a global path regardless.

A bad or missing agent name silently degrades to `host`+`none` with only
a stderr warning, discarding the runtime and permissions you asked for —
confirm the name resolves before trusting the isolation you requested.

---

# llm-tool-killer (ltk)

This project may run **ltk**, a pre-tool hook that inspects each shell
command before it executes and redirects it when a rule matches. Where
ctxloom shapes the context you see, ltk guides the commands you run.

## What it does

ltk parses the real command (resolving variables, unwrapping trivial
wrappers and sub-shells) and matches it against the project's rules in
`.ltk/config.yaml`. The first matching `deny` wins and returns a
`message`/`suggest` telling you what to run instead. Example:

    go test ./...   ->   blocked: "Run tests through the task runner."
                    ->   retry with `just test`

## How to work with it

- Treat a redirect as guidance, not a failure: read the suggestion and
  retry the command the way the rule asks.
- Prefer the project's task runner (e.g. `just <target>`) over invoking
  build/test/lint tools directly.
- **Agents do not cut releases.** ltk blocks `git tag` and release
  commands. Prepare the version bump and PR; a human (or CI) cuts the tag.

## What it is not

ltk is a cooperative redirect, not a sandbox. If explicitly instructed
to work around a rule the agent can, so it makes the easy, accidental
path the right one rather than enforcing hard isolation. For strict
"never" boundaries, run the agent in a container.

---

# taskloom

Persistent task tracking. Tasks live in a per-project append-only log
(~/.ctxloom/tasks/<project-id>.jsonl) and are keyed by harp IDs
(e.g. `swift-amber-falcon`). Statuses: `In Progress`, `To Do`,
`Deferred`, `Done`, `Archived`.

## MCP tools (served by `taskloom mcp`)

- `task_list({statuses?, term?, include_completed?, include_summary?})`
  — list/filter tasks. Set `include_summary: true` to also get
  per-status counts plus the in-progress harp IDs.
- `task_add({text, status?, trigger?})` — add a task with a fresh
  harp ID. Default status is `"To Do"`; `"Deferred"` requires a
  `trigger` (the condition that should revive it).
- `task_set_status({harp_id, status, trigger?})` — move a task
  between statuses.
- `task_edit({harp_id, text})` — replace a task's text in place.

Tasks are created and updated only through these tools (or the
`taskloom` CLI). The harp ID appears in `task_list` output so you can
reference a specific task in later calls.

## Use `--json` FIRST, paired with jq — on every command

EVERY taskloom command takes `--json` (shorthand for `--format json`;
`--format` also accepts yaml, toml, text, markdown). If you are an agent,
`--json` is your surface — reach for it before anything else, and pipe it
through jq. The default text output is for a person reading a terminal.

    # one task's exact body — no header to strip, no offset to guess
    taskloom show <harp> --json | jq -r '.text'

    # every harp; every harp carrying a tag
    taskloom list --all --json | jq -r '.[].harp_id'
    taskloom list --all --json | jq -r '.[]|select(.tags[]?=="urgent")|.harp_id'

`Task`'s JSON tags are a cross-surface contract: the CLI and the MCP tools
emit the same snake_case keys, so one jq filter works against either.

**If `--json` fails, or a field you need is missing from it, REPORT THAT
as a defect** — then route around it if you must to finish the job. Both,
not either. The workaround unblocks you; the report is what gets the gap
closed, and it is the half that gets skipped.

This matters more here than it would for a third-party tool: taskloom,
ctxloom and reprise are ours, and we are their alpha users. There is no
upstream to file against and no other user base to hit the gap first. A
missing field that nobody reports simply stays missing, and the next agent
pays the same tax — which is exactly how the text-scraping habit took hold
instead of anyone noticing every command already speaks JSON.

Why this is stated so firmly: every way text-parsing breaks here produces a
WRONG ANSWER rather than an error.

- `taskloom show`'s text form prints a multi-line header. Capturing a body
  by dropping a fixed number of lines can leave a `tags:` line as line 1 —
  and line 1 IS the subject, so a following `edit` (which replaces the
  whole text) silently renames the task.
- In `--compact` rows, `[x]` is one field but `[ ]` is two, so a positional
  field split drops every OPEN task and keeps only the closed ones. The
  output looks well-formed and is wrong in exactly one direction.

Note the CLI flag names differ from the MCP parameter names: `--status`
(singular) and `--all`, not `statuses`/`include_completed`. A mistyped flag
errors, but a mis-scoped query returns a confident empty list — confirm the
flag exists before believing emptiness.

## Check the log before you start, and again before you finish

**Before starting work**, look for open tasks that touch what you are
about to change. One may already hold the root cause, a decision
someone already made, or a constraint you would otherwise rediscover
the hard way — and it may show that someone else is mid-flight in the
same files. Search by AREA, not just by title, since a task about your
code may be named for its symptom:

    taskloom list --term <symbol, path, or error string>
    taskloom list --tag-query <area>

**Before finishing**, scan again for tasks in the same area. If your
change satisfies one, say so and offer to close it — quote what the
task asked for and what you actually did, so the reader can judge
rather than take your word. If it satisfies a task only in part, edit
the task to record what is now done and what remains, instead of
leaving it whole and letting the next person redo the finished half.

## Filing a task

ONE issue per task. A row carrying two questions gets one of them
answered and the other silently dropped; a row carrying a question AND
the work it gates cannot be closed without lying about one half. If you
are about to write "and also", stop and file two, each naming the other.

LOOK BEFORE YOU FILE. Search by AREA and by SYMPTOM, not just by the
title you have in mind — a task about your code is often named for the
thing it broke:

    taskloom list --all --json | jq -r '.[]|select(.text|test("<symbol>"))|.harp_id'
    taskloom list --term <symbol|path|error> --json
    taskloom list --tag-query <area>

A duplicate is worse than a missing task: two rows describing one
problem drift apart, and whichever is read first looks complete.

WRITE IT ACTIONABLE COLD. The reader has no memory of the session that
filed it and cannot ask you. State what is wrong or wanted, HOW IT WAS
FOUND so they can reproduce it, and what would settle it. For a decision,
state the QUESTION and the options — not just the problem — and tag it
`human` so it is findable as one queue.

RECORD THE LOCATION WHILE YOU HAVE IT. Scouting and verifying is the only
moment `touches:` and `sig:` are free: the path and the symbol are already
in front of you. Filled in then, they cost nothing; recovered later, they
cost the same search twice. A task without them cannot tell a dispatcher
whether it collides with anything.

LINK RELATED WORK with `relates:<harp>`, on both rows. Prose that says
"see <harp>" is invisible to a query, so the connection exists only for
whoever happens to read that paragraph. A split, a root cause and its
symptom, a decision and the work it gates — link them, or the second one
gets solved twice.

## The tag axes, and what each one answers

Independent axes, each answering a different question. Conflating them is
what makes a log unqueryable;
the full rules live in `docs/task-tagging-standard.md`.

- `triage:level=` — how bad is it if we ship without this, as an
  INTEGER 1-5, lower is worse. `1` data loss, or a trust/isolation
  boundary breached. `2` unusable, or SUCCEEDS WITHOUT DOING THE THING.
  `3` wrong, but a workaround exists. `4` low or no user impact. `5`
  does not exist yet. Security is folded in BY CONSEQUENCE — there is
  no second security scale to cross-reference.
  It is an integer so it SORTS and so relational queries work:
  `--tag-query 'triage:level<=2'` is everything that must not ship.
  Exactly one per task; the range is enforced, so `0` and `6` are
  refused.
- `area:` — which subsystem. Exactly one per task.
- `touches:` — a repo-relative FILE path this task will EDIT.
  Repeatable. This is what says whether two tasks can run at once: two
  agents editing one file collide whatever symbols each touched.
- `sig:` — `package.Symbol` whose contract changes. Repeatable.

Rate the level FROM THE CODE, never from the task's own prose: a task
description outlives the code it describes, and tasks here have named
functions deleted commits earlier.

Locate work by NAME — a file path or a symbol — never by line number.
A stale symbol fails loudly the moment someone greps for it; a stale
line number silently points at unrelated code and is believed.

When you rename a file or change a symbol's contract, fix the
`touches:`/`sig:` tags that name it. Nothing else will: a tag has no
compiler and no gate.

Both halves matter for the same reason: a task nobody rereads gets
solved twice, and a task silently satisfied but left open is
indistinguishable from work never done.

## Plan stamping

When you edit a plan file (`CURRENT_PLAN.md`, `*-plan.md`,
`docs/*-plan.md`), the active session's harp name is auto-stamped
into the file's YAML frontmatter `sessions:` list. Plans and
sessions cross-reference without a separate database.
