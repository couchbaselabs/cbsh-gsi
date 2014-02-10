package commands

import (
	"fmt"
	"github.com/couchbaselabs/cbsh/api"
	"github.com/couchbaselabs/cbsh/shells"
)

const killDescription = `Kill remote program`
const killHelp = `
    kill <programnames>

kill remote programs. 'programnames' can be a single program name or list of
program names separated by white-space.
`

type KillCommand struct{}

var killOptions struct {
	programs []string
}

func (cmd *KillCommand) Name() string {
	return "kill"
}

func (cmd *KillCommand) Description() string {
	return killDescription
}

func (cmd *KillCommand) Help() string {
	return killHelp
}

func (cmd *KillCommand) Shells() []string {
	return []string{api.SHELL_INDEX}
}

func (cmd *KillCommand) Complete(c *api.Context, cursor int) []string {
	return []string{}
}

func (cmd *KillCommand) Interpret(c *api.Context) (err error) {
	if idx, ok := c.Cursh.(*shells.Indexsh); ok {
		programs, _ := api.ParseCmdline(c.Line)
		for _, name := range programs[1:] {
			idx.Fabric.KillProgram(name)
		}
	} else {
		err = fmt.Errorf("Shell not supported")
	}
	return
}

func init() {
	knownCommands["kill"] = &KillCommand{}
}
