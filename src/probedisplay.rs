//! Display helpers mirroring pkg/list.go + the status row format.

use super::git::{Repo, Worktree};
use std::time::SystemTime;

/// Marker glyph: main first.
pub fn marker(repo: &Repo, wt: &Worktree) -> &'static str {
    if repo.is_main(wt) {
        "> "
    } else if wt.path == repo.current_path {
        "* "
    } else {
        "  "
    }
}

/// Bare "name [branch]" label without markers.
pub fn label(wt: &Worktree) -> String {
    if !wt.branch.is_empty() && wt.branch != wt.name {
        format!("{} [{}]", wt.name, wt.branch)
    } else {
        wt.name.clone()
    }
}

/// Row renders in canonical order: age, state, marker, label.
pub fn row(age: &str, state: &str, marker: &str, label: &str) -> String {
    format!("{age:<7} {state:<6} {marker}{label}")
}

/// Compact age: 45m, 3h, 5d, 2w, 2w3d — like HumanizeAge.
pub fn humanize_age(now: SystemTime, then: SystemTime) -> String {
    let secs = now
        .duration_since(then)
        .map(|d| d.as_secs())
        .unwrap_or(0);
    let days = secs / 86400;
    if days >= 7 {
        let (w, r) = (days / 7, days % 7);
        return if r > 0 {
            format!("{w}w{r}d")
        } else {
            format!("{w}w")
        };
    }
    if days >= 1 {
        return format!("{days}d");
    }
    if secs >= 3600 {
        return format!("{}h", secs / 3600);
    }
    format!("{}m", secs / 60)
}

