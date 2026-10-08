---
name: worktree
description: wrk and git worktrees — isolated per-branch working directories. Always use this skill when the user mentions wrk, worktrees, working on multiple branches at once, switching branches or directories, finding where a branch is checked out, or cleaning up stale branches — even if they never say the word worktree. Use it to create, find, switch, list, or clean up worktrees via the `wrk` CLI.
allowed-tools: Bash(wrk:*), Bash(worktree:*), Bash(git:*)
compatibility: Linux, macOS
metadata:
  author: jonathan
  version: "2.0.0"
---

# wrk CLI

`wrk --help` is authoritative; this covers agent-relevant usage and traps.
(`worktree` remains as a legacy alias for the same binary.)

## The `wrk` rule

The human says `wrk`, you run `wrk` — it is the binary, installed by
`go install github.com/bungogood/wrk@latest`. In interactive
shells a `wrk` function wraps it to change directories; you don't need that:
parse the `wrk:cd:<path>` line from the output and `cd` yourself.

## Stay on the tool

Handle worktree operations with `wrk` subcommands only — never run parallel
`git worktree ...` calls for the same job, because they bypass the tool's
naming, markers, and safety rails while burying the answer in duplicate
output. Plain `git` stays fine for diff/log/commit; reach for `git worktree`
only to diagnose a `wrk` failure, and say that's what you're doing.

## Layout

The main checkout holds `.git/`; linked worktrees live in
`.{repo}.worktrees/<name>/` next to it. The name usually equals the branch,
but `add`/`new` accept a custom name, so never assume they match — resolve
through the tool.

## Existence check

Run `wrk switch <name>` — it resolves exact names, branch names, and globs.
Exit 0 plus a `wrk:cd:<path>` line means it exists (and gives you the
directory); non-zero means missing or ambiguous. Run `wrk list` to enumerate
(`>` is main, `*` is current); run `wrk status` for dirty state, ahead/behind
vs the default branch, and last-commit age, and check it before deleting
anything.

## Commands

- Run `wrk new <branch> [name]` to create a worktree with a brand-new branch.
- Run `wrk add <branch> [name]` to attach an existing local or remote branch
  (remote is automatic — `origin` when configured, else the first remote — so
  `add abc` tracks `origin/abc` when there is no local `abc`; `-R` overrides).
- Both print the new path via `wrk:cd:`.
- Run `wrk switch [pattern]` to resolve a worktree path; no args targets main.
- Run `wrk switch -b [branch]` to switch branches without changing directory.
  Bare `-b` selects the remote default branch in the main worktree, or the
  branch matching the directory name in a linked worktree.
- Run `wrk rm [pattern...]` to remove worktrees (never main); `-D` also
  deletes the branch; globs accepted.
- Run `wrk clean [-o 2w] [-n] [-f]` to drop worktrees with no commits past the
  threshold. It skips main, current, and dirty worktrees (`-f` overrides);
  `-n` previews, including what would be pruned.
