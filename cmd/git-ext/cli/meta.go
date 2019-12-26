package cli

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/NonLogicalDev/cli.git-ext/lib/clitools"
	"github.com/NonLogicalDev/cli.git-ext/lib/shutils/git"
	"gopkg.in/alecthomas/kingpin.v2"
)

var metadataPattern = regexp.MustCompile(`^(.*) \| \[(.*)]$`)

func metadataFromString(message string) (title string, meta string, body string) {
	fistLineEndIndex := strings.Index(message, "\n")
	title = message

	if fistLineEndIndex > 0 {
		titleEndIndex := fistLineEndIndex

		bodyEndIndex := len(message)
		bodyStartIndex := fistLineEndIndex
		if bodyStartIndex > bodyEndIndex {
			bodyStartIndex = bodyEndIndex
		}

		title = message[0:titleEndIndex]
		body = message[bodyStartIndex:bodyEndIndex]
	}

	groups := metadataPattern.FindStringSubmatch(title)
	if groups == nil {
		return title, "", body
	}
	return groups[1], groups[2], body
}

func metadataToString(title string, meta string, body string) (message string) {
	message = title
	if len(meta) > 0 {
		message = fmt.Sprintf("%s | [%s]", message, meta)
	}
	if len(body) > 0 {
		message = fmt.Sprintf("%s\n%s", message, body)
	}
	return message
}

type metaCLI struct {
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

func RegisterMetaCLI(p *kingpin.Application) {
	cli := &stackCLI{CmdClause: *p.Command("meta", "Git macros to make annotating commits easier.").Alias("m")}
	var c *kingpin.CmdClause

	// Set
	c = cli.Command("set", "Add metadata to commit..").
		Alias("s").
		Action(cli.doMetaSet)

	c.Arg("value", "Value of the arg to set.").
		Required().
		StringsVar(&cli.metaValueArgs)

	// Clear
	c = cli.Command("clear", "Clear metadata from commit.").
		Alias("d").
		Action(cli.doMetaClear)

	// View
	c = cli.Command("view", "Clear metadata from commit.").
		Alias("v").
		Action(cli.doMetaView)

	// NoQA:
	_ = c
}

func (cli *stackCLI) doMetaSet(ctx *kingpin.ParseContext) error {
	message, err := git.GetCommitWithFormat("HEAD", "%B")
	clitools.UserError(err)

	title, _, body := metadataFromString(message)
	newMeta := strings.Join(cli.metaValueArgs, ",")
	newMessageBuffer := []byte(metadataToString(title, newMeta, body))

	err = git.Cmd("commit", "--amend", "--file=-").
		Unbuffer().
		PipeStdin(
			bytes.NewReader(newMessageBuffer),
		).
		Run().Err()

	clitools.UserError(err)
	return nil
}

func (cli *stackCLI) doMetaClear(ctx *kingpin.ParseContext) error {
	sha := "HEAD"
	if len(cli.metaValueArgs) > 0 {
		sha = cli.metaValueArgs[0]
	}

	message, err := git.GetCommitWithFormat(sha, "%B")
	clitools.UserError(err)

	title, _, body := metadataFromString(message)
	newMessageBuffer := []byte(metadataToString(title, "", body))

	err = git.Cmd("commit", "--amend", "--file=-").
		Unbuffer().
		PipeStdin(
			bytes.NewReader(newMessageBuffer),
		).
		Run().Err()

	return nil
}
func (cli *stackCLI) doMetaView(ctx *kingpin.ParseContext) error {
	message, err := git.GetCommitWithFormat("HEAD", "%B")
	clitools.UserError(err)

	title, meta, body := metadataFromString(message)

	fmt.Printf(">> META: %s\n\n", meta)
	fmt.Printf(">> Title: %s\n\n", title)
	fmt.Printf(">> Body: %s\n\n", body)

	return nil
}
