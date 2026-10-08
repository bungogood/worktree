---
name: worktree
description: wrk and git worktrees — isolated per-branch working directories. Triggers on wrk, worktree, worktrees. Use to create, find, switch, list, or clean up worktrees via the `worktree` CLI.
allowed-tools: Bash(worktree:*), Bash(git:*)
compatibility: Linux, macOS
metadata:
  author: jonathan
  version: "1.0.0"
---

# Worktree CLI

Each feature gets its own directory linked to the same repo, so in-progress
work stays isolated without stashing. `worktree --help` is authoritative for
flags; this covers the agent-relevant usage and traps.

## The `wrk` rule

When the human says `wrk`, run the `worktree` binary. Never run bare `wrk`:

- it is an interactive shell function whose only power is `cd` (a child
  process cannot change its parent's directory, so the function scrapes a
  `__WORKTREE_CD__<path>` line out of the binary's output and cds for you);
- on machines with the homebrew `wrk` HTTP benchmarker installed, bare `wrk`
  resolves to that instead.

You don't need the `cd` magic: parse the `__WORKTREE_CD__<path>` line and
`cd` (or set the working directory) yourself.

## Stay on the tool

For anything worktree-related, use `worktree` subcommands only — do not
duplicate the job with raw `git worktree ...` calls. They bypass the tool's
conventions and safety rails (naming, markers, dirty/main guards), and the
double output just buries the answer. Plain `git` stays fine for normal
version-control work (diff, log, commit); only reach for `git worktree`
when `worktree` itself errors unexpectedly and you need to diagnose — and
say that's what you're doing.

## Layout

- The main checkout holds `.git/`; linked worktrees live in
  `.{repo}.worktrees/<name>/` next to it.
- `name` usually equals the branch, but `add`/`new` accept a custom name.

## Existence check (scriptable)

`worktree switch <name>` resolves exact names, branch names, and globs:

- exit 0 + a `__WORKTREE_CD__<path>` line: it exists, `<path>` is its
  directory;
- exit non-zero (`no worktree found matching 'x'` / `matches multiple`):
  it does not exist, or the pattern is ambiguous.

`worktree list` enumerates all worktrees (`>` = main, `*` = current).
`worktree status` adds dirty state, ahead/behind vs the default branch, and
last-commit age — prefer it before deleting anything.

## Commands

- `new <branch> [name]` / `add <branch> [name]` — create a worktree (new
  branch / existing branch) and print its path via `__WORKTREE_CD__`.
- `switch [pattern]` — print the target path via `__WORKTREE_CD__`; no args
  targets the main worktree.
- `switch -b [branch]` — switch branch in place, no directory change. Bare
  `-b` targets the remote default branch in the main worktree, or the branch
  matching the directory name in a linked worktree.
- `rm [pattern...]` — remove worktrees (never main); `-D` also deletes the
  branch. Globs accepted.
- `clean [-o 2w] [-n] [-f]` — remove worktrees with no commits for longer
  than the threshold. Never touches main/current; skips dirty worktrees
  unless `-f`. `-n` previews, including what would be pruned.
