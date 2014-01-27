package commands

import (
	"fmt"
	"github.com/couchbaselabs/cbsh/api"
	"strings"
)

var helpDescription = `Detailed help on individual commands`
var helpHelp = `
    help [command-name]

show short description of command, long description of it and shells in which
the command is supported. If [command-name] is not supplied, list all commands
with its short description.
`

type HelpCommand struct{}

func (cmd *HelpCommand) Name() string {
	return "help"
}

func (cmd *HelpCommand) Description() string {
	return helpDescription
}

func (cmd *HelpCommand) Help() string {
	return helpHelp
}

func (cmd *HelpCommand) Shells() []string {
	return []string{api.SHELL_CB, api.SHELL_N1QL, api.SHELL_INDEX}
}

func (cmd *HelpCommand) Complete(c *api.Context, cursor int) []string {
	return []string{}
}

func (cmd *HelpCommand) Interpret(c *api.Context) (err error) {
	parts := api.SplitArgs(c.Line, " ")
	if len(parts) < 2 {
		for _, cmd := range c.Commands {
			fmt.Fprintf(c.W, "%-15s %s\n", cmd.Name(), cmd.Description())
		}
	} else {
		cmd := c.Commands[parts[1]]
		fmt.Fprintln(c.W, api.Yellow(cmd.Description()))
		help := cmd.Help()
		if len(strings.Trim(help, " \t\n\r")) > 0 {
			fmt.Fprintln(c.W, help)
		}
		fmt.Fprintf(c.W, "shells: %v\n\n", cmd.Shells())
	}
	return
}

func init() {
	knownCommands["help"] = &HelpCommand{}
}
