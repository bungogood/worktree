//! Git access layer (gitoxide). Mirrors pkg/repository.go + pkg/worktree.go.

use std::collections::HashMap;
use std::path::{Path, PathBuf};
use std::time::SystemTime;

#[derive(Clone, Debug)]
pub struct Worktree {
    pub path: PathBuf,
    pub branch: String,
    pub name: String,
}

pub struct Repo {
    /// Absolute path of the main worktree (holds `.git/`).
    pub main_path: PathBuf,
    /// Absolute path of the worktree containing the cwd.
    pub current_path: PathBuf,
    pub worktrees: Vec<Worktree>,
}

/// Discover the repo from the cwd: main worktree via the common git dir,
/// current worktree by prefix match — same rules as LoadRepo.
pub fn discover_repo() -> Result<Repo, String> {
    let cwd = std::env::current_dir().map_err(|e| format!("cannot get cwd: {e}"))?;
    let probe = gix::discover(&cwd).map_err(|e| format!("not in a git repository: {e}"))?;
    let common = probe
        .common_dir()
        .to_path_buf();
    let main_path = common
        .parent()
        .ok_or_else(|| "cannot locate main worktree".to_string())?
        .to_path_buf();

    // Worktrees (main first is handled at sort time).
    let mut worktrees = Vec::new();
    for repo in probe
        .worktrees_including_main()
        .map_err(|e| format!("failed to list worktrees: {e}"))?
    {
        let repo = repo.map_err(|e| format!("failed to open worktree: {e}"))?;
        let Some(dir) = repo.workdir() else { continue };
        let dir = dir.to_path_buf();
        let branch = repo
            .head_name()
            .ok()
            .flatten()
            .map(|n| n.shorten().to_string())
            .unwrap_or_default();
        let name = dir
            .file_name()
            .map(|n| n.to_string_lossy().into_owned())
            .unwrap_or_default();
        worktrees.push(Worktree {
            path: dir,
            branch,
            name,
        });
    }
    if worktrees.is_empty() {
        return Err("current directory is not inside any worktree".to_string());
    }

    let current_path = worktrees
        .iter()
        .find(|wt| cwd.starts_with(&wt.path))
        .map(|wt| wt.path.clone())
        .ok_or_else(|| "current directory is not inside any worktree".to_string())?;

    Ok(Repo {
        main_path,
        current_path,
        worktrees,
    })
}

impl Repo {
    pub fn is_main(&self, wt: &Worktree) -> bool {
        wt.path == self.main_path
    }

    /// Main first, current second, then alphabetical — like SortedWorktrees.
    pub fn sorted_worktrees(&self) -> Vec<Worktree> {
        let mut sorted = self.worktrees.clone();
        sorted.sort_by(|a, b| {
            let rank = |wt: &Worktree| {
                if self.is_main(wt) {
                    0
                } else if wt.path == self.current_path {
                    1
                } else {
                    2
                }
            };
            rank(a)
                .cmp(&rank(b))
                .then_with(|| a.name.cmp(&b.name))
        });
        sorted
    }

    /// Last-commit times for every local branch in one pass.
    pub fn branch_dates(&self) -> HashMap<String, SystemTime> {
        let mut dates = HashMap::new();
        let Ok(repo) = gix::open(&self.main_path) else {
            return dates;
        };
        let Ok(refs) = repo.references() else {
            return dates;
        };
        let Ok(iter) = refs.local_branches() else {
            return dates;
        };
        for r in iter.filter_map(Result::ok) {
            let name = r.name().shorten().to_string();
            if let Ok(commit) = repo.find_commit(r.id()) {
                if let Ok(sig) = commit.committer() {
                    if let Ok(t) = sig.time() {
                        dates.insert(
                            name,
                            std::time::UNIX_EPOCH
                                + std::time::Duration::from_secs(t.seconds as u64),
                        );
                    }
                }
            }
        }
        dates
    }
}

/// Dirty check: true on the first reported change (starship-style boolean).
pub fn is_dirty(path: &Path) -> Result<bool, String> {
    let repo = gix::open(path).map_err(|e| format!("cannot open repo: {e}"))?;
    let mut iter = repo
        .status(gix::progress::Discard)
        .map_err(|e| format!("status failed: {e}"))?
        .into_iter(None)
        .map_err(|e| format!("status failed: {e}"))?;
    Ok(matches!(iter.next(), Some(Ok(_))))
}

/// Last activity: branch tip from the batch map, else per-worktree HEAD.
pub fn last_activity(
    repo: &Repo,
    wt: &Worktree,
) -> Result<SystemTime, String> {
    if !wt.branch.is_empty() {
        if let Some(t) = repo.branch_dates().get(&wt.branch) {
            return Ok(*t);
        }
    }
    let r = gix::open(&wt.path).map_err(|e| format!("cannot open repo: {e}"))?;
    let id = r.head_id().map_err(|e| format!("no HEAD: {e}"))?;
    let commit = r
        .find_commit(id.detach())
        .map_err(|e| format!("cannot read HEAD: {e}"))?;
    let secs = commit
        .committer()
        .map_err(|e| format!("cannot read committer: {e}"))?
        .time()
        .map_err(|e| format!("cannot read time: {e}"))?
        .seconds as u64;
    std::time::UNIX_EPOCH
        .checked_add(std::time::Duration::from_secs(secs))
        .ok_or_else(|| "bad timestamp".to_string())
}
