package commands

import (
	"fmt"
	"github.com/couchbaselabs/cbsh/api"
	"github.com/couchbaselabs/cbsh/shells"
)

var connectDescription = `Connect with kv-cluster`
var connectHelp = `
    connect <url> [poolname] [bucketname]

connect to a server, specified by <url> in kv-cluster. If optional argument
[poolname] is supplied, change to pool. If optional argument
[bucketname] is supplied, change to bucket.
`

type ConnectCommand struct{}

func (cmd *ConnectCommand) Name() string {
	return "connect"
}

func (cmd *ConnectCommand) Description() string {
	return connectDescription
}

func (cmd *ConnectCommand) Help() string {
	return connectHelp
}

func (cmd *ConnectCommand) Shells() []string {
	return []string{api.SHELL_CB}
}

func (cmd *ConnectCommand) Complete(c *api.Context, cursor int) []string {
	return []string{}
}

func (cmd *ConnectCommand) Interpret(c *api.Context) (err error) {
	if cbsh, ok := c.Cursh.(*shells.Cbsh); ok {
		err = connectForCbsh(cbsh, c)
	} else {
		err = fmt.Errorf("Shell not supported")
	}
	return
}

func connectForCbsh(cbsh *shells.Cbsh, c *api.Context) (err error) {
	// Close existing client connection
	if cbsh.Bucket != nil {
		cbsh.Bucket.Close()
	}

	parts := api.SplitArgs(c.Line, " ")
	if len(parts) < 2 {
		return fmt.Errorf("Need argument to connect")
	} else {
		cbsh.Url = parts[1]
	}
	if len(parts) > 2 {
		cbsh.Poolname = parts[2]
	}
	if len(parts) > 3 {
		cbsh.Bucketname = parts[3]
	}
	return cbsh.Connect(c)
}

func init() {
	knownCommands["connect"] = &ConnectCommand{}
}
