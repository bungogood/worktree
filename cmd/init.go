package cmd

import (
	"fmt"
	"os"

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
    local result
    result="$(worktree "$@")"
    local exit_code=$?
    
    if [ $exit_code -eq 0 ]; then
        # Check if the last line contains the delimiter
        local last_line="$(echo "$result" | tail -n 1)"
        
        if [[ "$last_line" == %s* ]]; then
            # Extract directory path after delimiter
            local dir_path="${last_line#%s}"
            
            # Print all output except the last line (the delimiter line)
            # Use sed to remove the last line (compatible with macOS)
            local output="$(echo "$result" | sed '$d')"
            if [ -n "$output" ]; then
                echo "$output"
            fi
            
            # Change directory
            if [ -d "$dir_path" ]; then
                cd "$dir_path" || return 1
            fi
        elif [ -n "$result" ]; then
            # No delimiter, just echo the output
            echo "$result"
        fi
    fi

    return $exit_code
}
`, pkg.CD_DELIMITER, pkg.CD_DELIMITER)

		// Generate bash completions for the worktree command
		RootCmd.GenBashCompletion(os.Stdout)

		// Add completion for wrk alias with default file completion
		fmt.Println(`
# Enable completion for wrk alias
complete -o default -F __start_worktree wrk`)
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
}
