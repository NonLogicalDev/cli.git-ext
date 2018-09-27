package main

import (
	"fmt"
	"github.com/NonLogicalDev/nld.git-ext/lib/shutils/arc"
	"github.com/NonLogicalDev/nld.git-ext/lib/shutils/git"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"strings"
	"io/ioutil"
	"regexp"
)

var (
	cliParser = kingpin.New("git-ext", "Command line utils extending git functionality.")

	// Phab CMD
	phabCMD     = cliParser.Command("phab", "Integration with phabricator.")

	phabListCMD = phabCMD.Command("list", "List current pending stacked revisions on the current branch.")

	phabDiffCMD = phabCMD.Command("diff", "Update or create a diff based on current commit.")
	phabDiffCMDUpdate = phabDiffCMD.Flag("update", "A spefic revision to update.").String()
	phabDiffCMDArgs = phabDiffCMD.Arg("args", "Rest of the arguments will be passed to `arc diff`").Strings()

	// Stack CMD
	stackCMD     = cliParser.Command("stack", "Integration with phabricator.")

	stackEditCMD = stackCMD.Command("edit", "Launch interactive rebase session to edit a given commit from history.")
	stackEditCMDTarget = stackEditCMD.Arg("target", "Target SHA to edit.").String()

	stackRebaseEditCMD = stackCMD.Command("rebase-edit", "Rewrite rebase todo file.")
	stackRebaseEditCMDPrefix = stackRebaseEditCMD.Flag("prefix", "Target SHA prefix to mark for edits.").Required().String()
	stackRebaseEditCMDFile   = stackRebaseEditCMD.Arg("file", "Rebase file to read and overwrite.").Required().ExistingFile()
)

func userErrorStr(t, msg string) {
	fmt.Printf("Error [%v]: %v\n", t, msg)
	os.Exit(128)
}

func userError(err error) {
	if err != nil {
		userErrorStr("Go", err.Error())
	}
}

func userPanic() {
	if r := recover(); r != nil {
		if err, ok := r.(error); ok {
			userError(err)
		}
	}
}

func main() {
	defer userPanic()

	cliParser.HelpFlag.Short('h')
	command := kingpin.MustParse(cliParser.Parse(os.Args[1:]))

	switch command {
	case "phab list":
		doPhabList()
	case "phab diff":
		doPhabDiff()
	case "stack edit":
		doStackEdit(*stackEditCMDTarget)
	case "stack rebase-edit":
		doRebaseFileRewrite(*stackRebaseEditCMDPrefix, *stackRebaseEditCMDFile)
	default:
		os.Exit(128)
	}
}

func doPhabDiff() {
	var updateRev string
	if phabDiffCMDUpdate != nil {
		updateRev = *phabDiffCMDUpdate
	}
	arc.Diff("git:HEAD^1", updateRev, phabDiffCMDArgs)
}

func doRebaseFileRewrite(prefix, file string) {
	RGX := regexp.MustCompile(`^(\w+)\s+([A-Fa-f0-9]+)\s+(.*)$`)
	fileRaw, err := ioutil.ReadFile(file)
	userError(err)

	out, err := os.OpenFile(file, os.O_RDWR, 0666)
	userError(err)

	fmt.Println("[REBASE_TODO]")
	for _, line := range strings.Split(string(fileRaw), "\n") {
		groups := RGX.FindStringSubmatch(line)
		if len(groups) > 0 {
			gCMD := groups[1]
			gSHA := groups[2]
			gComment := groups[3]

			if strings.HasPrefix(gSHA, prefix) {
				gCMD = "edit"
			}

			outLine := fmt.Sprintf("%s %s %s", gCMD, gSHA, gComment)
			fmt.Println("| ", outLine)
			fmt.Fprintln(out, outLine)
		} else {
			fmt.Fprintln(out, line)
		}
	}
	fmt.Printf("[/REBASE_TODO]\n\n")
}

func doStackEdit(target string) {
	targetSha, err := git.GetSha(target)
	userError(err)
	upstreamName, err := git.GetUpstream()
	userError(err)
	mergeBaseCommit, err := git.GetMergeBase(upstreamName, "HEAD")
	userError(err)

	fmt.Println(mergeBaseCommit)
	gitEditCMD := fmt.Sprintf("%s stack rebase-edit --prefix=%s ", os.Args[0], targetSha[:7])

	fmt.Println(gitEditCMD)
	git.
		Cmd("rebase", "-i", mergeBaseCommit).
		SetENV( "GIT_SEQUENCE_EDITOR", gitEditCMD).
		SetENV( "LANG", "en_US.UTF-8").
		PipeStdout(os.Stdout).PipeStderr(os.Stderr).
		Run()
}

func doPhabList() {
	upstreamName, err := git.GetUpstream()
	userError(err)
	merrgeBaseCommit, err := git.GetMergeBase(upstreamName, "HEAD")
	userError(err)
	pendingCommitList, err := git.ListObjectsInRange(merrgeBaseCommit, "HEAD")
	userError(err)
	for _, sha := range pendingCommitList {
		rawCommit, cmd := git.RawGetObjectContents(sha)
		userError(cmd.Err())
		statsRaw, cmd := git.RawGetCommitStat(sha)
		userError(cmd.Err())
		contentsCmd := git.Cmd("-c", "color.ui=always", "log", "--pretty=%C(red)%h%C(yellow)%d%C(reset)\n%s", "-n1", sha).Run()
		var stats string
		{
			s := strings.Split(statsRaw, "\n")
			stats = s[len(s)-1]
		}
		contents := contentsCmd.StdoutStr()
		rev := arc.RevisionFromMessage(rawCommit)
		fmt.Printf("[%s]\n%s\n%s\n\n", rev, contents, stats)
	}
}
