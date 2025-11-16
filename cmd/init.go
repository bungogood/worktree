package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/bungogood/worktree/pkg"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:       "init <shell>",
	Short:     "Generate shell initialization script",
	Long:      `Generate the initialization script for worktree with the 'wrk' alias.`,
	ValidArgs: []string{"bash"},
	Hidden:    true,
	Args:      cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		shell := args[0]
		if shell != "bash" {
			fmt.Fprintln(os.Stderr, "Only bash is currently supported")
			os.Exit(1)
		}

		// Output the bash initialization script
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

		// Generate bash completions for the worktree command
		fmt.Print(genBashCompletion())

		// Add completion for wrk alias with default file completion
		fmt.Println(`
# Enable completion for wrk alias
complete -o default -F __start_worktree wrk`)
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
}

func genBashCompletion() string {
	var buf bytes.Buffer
	RootCmd.GenBashCompletion(&buf)

	// Rewrite compgen lines to avoid prefix filtering
	result := strings.ReplaceAll(
		buf.String(),
		`done < <(compgen -W "${out}" -- "$cur")`,
		`done < <(printf "%s" "${out}")`,
	)

	return result
}
