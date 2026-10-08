//! wrk (Rust port) — git worktree manager.
//!
//! Status: `list` and `status` are ported. Everything else reports
//! "not yet ported" until its turn comes. See PORTING.md for the order.

use clap::{CommandFactory, Parser, Subcommand};
use std::path::PathBuf;

mod git;
mod probedisplay;

const CD_DELIMITER: &str = "wrk:cd:";

#[derive(Parser)]
#[command(name = "wrk", about = "Git worktree manager")]
struct Cli {
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    /// List all worktrees
    #[command(alias = "ls")]
    List,
    /// Show worktree status (dirty, ahead/behind, age)
    Status {
        /// Remote to resolve the default branch from
        #[arg(short = 'R', long)]
        remote: Option<String>,
    },
    /// Generate shell hook script
    Hook {
        /// Shell to generate for
        shell: String,
    },
    /// Generate shell completion script
    Completion {
        /// Shell to generate for
        shell: String,
    },
    /// Create a new branch as a worktree
    New {
        _branch: String,
        _name: Option<String>,
    },
    /// Add an existing branch as a worktree
    Add {
        _branch: String,
        _name: Option<String>,
        /// Remote to resolve branches from
        #[arg(short = 'R', long)]
        _remote: Option<String>,
    },
    /// Switch to a worktree
    Switch {
        _pattern: Option<String>,
        /// Switch branch in place without changing directory
        #[arg(short = 'b', long)]
        _branch: bool,
        /// Remote to resolve the default branch from
        #[arg(short = 'R', long)]
        _remote: Option<String>,
    },
    /// Remove worktrees
    #[command(alias = "rm")]
    Remove {
        _patterns: Vec<String>,
    },
    /// Remove stale worktrees
    Clean {
        /// Remove worktrees with no commits for longer than this
        #[arg(short = 'o', long, default_value = "2w")]
        _older: String,
        /// Also remove worktrees with uncommitted changes
        #[arg(short = 'f', long)]
        _force: bool,
        /// List stale worktrees without removing them
        #[arg(short = 'n', long)]
        _dry_run: bool,
    },
}

fn main() {
    let cli = Cli::parse();
    let result = match cli.command {
        Commands::List => cmd_list(),
        Commands::Status { remote } => cmd_status(remote),
        Commands::Hook { shell } => cmd_hook(&shell),
        Commands::Completion { shell } => cmd_completion(&shell),
        _ => Err("not yet ported from Go — see PORTING.md".to_string()),
    };
    if let Err(e) = result {
        eprintln!("Error: {e}");
        std::process::exit(1);
    }
}

fn cmd_list() -> Result<(), String> {
    let repo = git::discover_repo().map_err(|e| e.to_string())?;
    for wt in repo.sorted_worktrees() {
        println!("{}", probedisplay::display(&repo, &wt));
    }
    Ok(())
}

fn cmd_status(remote: Option<String>) -> Result<(), String> {
    use rayon::prelude::*;
    let repo = git::discover_repo().map_err(|e| e.to_string())?;
    let base = repo
        .default_branch_for(repo.resolve_remote(remote.as_deref()))
        .unwrap_or_default();
    let worktrees: Vec<git::Worktree> = repo
        .sorted_worktrees()
        .into_iter()
        .filter(|wt| wt.path.exists())
        .collect();

    let rows: std::collections::HashMap<PathBuf, String> = worktrees
        .par_iter()
        .map(|wt| {
            let state = match git::is_dirty(&wt.path) {
                Ok(true) => "dirty",
                Ok(false) => "clean",
                Err(_) => "?",
            };
            let progress = if base.is_empty() {
                "?".to_string()
            } else {
                match git::ahead_behind(&wt.path, &base) {
                    Ok((a, b)) => format!("+{a}/-{b}"),
                    Err(_) => "?".to_string(),
                }
            };
            let age = match git::last_activity(&repo, wt) {
                Ok(t) => {
                    let now = std::time::SystemTime::now();
                    format!(
                        "{} ({})",
                        probedisplay::humanize_age(now, t),
                        probedisplay::format_date(t)
                    )
                }
                Err(_) => "?".to_string(),
            };
            (
                wt.path.clone(),
                format!("{state:6} {progress:>7} {age}"),
            )
        })
        .collect();

    for wt in repo.sorted_worktrees() {
        if !wt.path.exists() {
            println!(
                "{} {:6} {:>7} {}",
                probedisplay::pad_visible(&probedisplay::display(&repo, &wt), 28),
                "gone",
                "?",
                "?"
            );
            continue;
        }
        match rows.get(&wt.path) {
            Some(row) => println!(
                "{} {row}",
                probedisplay::pad_visible(&probedisplay::display(&repo, &wt), 28)
            ),
            None => println!(
                "{} {:6} {:>7} {}",
                probedisplay::pad_visible(&probedisplay::display(&repo, &wt), 28),
                "?",
                "?",
                "?"
            ),
        }
    }
    Ok(())
}

fn cmd_hook(shell: &str) -> Result<(), String> {
    if shell != "bash" {
        return Err("Only bash is currently supported".to_string());
    }
    print!("{}", hook_script());
    Ok(())
}

fn hook_script() -> String {
    format!(
        r#"# wrk shell setup
wrk() {{
    # If we're in completion mode, call the binary directly without processing.
    # 'command' bypasses this function so the binary runs instead of recursing.
    if [ -n "${{COMP_LINE}}" ]; then
        command wrk "$@"
        return $?
    fi

    local output exit_code dir_path line
    output="$(command wrk "$@" 2>&1)"
    exit_code=$?

    dir_path=""
    while IFS= read -r line; do
        if [[ "$line" == {d}* ]]; then
            dir_path="${{line#{d}}}"
        else
            echo "$line"
        fi
    done <<< "$output"

    if [ -n "$dir_path" ] && [ -d "$dir_path" ]; then
        cd "$dir_path" || return 1
    fi

    return $exit_code
}}
"#,
        d = CD_DELIMITER
    )
}

fn cmd_completion(shell: &str) -> Result<(), String> {
    use clap_complete::{generate, shells::Bash};
    use std::io::stdout;
    if shell != "bash" {
        return Err("Only bash is currently supported".to_string());
    }
    let mut cmd = Cli::command();
    generate(Bash, &mut cmd, "wrk", &mut stdout());
    Ok(())
}
