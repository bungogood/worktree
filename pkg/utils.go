package pkg

import "fmt"

const CD_DELIMITER = "__WORKTREE_CD__"

// ChangeDirectory outputs the directory change command for the wrk wrapper
func ChangeDirectory(path string) {
	fmt.Printf("%s%s\n", CD_DELIMITER, path)
}
