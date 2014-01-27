package commands

import (
	"flag"
	"fmt"
	"github.com/couchbaselabs/cbsh/api"
	"github.com/couchbaselabs/cbsh/shells"
)

var ppCbshOption struct {
	pool   bool
	bucket bool
}

var ppDescription = `Pretty print json documents and internal data structure`
var ppHelp = `
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
	return ppHelp
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

func (cmd *PpCommand) cbshArgParse(line string) (err error) {
	f := flag.NewFlagSet("ppcbsh", flag.ContinueOnError)
	f.BoolVar(&ppCbshOption.pool, "pool", false,
		"Pretty print current pool details")
	f.BoolVar(&ppCbshOption.bucket, "bucket", false,
		"Pretty print current bucket details")
	return f.Parse(api.ParseCmd(line)[1:])
}

func (cmd *PpCommand) indexArgParse(line string) (err error) {
	f := flag.NewFlagSet("ppindex", flag.ContinueOnError)
	return f.Parse(api.ParseCmd(line)[1:])
}

func (cmd *PpCommand) n1qlArgParse(line string) (err error) {
	f := flag.NewFlagSet("ppn1ql", flag.ContinueOnError)
	return f.Parse(api.ParseCmd(line)[1:])
}

func (cmd *PpCommand) ppForCbsh(cbsh *shells.Cbsh, c *api.Context) (err error) {
	var s string

	cmd.indexArgParse(c.Line)

	switch {
	case ppCbshOption.bucket:
		s, err = api.PrettyPrint(*cbsh.Bucket, "")
		fmt.Fprintln(c.W, s)
	case ppCbshOption.pool:
		s, err = api.PrettyPrint(cbsh.Pool, "")
		fmt.Fprintln(c.W, s)
	}
	return
}

func (cmd *PpCommand) ppForIndex(index *shells.Indexsh,
	c *api.Context) (err error) {
	return
}

func (cmd *PpCommand) ppForN1ql(n1ql *shells.N1qlsh,
	c *api.Context) (err error) {
	return
}

func init() {
	knownCommands["pp"] = &PpCommand{}
}
