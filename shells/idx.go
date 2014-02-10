package shells

import (
	"flag"
	"fmt"
	"github.com/couchbaselabs/cbsh/api"
	"github.com/couchbaselabs/cbsh/sshc"
	"path"
)

var idxDescription = `Shell to handle secondary index cluster`

// Global structure that maintains the current state of the index-shell
type Indexsh struct {
	ConfigFile  string     // path to configuration file
	Config      api.Config // configuration
	Fabric      *sshc.Fabric
	CommandList // commands loaded for this shell
	Printch     chan string
	quit        chan bool
}

func (idx *Indexsh) Description() string {
	return idxDescription
}

func (idx *Indexsh) Init(c *api.Context, commands api.CommandMap) (err error) {
	api.CreateFile(idx.HistoryFile(), false)
	idx.Config = nil
	idx.Commands = commands
	idx.Printch = make(chan string)
	idx.quit = make(chan bool)
	go idx.fabricPrinter(c)
	return
}

func (idx *Indexsh) HistoryFile() string {
	datadir := api.ShellDatadir()
	return path.Join(datadir, fmt.Sprintf(api.HISTORY_FILE_TMPL, api.SHELL_INDEX))
}

func (idx *Indexsh) ArgParse() {
	flag.StringVar(&idx.ConfigFile, "config", "",
		"specify the configuration file to load secondary index")
	return
}

func (idx *Indexsh) Name() string {
	return api.SHELL_INDEX
}

func (idx *Indexsh) Prompt() string {
	return "index:" + path.Base(idx.ConfigFile)
}

func (idx *Indexsh) Handle(c *api.Context) (err error) {
	return
}

func (idx *Indexsh) Close(c *api.Context) {
	if idx.Fabric.IsHealthy() {
		idx.Fabric.Killall()
	}
	close(idx.quit)
	fmt.Fprintf(c.W, "Exiting shell : %v\n", idx.Name())
}

func (idx *Indexsh) fabricPrinter(c *api.Context) {
loop:
	for {
		select {
		case s, ok := <-idx.Printch:
			if !ok {
				break loop
			}
			fmt.Fprintf(c.W, "%v", s)
		case <-idx.quit:
			break loop
		}
	}
}

func init() {
	knownShells[api.SHELL_INDEX] = &Indexsh{}
}
