package commands

import (
	"fmt"
	"github.com/couchbaselabs/cbsh/api"
	"github.com/couchbaselabs/cbsh/shells"
)

var getDescription = `Get key,value from current bucket`
var getHelp = `
    get <key>

get the value for <key> from current bucket.
`

type GetCommand struct{}

func (cmd *GetCommand) Name() string {
	return "get"
}

func (cmd *GetCommand) Description() string {
	return getDescription
}

func (cmd *GetCommand) Help() string {
	return getHelp
}

func (cmd *GetCommand) Shells() []string {
	return []string{api.SHELL_CB}
}

func (cmd *GetCommand) Complete(c *api.Context, cursor int) []string {
	return []string{}
}

func (cmd *GetCommand) Interpret(c *api.Context) (err error) {
	if cbsh, ok := c.Cursh.(*shells.Cbsh); ok {
		getForCbsh(cbsh, c)
	} else {
		err = fmt.Errorf("Shell not supported")
	}
	return
}

func getForCbsh(cbsh *shells.Cbsh, c *api.Context) (err error) {
	var ob interface{}
	parts := api.SplitArgs(c.Line, " ")
	if cbsh.Bucket == nil {
		err = fmt.Errorf("Not connected to bucket")
	} else if len(parts) < 2 {
		err = fmt.Errorf("Need argument to get")
	} else if err = cbsh.Bucket.Get(parts[1], &ob); err == nil {
		var s string
		s, err = api.PrettyPrint(ob, "")
		fmt.Fprintf(c.W, "%v\n", s)
	}
	return
}

func init() {
	knownCommands["get"] = &GetCommand{}
}
