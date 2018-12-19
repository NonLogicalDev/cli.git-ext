To install run
```
go get github.com/NonLogicalDev/nld.cli.git-ext/...
```

Documentation of commands:
```
usage: git-ext [<flags>] <command> [<args> ...]

Command line utils extending git functionality.

Flags:
  -h, --help  Show context-sensitive help (also try --help-long and --help-man).

Commands:
  help [<command>...]
    Show help.


  stack (alias=[st]) 
    Git macros to make working with a stack of commits easier.

  stack edit (alias=[e])  <target>
    Launch interactive rebase session to edit a given commit from history.


  stack rebase (alias=[rb])  [<args>...]
    Launch interactive rebase session against upstream.


  stack label (alias=[l])  [<flags>]
    Label the revisions on a stack.


  stack meta (alias=[m])  [<flags>] [<value>...]
    Operate on metadata of commit.



  phab
    Integration with phabricator.

  phab list
    List current pending stacked revisions on the current branch.


  phab diff [<flags>] [<args>...]
    Update or create a diff based on current commit.


  phab msg <revisionid>
    Get diff information from phab to HEAD commit.


  phab sync
    Get diff information from phab to HEAD commit.


  phab land
    Land current revision.
```
