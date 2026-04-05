package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:       "completion <shell>",
	Short:     "Generate shell completion script",
	Long:      `Generate the shell completion script for worktree.`,
	Aliases:   []string{"completions"},
	ValidArgs: []string{"bash"},
	Args:      cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		shell := args[0]
		if shell != "bash" {
			fmt.Fprintln(os.Stderr, "Only bash is currently supported")
			os.Exit(1)
		}

		fmt.Print(genBashCompletion())
		fmt.Println(`
# Enable completion for worktree
if [[ $(type -t compopt) = "builtin" ]]; then
    complete -o default -F __start_wrk worktree
else
    complete -o default -o nospace -F __start_wrk worktree
fi`)
	},
}

func init() {
	RootCmd.AddCommand(completionCmd)
}

func genBashCompletion() string {
	var buf bytes.Buffer
	RootCmd.GenBashCompletion(&buf)

	// Rewrite compgen lines to avoid prefix filtering.
	return strings.ReplaceAll(
		buf.String(),
		`done < <(compgen -W "${out}" -- "$cur")`,
		`done < <(printf "%s" "${out}")`,
	)
}
