package git

import "github.com/nonlogicaldev/nld.git-ext/lib/shutils"

func RunGit(args ...string) (string, error) {
	return shutils.Run("git", args...)
}

