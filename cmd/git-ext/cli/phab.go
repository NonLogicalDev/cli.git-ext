package cli

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/NonLogicalDev/nld.git-ext/lib/clitools"
	"github.com/NonLogicalDev/nld.git-ext/lib/shutils/arc"
	"github.com/NonLogicalDev/nld.git-ext/lib/shutils/git"

	"gopkg.in/alecthomas/kingpin.v2"
)

type jsA = []interface{}
type jsM = map[string]interface{}

type phabCLI struct {
	kingpin.CmdClause

	diffUpdateFlag   string
	diffCatchAllArgs []string

	diffMessageCopySrc string
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

	// Msg
	c = cli.Command("msg", "Get diff information from phab to HEAD commit.").
		Action(cli.doDiffMessageCopy)
	c.Arg("revisionid", "The revision id to copy the message from.").Required().
		StringVar(&cli.diffMessageCopySrc)

	// Sync
	c = cli.Command("sync", "Get diff information from phab to HEAD commit.").
		Action(cli.doSyncRevision)

	// NoQA:
	_ = c
}

func (cli *phabCLI) doList(ctx *kingpin.ParseContext) error {
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

func (cli *phabCLI) doDiff(ctx *kingpin.ParseContext) error {
	var updateRev string
	if cli.diffUpdateFlag != "" {
		updateRev = cli.diffUpdateFlag
	}

	arc.Diff("git:HEAD^1", updateRev, cli.diffCatchAllArgs)
	return nil
}

func (cli *phabCLI) doDiffMessageCopy(ctx *kingpin.ParseContext) error {
	rx := regexp.MustCompile(`\d+`)

	rev_id_str := rx.FindString(cli.diffMessageCopySrc)
	if len(rev_id_str) == 0 {
		clitools.UserErrorStr("Diff", "Incorrect revision name %v.", cli.diffMessageCopySrc)
	}
	rev_id, err := strconv.Atoi(rev_id_str)
	clitools.UserError(err)

	request, _ := json.MarshalIndent(jsM{
		"revision_id": rev_id,
	}, "", "  ")

	res, err := arc.ConduitCall("differential.getcommitmessage", request)
	clitools.UserError(err)

	output := map[string]interface{}{}
	err = json.Unmarshal([]byte(res), &output)
	fmt.Println(output["response"])
	return nil
}

func (cli *phabCLI) doSyncRevision(ctx *kingpin.ParseContext) error {
	message, err := git.Cmd("show", "-s", "--format=%B", "HEAD").Run().Value()
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
