package shells

import (
	"fmt"
	"github.com/couchbaselabs/cbsh/api"
	"path"
)

var n1qlDescription = `N1QL query shell`

// Global structure that maintains the current state of the index-shell
type N1qlsh struct {
	CommandList // commands loaded for this shell
}

func (n1ql *N1qlsh) Description() string {
	return n1qlDescription
}

func (n1ql *N1qlsh) Init(c *api.Context, commands api.CommandMap) (err error) {
	api.CreateFile(n1ql.HistoryFile(), false)
	return
}

func (n1ql *N1qlsh) HistoryFile() string {
	datadir := api.ShellDatadir()
	return path.Join(datadir, fmt.Sprintf(api.HISTORY_FILE_TMPL, api.SHELL_N1QL))
}

func (n1ql *N1qlsh) ArgParse() {
	return
}

func (n1ql *N1qlsh) Name() string {
	return ""
}

func (n1ql *N1qlsh) Prompt() string {
	return ""
}

func (n1ql *N1qlsh) Handle(c *api.Context) (err error) {
	return
}

func (n1ql *N1qlsh) Close(c *api.Context) {
	fmt.Fprintf(c.W, "Exiting shell : %v\n", n1ql.Name())
}

func init() {
	knownShells[api.SHELL_N1QL] = &N1qlsh{}
}
