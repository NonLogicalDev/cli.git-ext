package arc

import (
	"bytes"
	"fmt"
	"os"
	"regexp"

	"github.com/NonLogicalDev/nld.git-ext/lib/shutils"
)

var PhabDiffRe = regexp.MustCompile(`(?m)^\s*Differential Revision:\s*(.+)`)

func Cmd(args ...interface{}) *shutils.ShCMD {
	return shutils.Cmd("arc", args...)
}

func RevisionFromMessage(message string) string {
	groups := PhabDiffRe.FindStringSubmatch(message)
	if len(groups) == 0 {
		return "No Revision"
	}
	return groups[1]
}

func Diff(base, updateRevision string, extArgs []string) error {
	args := []interface{}{
		"diff", fmt.Sprintf("--base=%v", base),
	}

	if len(updateRevision) != 0 {
		args = append(args, fmt.Sprintf("--update=%v", updateRevision))
	}

	for _, a := range extArgs {
		args = append(args, a)
	}

	return shutils.
		Cmd("arc", args...).
		PipeStderr(os.Stderr).PipeStdout(os.Stdout).PipeStdin(os.Stdin).
		Run().Err()
}

func ConduitCall(endpoint string, params []byte) (string, error) {
	request := bytes.NewReader(params)
	return shutils.
		Cmd("arc", "call-conduit", endpoint).
		PipeStdin(request).
		Run().Value()
}
