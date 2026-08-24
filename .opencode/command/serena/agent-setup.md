---
description: serena/agent-setup
---

## Serena (MCP code intelligence) — setup

This project has the `serena` bundle installed. Serena is
configured ENTIRELY by start-up flags on `serena start-mcp-server`,
recorded in the bundle's `mcp:` block. There is no runtime setting
the agent can change later, so any choice not made here is made by
default and is invisible afterwards.

If serena is being registered or reconfigured, put these to the
user. Do not assume defaults — each one changes agent behavior.

1. **Context** (`--context`). Determines which tools serena exposes
   and what behavioral prompt it injects.
   - `claude-code` — drops serena's redundant file tools and
     instructs the agent that native Read is forbidden for
     discovery and native Edit is forbidden on code files. This is
     what makes serena actually get used; it also overrides the
     project's own guidance on tool choice.
   - `agent` — neutral, no behavioral mandate. Serena's tools are
     available but the agent will keep defaulting to Read/Grep.
   - `ide`, `codex`, `antigravity` and others exist; `serena
     context list` enumerates them. Registration is per-server, so
     a project driving several engines picks ONE.

2. **Memory and onboarding** (`--mode no-memories`,
   `--mode no-onboarding`). Serena runs an onboarding pass and
   keeps its own notes under `.serena/memories/*.md`. Ask whether
   that should be on: it is a persistent context store sitting
   alongside ctxloom's session memory and taskloom, and nothing
   reconciles the three. Disabling both leaves serena as a pure
   code-intelligence layer.

3. **Other modes** (`--mode`). `planning`, `editing`,
   `interactive`, `one-shot`; `serena mode list` enumerates them.
   Repeatable, and they override the project/global defaults.

4. **Project resolution**. `--project-from-cwd` keeps an absolute
   path out of the bundle, resolving from the cwd of whichever
   session starts the server. It resolves ONCE at start-up and does
   not follow an agent into a worktree — an agent working in a
   different checkout is answered about the launching one, for
   reads as well as edits. Use `--project <path>` only to pin one
   fixed checkout deliberately.

5. **Container reach**. If any agent in this project is bound to
   `runtime: container`, confirm serena exists inside that image.
   A host-installed serena is invisible to a containerized agent
   and the server will simply fail to start.

Record the answers by editing the `mcp.serena.args` list in
`.ctxloom/content/bundles/serena.yaml`, then re-run `ctxloom trust`
so the changed executable surface is acknowledged.
