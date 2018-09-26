package arc

import (
	"regexp"
	"github.com/nonlogicaldev/nld.git-ext/lib/shutils"
	"fmt"
	"os"
)

var PhabDiffRe = regexp.MustCompile(`(?m)^\s*Differential Revision:\s*(.+)`)

func RevisionFromMessage(message string) (string) {
	 groups := PhabDiffRe.FindStringSubmatch(message)
	 if len(groups) == 0 {
	 	return "No Revision"
	 }
	 return groups[1]
}

func Diff(base, updateRevision string, extArgs *[]string) (error) {
	args := []interface{}{
		"diff", fmt.Sprintf("--base=%v", base),
	}

	if len(updateRevision) != 0 {
		args = append(args, fmt.Sprintf("--update=%v", updateRevision))
	}

	if extArgs != nil {
		for _, a := range *extArgs {
			args = append(args, a)
		}
	}

	return shutils.
		Cmd("arc", args...).
		PipeStderr(os.Stderr).PipeStdout(os.Stdout).
		Run().Err()
}
