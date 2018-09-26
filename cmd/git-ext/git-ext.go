package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"fmt"
	"github.com/nonlogicaldev/nld.git-ext/lib/shutils/git"
)

var  (
	cliParser  = kingpin.New("git-ext", "Command line utils extending git functionality.")

	dbgCMD = cliParser.Command("dbg", "Debug command for development")

	phabCMD = cliParser.Command("phab", "Integration with phabricator.")
	phabListCMD = phabCMD.Command("list", "List current pending stacked revisions on the current branch.")

	stackCMD = cliParser.Command("stack", "Integration with phabricator.")
	stackEditCMD = stackCMD.Command("edit", "Launch interactive rebase session to edit a given commit from history.")
)

func main()  {
	command := kingpin.MustParse(cliParser.Parse(os.Args[1:]))

	switch command {
	case "dbg":
		doDbg()
	case "phab list":
		doPhabList()
	}
}

func doPhabList()  {

}

func doDbg()  {
	fmt.Println(git.GetRoot())
}
