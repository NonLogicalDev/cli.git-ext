package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/NonLogicalDev/nld.cli.git-ext/lib/clitools"
	"github.com/NonLogicalDev/nld.cli.git-ext/lib/shutils/git"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"bytes"
)

const (
	branchLabelPrefix = "D/"
)

var metadataPattern = regexp.MustCompile(`^\|(.*)\| (?s:(.*))`)
var metadataFormat = "|%v| %v"

var branchPattern = regexp.MustCompile(`D/\d+`)
var branchFormat = "D/%02d"

func metadataFromString(input string) (metadata string, message string) {
	groups := metadataPattern.FindStringSubmatch(input)
	if len(groups) > 2 {
		return groups[1], groups[2]
	}
	return "", input
}

func metadataToString(metadata, message string) (output string) {
	_, nMessage := metadataFromString(message)
	return fmt.Sprintf(metadataFormat, metadata, nMessage)
}

type stackCLI struct {
	kingpin.CmdClause

	rebaseEditPrefix string
	rebaseEditFile   string

	editTargetRef string

	rebaseExtraArgs []string

	labelBurninBranch   bool
	labelDeleteBranches bool

	metaGetFlag   bool
	metaPutFlag   bool
	metaValueArgs []string
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
	c.Flag("delete", "Delete the labels from the commits.").Short('d').
		BoolVar(&cli.labelDeleteBranches)
	c.Flag("burnin", "Burnin the labels into the commits.").Short('b').
		BoolVar(&cli.labelBurninBranch)

	// Meta
	c = cli.Command("meta", "Operate on metadata of commit.").
		Alias("m").
		Action(cli.doMeta)

	c.Flag("put", "Put metadata on commit.").Short('p').
		BoolVar(&cli.metaPutFlag)
	c.Flag("get", "Put metadata on commit.").Short('g').
		BoolVar(&cli.metaGetFlag)
	c.Arg("value", "Value of the arg to set.").
		StringsVar(&cli.metaValueArgs)

	// NoQA:
	_ = c
}

func (cli *stackCLI) doRebaseFileRewrite(ctx *kingpin.ParseContext) error {
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

func (cli *stackCLI) doRebase(ctx *kingpin.ParseContext) error {
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

func (cli *stackCLI) doEdit(ctx *kingpin.ParseContext) error {
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
			SetENV("GIT_SEQUENCE_EDITOR", gitEditCMD).
			SetENV("LANG", "en_US.UTF-8").
			PipeStdout(os.Stdout).PipeStderr(os.Stderr).
			Run().Err(),
	)

	return nil
}

func (cli *stackCLI) doLabel(ctx *kingpin.ParseContext) error {
	if cli.labelBurninBranch {
		return cli.doLabelBurnin(ctx)
	} else if cli.labelDeleteBranches {
		return cli.doLabelDelete(ctx)
	} else {
		return cli.doLabelCreate(ctx)
	}
}

func (cli *stackCLI) doLabelBurnin(ctx *kingpin.ParseContext) error {
	refs, err := git.GetSymbolicRefsForSHA("HEAD")
	clitools.UserError(err)
	for _, ref := range refs {
		if branchPattern.MatchString(ref) {
			cli.metaPutFlag = true
			cli.metaValueArgs = []string{ref}
			return cli.doMeta(ctx)
		}
	}
	return nil
}

func (cli *stackCLI) doLabelDelete(ctx *kingpin.ParseContext) error {
	branches, err := git.ListBranches()
	clitools.UserError(err)
	for _, branchName := range branches {
		if branchPattern.MatchString(branchName) {
			clitools.UserError(
				git.RawUnSetBranch(branchName, true).Run().
					PipeStdout(os.Stdout).
					Run().Err(),
			)
		}
	}
	return nil
}

func (cli *stackCLI) doLabelCreate(ctx *kingpin.ParseContext) error {
	upstreamName, err := git.GetUpstream()
	clitools.UserError(err)

	merrgeBaseCommit, err := git.GetMergeBase(upstreamName, "HEAD")
	clitools.UserError(err)

	pendingCommitList, err := git.ListObjectsInRange(fmt.Sprintf("%v^1", merrgeBaseCommit), "HEAD")
	clitools.UserError(err)

	for idx := range pendingCommitList {
		branchName := fmt.Sprintf(branchFormat, idx)
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

func (cli *stackCLI) doMeta(ctx *kingpin.ParseContext) error {
	if cli.metaPutFlag {
		message, err := git.GetCommitWithFormat("HEAD", "%B")
		clitools.UserError(err)

		_, rawMessage := metadataFromString(message)

		newMessage := rawMessage
		if len(cli.metaValueArgs) > 0 {
			rawMetadata := strings.Join(cli.metaValueArgs, ",")
			newMessage = metadataToString(rawMetadata, rawMessage)
		}

		messageReader := bytes.NewReader([]byte(newMessage))
		err = git.Cmd("commit", "--amend", "--file=-").Unbuffer().
			PipeStdin(messageReader).Run().Err()

		clitools.UserError(err)
		return nil
	} else {
		sha := "HEAD"
		if len(cli.metaValueArgs) > 0 {
			sha = cli.metaValueArgs[0]
		}

		message, err := git.GetCommitWithFormat(sha, "%B")
		clitools.UserError(err)

		metadata, _ := metadataFromString(message)
		if len(metadata) > 0 {
			fmt.Println(metadata)
		}

		return nil
	}
}
