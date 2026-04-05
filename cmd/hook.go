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
	Long:      `Generate the shell hook script for worktree with the 'wrk' function.`,
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
	fmt.Printf(`# worktree shell setup
wrk() {
    # If we're in completion mode, call worktree directly without processing
    if [ -n "${COMP_LINE}" ]; then
        worktree "$@"
        return $?
    fi

    local dir_path=""
    local exit_code=0

    # Stream output line by line and check for delimiter
    while IFS= read -r line; do
        if [[ "$line" == %s* ]]; then
            # Found delimiter, extract directory path
            dir_path="${line#%s}"
        else
            # Regular output, print immediately
            echo "$line"
        fi
    done < <(worktree "$@" 2>&1)

    # Capture the exit code from the worktree command
    exit_code=${PIPESTATUS[0]}

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
