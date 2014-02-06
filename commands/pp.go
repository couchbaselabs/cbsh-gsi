package commands

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/couchbaselabs/cbsh/api"
	"github.com/couchbaselabs/cbsh/shells"
)

type ppOptions struct {
	pool   bool
	bucket bool
}

const ppDescription = `Pretty print json documents and internal data structure`
const ppHelp = `
for Cbsh shell:
    pp [-pool] [-bucket]

pretty prints is based on the shell in which it is invoked.
`

type PpCommand struct{}

func (cmd *PpCommand) Name() string {
	return "pp"
}

func (cmd *PpCommand) Description() string {
	return ppDescription
}

func (cmd *PpCommand) Help() string {
	var fl *flag.FlagSet
	options := ppOptions{}
	fl = cmd.argParse(&options, []string{})
	buf := bytes.NewBuffer([]byte{})
	fl.SetOutput(buf)
	fl.PrintDefaults()
	return ppHelp + string(buf.Bytes())
}

func (cmd *PpCommand) Shells() []string {
	return []string{api.SHELL_CB, api.SHELL_INDEX, api.SHELL_N1QL}
}

func (cmd *PpCommand) Complete(c *api.Context, cursor int) []string {
	return []string{}
}

func (cmd *PpCommand) Interpret(c *api.Context) (err error) {
	if cbsh, ok := c.Cursh.(*shells.Cbsh); ok {
		cmd.ppForCbsh(cbsh, c)
	} else if index, ok := c.Cursh.(*shells.Indexsh); ok {
		cmd.ppForIndex(index, c)
	} else if n1ql, ok := c.Cursh.(*shells.N1qlsh); ok {
		cmd.ppForN1ql(n1ql, c)
	} else {
		err = fmt.Errorf("Shell not supported")
	}
	return
}

func (cmd *PpCommand) argParse(options *ppOptions, args []string) *flag.FlagSet {
	fl := flag.NewFlagSet("ppcbsh", flag.ContinueOnError)
	fl.BoolVar(&options.pool, "pool", false,
		"Pretty print current pool details")
	fl.BoolVar(&options.bucket, "bucket", false,
		"Pretty print current bucket details")
	fl.Parse(args)
	return fl
}

func (cmd *PpCommand) ppForCbsh(cbsh *shells.Cbsh, c *api.Context) (err error) {
	var s string

	args, _ := api.ParseCmdline(c.Line)
	options := ppOptions{}
	cmd.argParse(&options, args[1:])

	switch {
	case options.bucket:
		s, err = api.PrettyPrint(*cbsh.Bucket, "")
		fmt.Fprintln(c.W, s)
	case options.pool:
		s, err = api.PrettyPrint(cbsh.Pool, "")
		fmt.Fprintln(c.W, s)
	}
	return
}

func (cmd *PpCommand) ppForIndex(index *shells.Indexsh, c *api.Context) (err error) {
	return
}

func (cmd *PpCommand) ppForN1ql(n1ql *shells.N1qlsh, c *api.Context) (err error) {
	return
}

func init() {
	knownCommands["pp"] = &PpCommand{}
}
