# Push Prompt — Git Workflow for TaxiCheck

Use this prompt every time a fix or new feature needs to be committed and
pushed to the repository. This keeps the history clean, branch names
consistent, and ensures everything lands on `dev`.

## Rules

1. **Never push directly to `dev`, `main`, or any stable/version branch.**
2. Every fix or feature gets its **own branch**.
   - Fixes: `fix/<short-description>`
   - Features: `feature/<short-description>`
   - Docs: `docs/<short-description>`
3. Branch names are `kebab-case`, all lowercase, short and descriptive.
4. All branches are created **from `dev`** and all fixes/features **merge/push to `dev`**.
5. Test before committing (`go test ./...`, `go vet ./...`, `gofmt -w .`).
6. Never commit `.env`, build artifacts (`taxiprijs`, `dist/`), or secrets.
7. Keep commits small, focused, and message-style: `fix: ...`, `feat: ...`, `docs: ...`.

## Workflow

### 1. Make sure `dev` is up to date

```bash
git checkout dev
git pull origin dev
```

### 2. Create a dedicated branch

```bash
git checkout -b fix/<short-description>   # or feature/<...> or docs/<...>
```

### 3. Implement the change

- Only touch files relevant to this one fix/feature.
- Do not mix unrelated changes into the branch.

### 4. Test and format

```bash
go test ./...
go vet ./...
gofmt -w .
```

### 5. Stage and review

```bash
git status
git diff
git add <only the intended files>
```

Never `git add .` blindly — stage only what belongs to this change.

### 6. Commit with a descriptive message

```bash
git commit -m "fix: describe the fix concisely"
```

Use conventional prefixes:
- `fix: ` for bug fixes
- `feat: ` for new features
- `docs: ` for documentation
- `refactor: ` for non-behavior changes

Reference the issue number when one exists (e.g. `fix: ... (resolves #12)`).

### 7. Push the branch

```bash
git push -u origin fix/<short-description>
```

### 8. Open a PR / merge into `dev`

```bash
gh pr create --base dev --head fix/<short-description> --title "..." --body "..."
```

or, if you are allowed to merge directly:

```bash
git checkout dev
git merge fix/<short-description>
git push origin dev
```

> Prefer a PR to `dev` so changes are reviewable. After merging, delete the
> feature branch locally and remotely (`git branch -d`, `git push origin --delete`).

## Summary

| Action            | Branch                 |
|-------------------|------------------------|
| New feature       | `feature/<desc>` -> `dev` |
| Bug fix           | `fix/<desc>` -> `dev`   |
| Docs              | `docs/<desc>` -> `dev`  |

Never touch `main` or stable version branches from feature work.
