package cli

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/NonLogicalDev/nld.cli.git-ext/lib/clitools"
	"github.com/NonLogicalDev/nld.cli.git-ext/lib/shutils/arc"
	"github.com/NonLogicalDev/nld.cli.git-ext/lib/shutils/git"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type jsA = []interface{}
type jsM = map[string]interface{}

type phabCLI struct {
	kingpin.CmdClause

	listBaseFlag string

	diffUpdateFlag   string
	diffCatchAllArgs []string

	diffMessageCopySrc string
}

func RegisterPhabCLI(p *kingpin.Application) {
	cli := &phabCLI{CmdClause: *p.Command("phab", "Integration with phabricator.")}
	var c *kingpin.CmdClause

	// List Command: ---------------------------------------------
	c = cli.Command("list", "List current pending stacked revisions on the current branch.").
		Action(func(context *kingpin.ParseContext) error {
			return cli.doList()
		})
	c.Flag("base", "Specifies the common base commit to start the listing from.").Short('b').
		StringVar(&cli.listBaseFlag)

	//------------------------------------------------------------

	// Diff Command: ---------------------------------------------
	c = cli.Command("diff", "Update or create a diff based on current commit.").
		Action(func(context *kingpin.ParseContext) error {
			return cli.doDiff(cli.diffUpdateFlag, cli.diffCatchAllArgs)
		})

	c.Flag("update", "A spefic revision to update.").
		StringVar(&cli.diffUpdateFlag)
	c.Arg("args", "Rest of the arguments will be passed to `arc diff`").
		StringsVar(&cli.diffCatchAllArgs)
	//------------------------------------------------------------

	// Msg Command: ----------------------------------------------
	c = cli.Command("msg", "Get message of a Phab revision in Git Commit format.").
		Action(func(context *kingpin.ParseContext) error {
			return cli.doDiffMessagePrint(cli.diffMessageCopySrc)
		})
	c.Arg("revisionid", "The revision id to show the message from.").Required().
		StringVar(&cli.diffMessageCopySrc)
	//------------------------------------------------------------

	// Sync Command: ---------------------------------------------
	c = cli.Command("sync", "Sync local HEAD commit's title to Phab.").
		Action(func(context *kingpin.ParseContext) error {
			return cli.doSyncRevision()
		})
	//------------------------------------------------------------

	// Land Command: ---------------------------------------------
	c = cli.Command("land", "Land current revision.").
		Action(func(context *kingpin.ParseContext) error {
			return cli.doLandRevision()
		})
	//------------------------------------------------------------

	// NoQA:
	_ = c
}

func (cli *phabCLI) doList() error {
	var err error

	upstreamName := cli.listBaseFlag
	if len(cli.listBaseFlag) == 0 {
		upstreamName, err = git.GetUpstream()
		clitools.UserError(err)
	}

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

func (cli *phabCLI) doDiff(diffUpdate string, extraFlags []string) error {
	var updateRev string
	if diffUpdate != "" {
		updateRev = diffUpdate
	}

	return arc.Diff("git:HEAD^1", updateRev, extraFlags)
}

func (cli *phabCLI) doDiffMessagePrint(revisionID string) error {
	out, err := arc.GetMSGForRevision(revisionID)
	clitools.UserError(err)

	fmt.Println(out)
	return nil
}

func (cli *phabCLI) doSyncRevision() error {
	message, err := git.GetCommitWithFormat("HEAD", "%B")
	clitools.UserError(err)

	revIdUrl, err := url.Parse(arc.RevisionFromMessage(message))
	clitools.UserError(err)

	revId := regexp.MustCompile(`D\d+`).FindString(revIdUrl.Path)
	title := strings.Split(message, "\n")[0]

	request, _ := json.MarshalIndent(jsM{
		"objectIdentifier": revId,
		"transactions": jsA{
			jsM{
				"type":  "title",
				"value": title,
			},
		},
	}, "", "  ")
	fmt.Println(string(request))

	res, err := arc.ConduitCall("differential.revision.edit", request)
	clitools.UserError(err)

	fmt.Println(res)
	return nil
}

func (cli *phabCLI) doLandRevision() error {
	// It is recommened to create the following shortcut
	// `git xland: "phab stack meta -p && phab sync && phab land"`

	arcLand := arc.Cmd("land", "--keep-branch").Unbuffer().Run()
	if arcLand.HasError() {
		clitools.UserError(arcLand.Err())
	}

	// gitLand := git.Cmd("push", "origin", "--", "HEAD:master").Unbuffer().Run()
	// if gitLand.HasError() {
	// 	clitools.UserError(gitLand.Err())
	// }

	return nil
}
