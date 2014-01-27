package shells

import (
	"github.com/couchbaselabs/cbsh/api"
)

type CommandList struct {
	Commands api.CommandMap // commands loaded for this shell
}

func (cmd *CommandList) GetCommand(cmdname string) api.CommandHandler {
	return cmd.Commands[cmdname]
}

var knownShells = map[string]api.ShellHandler{}

func Allshells() map[string]api.ShellHandler {
	return knownShells
}
