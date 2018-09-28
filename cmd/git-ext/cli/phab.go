package cli

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"github.com/NonLogicalDev/nld.git-ext/lib/shutils/git"
	"github.com/NonLogicalDev/nld.git-ext/lib/clitools"
	"strings"
	"github.com/NonLogicalDev/nld.git-ext/lib/shutils/arc"
	"fmt"
)

type phabCLI struct {
	kingpin.CmdClause

	diffUpdateFlag string
	diffCatchAllArgs []string
}

func RegisterPhabCLI(p *kingpin.Application) {
	cli := &phabCLI{CmdClause: *p.Command("phab", "Integration with phabricator.")}
	var c *kingpin.CmdClause

	// List:
	c = cli.Command("list", "List current pending stacked revisions on the current branch.").
		Action(cli.doList)

	// Diff:
	c = cli.Command("diff", "Update or create a diff based on current commit.").
		Action(cli.doDiff)
	c.Flag("update", "A spefic revision to update.").
		StringVar(&cli.diffUpdateFlag)
	c.Arg("args", "Rest of the arguments will be passed to `arc diff`").
		StringsVar(&cli.diffCatchAllArgs)

	// NoQA:
	_ = c
}

func (cli *phabCLI) doList(ctx *kingpin.ParseContext) (error) {
	upstreamName, err := git.GetUpstream()
	clitools.UserError(err)

	merrgeBaseCommit, err := git.GetMergeBase(upstreamName, "HEAD")
	clitools.UserError(err)

	pendingCommitList, err := git.ListObjectsInRange(merrgeBaseCommit, "HEAD")
	clitools.UserError(err)

	for _, sha := range pendingCommitList {
		rawCommit, err := git.RawGetObjectContents(sha).Run().Value()
		clitools.UserError(err)

		statsRaw, err := git.RawGetCommitStat(sha).Run().Value()
		clitools.UserError(err)

		var stats string
		{
			s := strings.Split(statsRaw, "\n")
			stats = s[len(s)-1]
		}

		contents, err := git.Cmd("-c", "color.ui=always", "log", "--pretty=%C(red)%h%C(yellow)%d%C(reset)\n%s", "-n1", sha).Run().Value()
		clitools.UserError(err)

		rev := arc.RevisionFromMessage(rawCommit)

		fmt.Printf("[%s]\n%s\n%s\n\n", rev, contents, stats)
	}
	return nil
}

func (cli *phabCLI) doDiff(ctx *kingpin.ParseContext) (error) {
	var updateRev string
	if cli.diffUpdateFlag != "" {
		updateRev = cli.diffUpdateFlag
	}
	arc.Diff("git:HEAD^1", updateRev, cli.diffCatchAllArgs)
	return nil
}
