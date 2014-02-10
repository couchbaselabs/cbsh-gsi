package commands

import (
	"flag"
	"fmt"
	"github.com/couchbaselabs/cbsh/api"
	"github.com/couchbaselabs/cbsh/shells"
)

const uninstallDescription = `Uninstall remote program`
const uninstallHelp = `
    uninstall <programname>

uninstall will remote the target repository from target-host.

'programnames' can be a single program name or list of program names separated
by white-space.
`

type UninstallCommand struct{}

func (cmd *UninstallCommand) Name() string {
	return "uninstall"
}

func (cmd *UninstallCommand) Description() string {
	return uninstallDescription
}

func (cmd *UninstallCommand) Help() string {
	return uninstallHelp
}

func (cmd *UninstallCommand) Shells() []string {
	return []string{api.SHELL_INDEX}
}

func (cmd *UninstallCommand) Complete(c *api.Context, cursor int) []string {
	return []string{}
}

func (cmd *UninstallCommand) argParse(args []string) *flag.FlagSet {
	fl := flag.NewFlagSet("uninstall", flag.ContinueOnError)
	fl.Parse(args)
	return fl
}

func (cmd *UninstallCommand) Interpret(c *api.Context) (err error) {
	if idx, ok := c.Cursh.(*shells.Indexsh); ok {
		err = cmd.uninstallForIndex(idx, c)
	} else {
		err = fmt.Errorf("Shell not supported")
	}
	return
}

func (cmd *UninstallCommand) uninstallForIndex(idx *shells.Indexsh, c *api.Context) (err error) {
	args, _ := api.ParseCmdline(c.Line)
	for _, progname := range args[1:] {
		idx.Printch <- fmt.Sprintf("** Uninstalling %v ...\n", progname)
		err = idx.Fabric.UninstallProgram(progname, idx.Printch)
	}
	return
}

func init() {
	knownCommands["uninstall"] = &UninstallCommand{}
}
