package git

import (
	"github.com/NonLogicalDev/nld.git-ext/lib/shutils"
	"os/exec"
	"github.com/pkg/errors"
	"fmt"
	"strings"
)

func GetUpstream() (string, error) {
	return wrapRaw(RawGetAbbrevRef("@{upstream}"))
}

func GetRoot() (string, error) {
	return wrapRaw(RawGetRoot())
}

func GetSha(ref string) (string, error) {
	return wrapRaw(RawGetSha(ref))
}

func GetMergeBase(refA, refB string) (string, error) {
	return wrapRaw(RawGetMergeBase(refA, refB))
}

func ListObjectsInRange(refA, refB string) ([]string, error) {
	listStr, err := wrapRaw(RawListObjectsInRange(refA, refB))
	if err != nil {
		return nil, err
	}
	return strings.Split(listStr, "\n"), nil
}

func ListBranches() ([]string, error) {
	listStr, err := wrapRaw(RawListBranches())
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

func RawGetRoot() (string, *shutils.ShCMD) {
	cmd := Cmd("rev-parse", "--show-toplevel").Run()
	return cmd.StdoutStr(), cmd
}

func RawGetSha(ref string) (string, *shutils.ShCMD) {
	cmd := Cmd("rev-parse", ref).Run()
	return cmd.StdoutStr(), cmd
}

func RawGetAbbrevRef(ref string) (string, *shutils.ShCMD) {
	cmd := Cmd("rev-parse", "--abbrev-ref", ref).Run()
	return cmd.StdoutStr(), cmd
}

func RawGetMergeBase(refA, refB string) (string, *shutils.ShCMD) {
	cmd := Cmd("merge-base", refA, refB).Run()
	return cmd.StdoutStr(), cmd
}

func RawListObjectsInRange(refA, refB string) (string, *shutils.ShCMD) {
	cmd := Cmd("rev-list", fmt.Sprintf("%v..%v", refA, refB)).Run()
	return cmd.StdoutStr(), cmd
}

func RawGetObjectContents(ref string) (string, *shutils.ShCMD) {
	cmd := Cmd("cat-file", "-p", ref).Run()
	return cmd.StdoutStr(), cmd
}

func RawGetCommitStat(ref string) (string, *shutils.ShCMD) {
	cmd := Cmd("show", "--oneline", "--stat", ref).Run()
	return cmd.StdoutStr(), cmd
}

func RawSetBranch(ref, name string, force bool) (string, *shutils.ShCMD) {
	var args = []interface{}{
		"branch", "--create-reflog",
	}
	if force {
		args = append(args, "-f")
	}
	args = append(args, name)
	args = append(args, ref)

	cmd := Cmd(args...).Run()
	return cmd.StdoutStr(), cmd
}

func RawUnSetBranch(name string, force bool) (string, *shutils.ShCMD) {
	var args = []interface{}{
		"branch", "--create-reflog", "-D",
	}
	if force {
		args = append(args, "-f")
	}
	args = append(args, name)

	cmd := Cmd(args...).Run()
	return cmd.StdoutStr(), cmd
}

func RawListBranches() (string, *shutils.ShCMD) {
	cmd := Cmd("branch", "--list", "-a", "--format=%(refname:short)").Run()
	return cmd.StdoutStr(), cmd
}

/*
	Internal
 */

 func wrapRaw(out string, cmd *shutils.ShCMD) (string, error) {
	 if cmd.HasError() {
		 return "", cmd.Err()
	 }
	 return out, nil
 }

func panicOnErr(err error) {
	if e, ok := err.(*exec.ExitError); ok {
		panic(errors.Errorf( "GitError[%v]: %s", string(e.String()), string(e.Stderr)))
	}
	if err != nil {
		panic(err)
	}
}
