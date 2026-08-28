# Issue Prompt — Debugging & Fixing TaxiCheck Issues

Use this prompt whenever you are assigned a GitHub issue to fix or debug
TaxiCheck on a specific platform / operating system / distro.

## Critical: Scope to the reported environment only

An issue almost always points to **one specific environment** (e.g. Windows,
a particular Linux distro, a particular terminal). Follow these rules:

1. **Reproduce and debug ONLY on the OS / distro / terminal reported in the
   issue.** Do not "fix" or change anything for other platforms unless the issue
   explicitly says so.
2. If you are debugging for **Windows**, only change Windows-specific code and
   behavior. Leave Linux/macOS/Arch/Omarchy code untouched.
3. If you are debugging for **Arch** (or any specific distro), do not modify
   code or defaults that would affect other distros.
4. Keep platform-specific work in its own task / commit, clearly labeled with
   the target platform.
5. Do not introduce unrelated refactors or "nice to haves". Change only what
   fixes the reported issue.
6. After fixing, verify the fix on the reported environment. If you cannot
   reproduce/verify there, say so clearly in the PR/comment.

## Reproduce first

- Read the issue carefully: OS, distro, terminal, Go version, TaxiCheck
  version/branch, exact error output.
- Reproduce the error using the exact steps, input, and environment the user
  gave.
- Capture the exact error output — this is the most important piece of info.

## When working on a fix

- Check if the issue number was assigned (e.g. `#12`).
- Create a dedicated branch from `dev` following `prompts/push_prompt.md`:
  - `git checkout dev && git pull origin dev`
  - `git checkout -b fix/<short-description>`
- Implement the minimal fix.
- Test: `go test ./...`, `go vet ./...`, `gofmt -w .`
- Commit and push the branch.
- If a GitHub issue number exists, mention it in the commit and PR body.

## Before pushing

Read `prompts/push_prompt.md` and follow it exactly:

- Branch from `dev`, push branch to `origin`, PR / merge into `dev`.
- Conventional commit message (`fix:`).
- Never push directly to `dev`, `main`, or stable version branches.

## Report format

When done, provide:

- Environment: OS / distro / terminal / Go version / TaxiCheck version
- Root cause
- What you changed (files + summary)
- How it was verified, or the note that you couldn't verify on the target env
- Branch name + PR/commit reference, and the issue number it references
