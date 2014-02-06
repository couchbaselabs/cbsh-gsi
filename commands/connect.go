package commands

import (
	"fmt"
	"github.com/couchbaselabs/cbsh/api"
	"github.com/couchbaselabs/cbsh/shells"
)

const connectDescription = `Connect with kv-cluster`
const connectHelp = `
    connect <url> [poolname] [bucketname]

connect to a server, specified by <url> in kv-cluster. If optional argument
[poolname] is supplied, change to pool. If optional argument
[bucketname] is supplied, change to bucket.
`
const defaultPool = "default"
const defaultBucket = "default"

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

// Local functions

func connectForCbsh(cbsh *shells.Cbsh, c *api.Context) (err error) {
	// Close existing client connection
	if cbsh.Bucket != nil {
		cbsh.Bucket.Close()
	}

	toks, _ := api.ParseCmdline(c.Line)
	if len(toks) < 2 {
		return fmt.Errorf("Need argument to connect")
	} else {
		cbsh.Url = toks[1]
	}

	cbsh.Poolname = defaultPool
	if len(toks) > 2 {
		cbsh.Poolname = toks[2]
	}

	cbsh.Bucketname = defaultBucket
	if len(toks) > 3 {
		cbsh.Bucketname = toks[3]
	}

	if err = cbsh.Connect(c); err == nil {
		fmt.Fprintln(c.W, "Connected to :", cbsh.Prompt())
	}
	return
}

func init() {
	knownCommands["connect"] = &ConnectCommand{}
}
