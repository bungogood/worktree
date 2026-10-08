//! Display helpers mirroring pkg/list.go + the status row format.

use super::git::{Repo, Worktree};
use chrono::{DateTime, Local};
use std::time::SystemTime;

/// Marker glyph: main first — like GetWorktreeMarker.
fn marker(repo: &Repo, wt: &Worktree) -> &'static str {
    if repo.is_main(wt) {
        "> "
    } else if wt.path == repo.current_path {
        "* "
    } else {
        "  "
    }
}

/// `name [branch]` display, prefixed with the marker.
pub fn display(repo: &Repo, wt: &Worktree) -> String {
    let mut s = String::from(marker(repo, wt));
    s.push_str(&wt.name);
    if wt.branch != wt.name {
        s.push_str(&format!(" [{}]", wt.branch));
    }
    s
}

/// Pad to visible width (no color codes emitted, so plain padding).
pub fn pad_visible(s: &str, width: usize) -> String {
    let len = s.chars().count();
    if len >= width {
        return s.to_string();
    }
    format!("{s}{}", " ".repeat(width - len))
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

/// YYYY-MM-DD in local time.
pub fn format_date(t: SystemTime) -> String {
    let dt: DateTime<Local> = t.into();
    dt.format("%Y-%m-%d").to_string()
}
