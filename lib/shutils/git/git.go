package git

import (
	"github.com/nonlogicaldev/nld.git-ext/lib/shutils"
	"strings"
)

func Run(args ...interface{}) *shutils.ShCMD {
	return shutils.Cmd("git", args...)
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

func GetRoot() (string) {
	output, err := Run("rev-parse", "--show-toplevel").X.CombinedOutput()
	panicOnErr(err)
	return strings.TrimSpace(string(output))
}

func GetSha(ref string) (string) {
	output, err := Run("rev-parse", ref).X.CombinedOutput()
	panicOnErr(err)
	return strings.TrimSpace(string(output))
}

func GetAbbrevRef(ref string) (string) {
	output, err := Run("rev-parse", "--abbrev-ref", ref).X.CombinedOutput()
	panicOnErr(err)
	return strings.TrimSpace(string(output))
}

func GetMergeBase(refA, refB string) (string) {
	output, err := Run("merge-base", refA, refB).X.CombinedOutput()
	panicOnErr(err)
	return strings.TrimSpace(string(output))
}
