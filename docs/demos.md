# Demo Recording Guide

Use this guide when recording short terminal demos for `README.md`.

## Prompt setup

Use a short path + git branch prompt with blue branch text:

```bash
PS1='\W \[\e[38;5;33m\]$(__git_prompt_segment)\[\e[0m\]\$ '
```

Suggested flow:

wrk new feat-shell-hook

```bash
wrk new feat-shell-hook
wrk list
wrk rm chore-*
wrk switch
```

Stop recording with `Ctrl-D`.

## Record with asciinema

```bash
asciinema rec --overwrite /tmp/worktree-demo.cast
```

## Convert cast to GIF with agg

```bash
agg \
  --speed 1.25 \
  --idle-time-limit 0.8 \
  --fps-cap 60 \
  --cols 100 --rows 16 \
  /tmp/worktree-demo.cast docs/worktree-demo.gif
```
