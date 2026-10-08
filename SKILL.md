---
name: worktree
description: wrk and git worktrees — isolated per-branch working directories. Always use this skill when the user mentions wrk, worktrees, working on multiple branches at once, switching branches or directories, finding where a branch is checked out, or cleaning up stale branches — even if they never say the word worktree. Use it to create, find, switch, list, or clean up worktrees via the `worktree` CLI.
allowed-tools: Bash(worktree:*), Bash(git:*)
compatibility: Linux, macOS
metadata:
  author: jonathan
  version: "1.1.0"
---

# Worktree CLI

`worktree --help` is authoritative; this covers agent-relevant usage and traps.

## The `wrk` rule

The human says `wrk`, you run `worktree`. Never run bare `wrk`, because it is
an interactive-only shell function for `cd`, and on machines with homebrew's
`wrk` benchmarker it resolves to that tool instead. Parse the
`__WORKTREE_CD__<path>` line from the binary's output and `cd` yourself.

## Stay on the tool

Handle worktree operations with `worktree` subcommands only — never run
parallel `git worktree ...` calls for the same job, because they bypass the
tool's naming, markers, and safety rails while burying the answer in duplicate
output. Plain `git` stays fine for diff/log/commit; reach for `git worktree`
only to diagnose a `worktree` failure, and say that's what you're doing.

## Layout

The main checkout holds `.git/`; linked worktrees live in
`.{repo}.worktrees/<name>/` next to it. The name usually equals the branch,
but `add`/`new` accept a custom name, so never assume they match — resolve
through the tool.

## Existence check

Run `worktree switch <name>` — it resolves exact names, branch names, and
globs. Exit 0 plus a `__WORKTREE_CD__<path>` line means it exists (and gives
you the directory); non-zero means missing or ambiguous. Run `worktree list`
to enumerate (`>` is main, `*` is current); run `worktree status` for dirty
state, ahead/behind vs the default branch, and last-commit age, and check it
before deleting anything.

## Commands

- Run `new`/`add <branch> [name]` to create a worktree; read its path from
  `__WORKTREE_CD__`.
- Run `switch [pattern]` to resolve a worktree path; no args targets main.
- Run `switch -b [branch]` to switch branches without changing directory.
  Bare `-b` selects the remote default branch in the main worktree, or the
  branch matching the directory name in a linked worktree.
- Run `rm [pattern...]` to remove worktrees (never main); `-D` also deletes
  the branch; globs accepted.
- Run `clean [-o 2w] [-n] [-f]` to drop worktrees with no commits past the
  threshold. It skips main, current, and dirty worktrees (`-f` overrides);
  `-n` previews, including what would be pruned.
