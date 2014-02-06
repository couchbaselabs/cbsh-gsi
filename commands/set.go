package commands

import (
	"fmt"
	"github.com/couchbaselabs/cbsh/api"
	"github.com/couchbaselabs/cbsh/shells"
	"strconv"
)

const setDescription = `Set key,value from current bucket`
const setHelp = `
    set <key> <expiry> <value>

set the <value> for <key> into current bucket. <value> is expected as string
and <expiry> is expected as integer
`

type SetCommand struct{}

func (cmd *SetCommand) Name() string {
	return "set"
}

func (cmd *SetCommand) Description() string {
	return setDescription
}

func (cmd *SetCommand) Help() string {
	return setHelp
}

func (cmd *SetCommand) Shells() []string {
	return []string{api.SHELL_CB}
}

func (cmd *SetCommand) Complete(c *api.Context, cursor int) []string {
	return []string{}
}

func (cmd *SetCommand) Interpret(c *api.Context) (err error) {
	if cbsh, ok := c.Cursh.(*shells.Cbsh); ok {
		setForCbsh(cbsh, c)
	} else {
		err = fmt.Errorf("Shell not supported")
	}
	return
}

func setForCbsh(cbsh *shells.Cbsh, c *api.Context) (err error) {
	toks := api.CommandLineTokens(c.Line)
	if cbsh.Bucket == nil {
		err = fmt.Errorf("Not connected to bucket")
	} else if len(toks) < 4 {
		err = fmt.Errorf("Not argument to set")
	} else {
		var expiry int
		if expiry, err = strconv.Atoi(toks[2]); err == nil {
			err = cbsh.Bucket.Set(toks[1], expiry, toks[3])
		}
	}
	return
}

func init() {
	knownCommands["set"] = &SetCommand{}
}
