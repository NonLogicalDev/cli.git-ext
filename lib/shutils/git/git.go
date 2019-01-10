package git

import (
	"fmt"
	"strings"

	shutils "github.com/NonLogicalDev/nld.lib.go.shutils"
)

func GetUpstream() (string, error) {
	return RawGetAbbrevRef("@{upstream}").Run().Value()
}

func GetRoot() (string, error) {
	return RawGetRoot().Run().Value()
}

func GetSha(ref string) (string, error) {
	return RawGetSha(ref).Run().Value()
}

func GetMergeBase(refA, refB string) (string, error) {
	return RawGetMergeBase(refA, refB).Run().Value()
}

func ListObjectsInRange(refA, refB string) ([]string, error) {
	listStr, err := RawListObjectsInRange(refA, refB).Run().Value()
	if err != nil {
		return nil, err
	}
	return strings.Split(listStr, "\n"), nil
}

func ListBranches() ([]string, error) {
	listStr, err := RawListBranches().Run().Value()
	if err != nil {
		return nil, err
	}
	return strings.Split(listStr, "\n"), nil
}

func GetCommitWithFormat(sha, format string) (string, error) {
	message, err := Cmd(
		"show", "-s", fmt.Sprintf("--format=%v", format), sha,
	).Run().Value()

	return message, err
}

func GetSymbolicRefsForSHA(sha string) ([]string, error) {
	listStr, err := Cmd("for-each-ref", "--points-at", sha, "--format", "%(refname:short)").Run().Value()
	if err != nil {
		return nil, err
	}

	return strings.Split(listStr, "\n"), nil
}

/*
	Raw Command Helpers
*/

func Cmd(args ...interface{}) *shutils.ShCMD {
	return shutils.Cmd("git", args...)
}

func RawGetRoot() *shutils.ShCMD {
	return Cmd("rev-parse", "--show-toplevel")
}

func RawGetSha(ref string) *shutils.ShCMD {
	return Cmd("rev-parse", ref)
}

func RawGetAbbrevRef(ref string) *shutils.ShCMD {
	return Cmd("rev-parse", "--abbrev-ref", ref)
}

func RawGetMergeBase(refA, refB string) *shutils.ShCMD {
	return Cmd("merge-base", refA, refB)
}

func RawListObjectsInRange(refA, refB string) *shutils.ShCMD {
	return Cmd("rev-list", fmt.Sprintf("%v..%v", refA, refB))
}

func RawGetObjectContents(ref string) *shutils.ShCMD {
	return Cmd("cat-file", "-p", ref)
}

func RawGetCommitStat(ref string) *shutils.ShCMD {
	return Cmd("show", "--oneline", "--stat", ref)
}

func RawListBranches() *shutils.ShCMD {
	return Cmd("branch", "--list", "-a", "--format=%(refname:short)")
}

func RawSetBranch(ref, name string, force bool) *shutils.ShCMD {
	var args = []interface{}{
		"branch", "--create-reflog",
	}
	if force {
		args = append(args, "-f")
	}
	args = append(args, name)
	args = append(args, ref)

	return Cmd(args...)
}

func RawUnSetBranch(name string, force bool) *shutils.ShCMD {
	var args = []interface{}{
		"branch", "--create-reflog", "-D",
	}
	if force {
		args = append(args, "-f")
	}
	args = append(args, name)

	return Cmd(args...)
}
