package cli

import (
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/NonLogicalDev/nld.git-ext/lib/shutils/git"
	"github.com/NonLogicalDev/nld.git-ext/lib/clitools"
	"strings"
	"fmt"
	"regexp"
	"io/ioutil"
	"os"
)

const(
	branchLabelPrefix = "D/"
)

type stackCLI struct {
	kingpin.CmdClause

	rebaseEditPrefix string
	rebaseEditFile string

	editTargetRef string

	rebaseExtraArgs []string

	labelDeleteBranches bool
}

func RegisterStackCLI(p *kingpin.Application) {
	cli := &stackCLI{CmdClause: *p.Command("stack", "Git macros to make working with a stack of commits easier.").Alias("st")}
	var c *kingpin.CmdClause

	// Rebase Edit
	c = cli.Command("rebase-edit", "Rewrite rebase todo file.").Hidden().
		Action(cli.doRebaseFileRewrite)
	c.Flag("branchLabelPrefix", "Target SHA branchLabelPrefix to mark for edits.").
		Required().
		StringVar(&cli.rebaseEditPrefix)
	c.Arg("file", "Rebase file to read and overwrite.").
		Required().
		ExistingFileVar(&cli.rebaseEditFile)

	// Edit
	c = cli.Command("edit", "Launch interactive rebase session to edit a given commit from history.").
		Alias("e").
		Action(cli.doEdit)
	c.Arg("target", "Target commit sha or ref to edit in rebase session.").
		Required().
		HintAction(func() (choices []string) {
			branches, _ := git.ListBranches()
			for _, branch := range branches {
				if strings.HasPrefix(branch, branchLabelPrefix) {
					choices = append(choices, branch)
				}
			}
			return
		}).
		StringVar(&cli.editTargetRef)

	c = cli.Command("rebase", "Launch interactive rebase session against upstream.").
		Alias("rb").
		Action(cli.doRebase)
	c.Arg("args", "Extra args to pass to `git rebase`, example `rebase -- -x 'make build'`").
		StringsVar(&cli.rebaseExtraArgs)

	// Label
	c = cli.Command("label", "Label the revisions on a stack.").
		Alias("l").
		Action(cli.doLabel)
	c.Flag("delete", "Target commit sha or ref to edit in rebase session.").Short('D').
		BoolVar(&cli.labelDeleteBranches)

	// NoQA:
	_ = c
}

func (cli *stackCLI) doRebaseFileRewrite(ctx *kingpin.ParseContext) (error) {
	file := cli.rebaseEditFile
	prefix := cli.rebaseEditPrefix

	RGX := regexp.MustCompile(`^(\w+)\s+([A-Fa-f0-9]+)\s+(.*)$`)

	fileRaw, err := ioutil.ReadFile(file)
	clitools.UserError(err)

	out, err := os.OpenFile(file, os.O_RDWR, 0666)
	clitools.UserError(err)

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

	return nil
}

func (cli *stackCLI) doRebase(ctx *kingpin.ParseContext) (error) {
	upstreamName, err := git.GetUpstream()
	clitools.UserError(err)

	gitArgs := []interface{}{
		"rebase", "-i", upstreamName,
	}
	for _, a := range cli.rebaseExtraArgs {
		gitArgs = append(gitArgs, a)
	}
	clitools.UserError(
		git.Cmd(gitArgs...).
			PipeStdout(os.Stdout).PipeStderr(os.Stderr).
			Run().Err(),
	)

	return nil
}

func (cli *stackCLI) doEdit(ctx *kingpin.ParseContext) (error) {
	targetSha, err := git.GetSha(cli.editTargetRef)
	clitools.UserError(err)

	upstreamName, err := git.GetUpstream()
	clitools.UserError(err)

	mergeBaseCommit, err := git.GetMergeBase(upstreamName, "HEAD")
	clitools.UserError(err)

	fmt.Println(mergeBaseCommit)
	gitEditCMD := fmt.Sprintf("%s stack rebase-edit --branchLabelPrefix=%s ", os.Args[0], targetSha[:7])

	fmt.Println(gitEditCMD)
	clitools.UserError(
		git.
			Cmd("rebase", "-i", mergeBaseCommit).
			SetENV( "GIT_SEQUENCE_EDITOR", gitEditCMD).
			SetENV( "LANG", "en_US.UTF-8").
			PipeStdout(os.Stdout).PipeStderr(os.Stderr).
			Run().Err(),
	)

	return nil
}

func (cli *stackCLI) doLabel(ctx *kingpin.ParseContext) (error) {
	if cli.labelDeleteBranches {
		return cli.doLabelDelete(ctx)
	}
	return cli.doLabelCreate(ctx)
}

func (cli *stackCLI) doLabelDelete(ctx *kingpin.ParseContext) (error) {
	branches, err := git.ListBranches()
	clitools.UserError(err)

	for _, branchName := range branches {
		if strings.HasPrefix(branchName, branchLabelPrefix) {
			clitools.UserError(
				git.RawUnSetBranch(branchName, true).Run().
					PipeStdout(os.Stdout).
					Run().Err(),
			)
		}
	}

	return nil
}

func (cli *stackCLI) doLabelCreate(ctx *kingpin.ParseContext) (error) {
	upstreamName, err := git.GetUpstream()
	clitools.UserError(err)

	merrgeBaseCommit, err := git.GetMergeBase(upstreamName, "HEAD")
	clitools.UserError(err)

	pendingCommitList, err := git.ListObjectsInRange(fmt.Sprintf("%v^1", merrgeBaseCommit), "HEAD")
	clitools.UserError(err)

	for idx := range pendingCommitList {
		branchName := fmt.Sprintf("%v%02d", branchLabelPrefix, idx)
		sha := pendingCommitList[len(pendingCommitList)-1-idx]

		fmt.Printf("%02d| Creating branch: %v -> %v\n", idx, branchName, sha)
		clitools.UserError(
			git.RawSetBranch(sha, branchName, true).
				PipeStdout(os.Stdout).
				Run().Err(),
		)
	}

	return nil
}
