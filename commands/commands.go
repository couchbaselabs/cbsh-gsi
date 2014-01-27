package commands

import (
	"fmt"
	"github.com/couchbaselabs/cbsh/api"
)

var commandsDescription = `Short description of all commands for this shell`
var commandsHelp = `
`

type CommandsCommand struct{}

func (cmd *CommandsCommand) Name() string {
	return "commands"
}

func (cmd *CommandsCommand) Description() string {
	return commandsDescription
}

func (cmd *CommandsCommand) Help() string {
	return commandsHelp
}

func (cmd *CommandsCommand) Shells() []string {
	return []string{api.SHELL_CB, api.SHELL_INDEX}
}

func (cmd *CommandsCommand) Complete(c *api.Context, cursor int) []string {
	return []string{}
}

func (cmd *CommandsCommand) Interpret(c *api.Context) (err error) {
	for name, cmd := range c.ShellCommands(c.Cursh.Name()) {
		fmt.Fprintf(c.W, "  %-15v %v\n", name, cmd.Description())
	}
	return
}

func init() {
	knownCommands["commands"] = &CommandsCommand{}
}
