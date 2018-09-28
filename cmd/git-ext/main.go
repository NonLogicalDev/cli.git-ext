package main

import (
	"os"
	"github.com/NonLogicalDev/nld.git-ext/lib/clitools"
	"github.com/NonLogicalDev/nld.git-ext/cmd/git-ext/cli"

	"gopkg.in/alecthomas/kingpin.v2"
)

func init()  {
	kingpin.DefaultUsageTemplate = helpTemplate
	kingpin.LongHelpTemplate = helpTemplateLong
	kingpin.ManPageTemplate = helpTemplateManPage
}

func setUpParser() (*kingpin.Application) {

	cliParser := kingpin.New("git-ext", "Command line utils extending git functionality.")
	cliParser.HelpFlag.Short('h')

	// Register Handlers
	cli.RegisterPhabCLI(cliParser)
	cli.RegisterStackCLI(cliParser)

	return cliParser
}

func main() {
	defer clitools.UserFriendlyPanic(true)
	kingpin.MustParse(setUpParser().Parse(os.Args[1:]))
}
