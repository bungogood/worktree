package cmd

import (
	"fmt"
	"os"

	"github.com/bungogood/worktree/pkg"
	"github.com/spf13/cobra"
)

var hookCmd = &cobra.Command{
	Use:       "hook <shell>",
	Short:     "Generate shell hook script",
	Long:      `Generate the shell hook script for wrk with directory switching.`,
	ValidArgs: []string{"bash"},
	Args:      cobra.ExactArgs(1),
	Run:       runHook,
}

func runHook(cmd *cobra.Command, args []string) {
	shell := args[0]
	if shell != "bash" {
		fmt.Fprintln(os.Stderr, "Only bash is currently supported")
		os.Exit(1)
	}

	// Output the bash hook script.
	fmt.Printf(`# wrk shell setup
wrk() {
    # If we're in completion mode, call the binary directly without processing.
    # 'command' bypasses this function so the binary runs instead of recursing.
    if [ -n "${COMP_LINE}" ]; then
        command wrk "$@"
        return $?
    fi

    # Capture output first so the binary's exit code survives: reading it
    # through process substitution would report 'read' hitting EOF (1)
    # instead (PIPESTATUS only covers real pipelines).
    local output exit_code dir_path line
    output="$(command wrk "$@" 2>&1)"
    exit_code=$?

    # Stream output line by line and check for delimiter
    dir_path=""
    while IFS= read -r line; do
        if [[ "$line" == %s* ]]; then
            # Found delimiter, extract directory path
            dir_path="${line#%s}"
        else
            # Regular output, print immediately
            echo "$line"
        fi
    done <<< "$output"

    # If we found a directory path, change to it
    if [ -n "$dir_path" ] && [ -d "$dir_path" ]; then
        cd "$dir_path" || return 1
    fi

    return $exit_code
}
`, pkg.CD_DELIMITER, pkg.CD_DELIMITER)
}

func init() {
	RootCmd.AddCommand(hookCmd)
}
