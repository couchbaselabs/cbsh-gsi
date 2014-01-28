package shells

import (
	"flag"
	"fmt"
	"github.com/couchbaselabs/cbsh/api"
	"github.com/couchbaselabs/cbsh/sshc"
	"path"
	"sync"
)

var idxDescription = `Shell to handle secondary index cluster`

// Global structure that maintains the current state of the index-shell
type Indexsh struct {
	ConfigFile  string                 // path to configuration file
	Config      map[string]interface{} // configuration
	CommandList                        // commands loaded for this shell
	mu          sync.Mutex
	Programs    map[string]*sshc.Program // List of running programs
	printch     chan string
	quit        chan bool
}

func (idx *Indexsh) Description() string {
	return idxDescription
}

func (idx *Indexsh) Init(c *api.Context, commands api.CommandMap) (err error) {
	api.CreateFile(idx.HistoryFile(), false)
	idx.Config = nil
	idx.Programs = make(map[string]*sshc.Program)
	idx.Commands = commands
	idx.printch = make(chan string)
	idx.quit = make(chan bool)
	go idx.Print(c)
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
	idx.Killall(c)
	close(idx.quit)
	fmt.Fprintf(c.W, "Exiting shell : %v\n", idx.Name())
}

func (idx *Indexsh) Killall(c *api.Context) bool {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	if idx.Programs == nil {
		return false
	}
	for _, p := range idx.Programs {
		p.Kill()
	}
	idx.Programs = make(map[string]*sshc.Program)
	return true
}

func (idx *Indexsh) GetLog(p *sshc.Program, c *api.Context) {
loop:
	for {
		select {
		case s, ok := <-p.Outch:
			if !ok {
				break loop
			}
			idx.printch <- s
		case s, ok := <-p.Errch:
			if !ok {
				break loop
			}
			idx.printch <- s
		case <-idx.quit:
			break loop
		}
	}
	idx.mu.Lock()
	defer idx.mu.Unlock()
	delete(idx.Programs, p.Name)
}

func (idx *Indexsh) Print(c *api.Context) {
loop:
	for {
		select {
		case s, ok := <-idx.printch:
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
