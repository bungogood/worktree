// Spike: exercise gitoxide (gix) for wrk's hot paths and time them.
// Usage: gitoxide-spike <main-worktree> <linked-worktree> <base-branch>
use std::collections::HashSet;
use std::time::Instant;

fn main() {
    let args: Vec<String> = std::env::args().collect();
    // Fast path: `first <path>` prints only the dirty boolean for one repo.
    if args.len() == 3 && args[1] == "first" {
        let repo = gix::open(&args[2]).expect("open");
        let mut iter = repo
            .status(gix::progress::Discard)
            .expect("status builder")
            .into_iter(None)
            .expect("status iter");
        println!("{}", matches!(iter.next(), Some(Ok(_))));
        return;
    }
    let repo_path = &args[1];
    let linked_path = &args[2];
    let base_name = &args[3];

    // 1+2. open main repo and linked worktree (.git file -> commondir)
    let t = Instant::now();
    let repo = gix::open(repo_path).expect("open main");
    println!("open-main: {:?}", t.elapsed());

    let t = Instant::now();
    let lrepo = gix::open(linked_path).expect("open linked");
    println!("open-linked: {:?}", t.elapsed());

    // 3. status (dirty check + full matrix) on both
    for (name, r) in [("main", &repo), ("linked", &lrepo)] {
        let t = Instant::now();
        let mut files = 0usize;
        let status = r
            .status(gix::progress::Discard)
            .expect("status builder")
            .into_iter(None)
            .expect("status iter");
        for change in status {
            change.expect("status item");
            files += 1;
        }
        println!(
            "status-{}-full: {:?} dirty={} files={}",
            name,
            t.elapsed(),
            files > 0,
            files
        );

        // Starship-style: one symbol needs just the first change, and the
        // lazy iterator short-circuits the walk.
        let t = Instant::now();
        let mut first = r
            .status(gix::progress::Discard)
            .expect("status builder")
            .into_iter(None)
            .expect("status iter");
        let dirty = matches!(first.next(), Some(Ok(_)));
        println!("status-{}-first: {:?} dirty={}", name, t.elapsed(), dirty);
    }

    // 4. branch last-commit dates (all local branches)
    let t = Instant::now();
    let mut count = 0usize;
    let platform = repo.references().expect("references");
    for r in platform.local_branches().expect("local branches") {
        let r = r.expect("branch ref");
        let id = r.id();
        let commit = repo.find_commit(id).expect("peel branch");
        let _when = commit.committer().expect("committer").time;
        let _name = r.name().shorten().to_string();
        count += 1;
    }
    println!("branch-dates: {:?} count={}", t.elapsed(), count);

    // 5. ahead/behind of linked HEAD vs base (two-sided revwalk)
    let t = Instant::now();
    let head_id = lrepo.head_id().expect("linked head");
    let base_id = lrepo
        .rev_parse_single(base_name.as_str())
        .expect("resolve base");
    let ahead = count_reachable(&lrepo, head_id.detach(), base_id.detach());
    let behind = count_reachable(&lrepo, base_id.detach(), head_id.detach());
    println!("ahead-behind: {:?} +{}/-{}", t.elapsed(), ahead, behind);
}

// Commits reachable from `from` but not from `exclude`.
fn count_reachable(
    repo: &gix::Repository,
    from: gix::ObjectId,
    exclude: gix::ObjectId,
) -> usize {
    let mut excl = HashSet::new();
    for info in repo.rev_walk([exclude]).all().expect("walk exclude") {
        excl.insert(info.expect("exclude item").id);
    }
    let mut n = 0usize;
    for info in repo.rev_walk([from]).all().expect("walk from") {
        let info = info.expect("from item");
        if !excl.contains(&info.id) {
            n += 1;
        }
    }
    n
}
