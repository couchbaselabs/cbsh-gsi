package shells

import (
	"flag"
	"fmt"
	"github.com/couchbaselabs/cbsh/api"
	"github.com/couchbaselabs/go-couchbase"
	"net/url"
	"path"
)

var cbDescription = `Shell to connect and interface with couchbase kv-cluster`

// Global structure that maintains the current state of the couchbase-shell
type Cbsh struct {
	Client     couchbase.Client  // current couchbase client
	Poolname   string            // name of the current active pool
	Pool       couchbase.Pool    // current active pool
	Bucketname string            // name of the current active bucket
	Bucket     *couchbase.Bucket // current active bucket
	Url        string
	U          *url.URL // url used to connect to the server
	CommandList
}

func (cbsh *Cbsh) Description() string {
	return cbDescription
}

func (cbsh *Cbsh) Init(c *api.Context, commands api.CommandMap) (err error) {
	api.CreateFile(cbsh.HistoryFile(), false)
	cbsh.Commands = commands
	// Url, Poolname, Bucketname are already initialized with ArgParse
	err = cbsh.Connect(c)
	return
}

func (cbsh *Cbsh) HistoryFile() string {
	datadir := api.ShellDatadir()
	return path.Join(datadir, fmt.Sprintf(api.HISTORY_FILE_TMPL, api.SHELL_CB))
}

func (cbsh *Cbsh) ArgParse() {
	flag.StringVar(&cbsh.Poolname, "pool", "default",
		"pool to connect to (defaults to `default`)")
	flag.StringVar(&cbsh.Bucketname, "bucket", "default",
		"bucket to connect to (defaults to `default`)")
	flag.StringVar(&cbsh.Url, "url", "http://localhost:8091",
		"select the server to connect")
}

func (cbsh *Cbsh) Name() string {
	return api.SHELL_CB
}

func (cbsh *Cbsh) Prompt() string {
	if cbsh.U == nil {
		return "cbsh"
	} else {
		return cbsh.U.Host + "/" + cbsh.Poolname + "/" + cbsh.Bucketname
	}
}

func (cbsh *Cbsh) Handle(c *api.Context) (err error) {
	return
}

func (cbsh *Cbsh) Close(c *api.Context) {
	fmt.Fprintf(c.W, "Exiting shell : %v\n", cbsh.Name())
}

func (cbsh *Cbsh) Connect(c *api.Context) (err error) {
	// Connect to client
	if cbsh.U, err = url.Parse(cbsh.Url); err != nil {
		cbsh.U, cbsh.Bucket = nil, nil
		return
	}
	if cbsh.Client, err = couchbase.Connect(cbsh.U.String()); err != nil {
		cbsh.U, cbsh.Bucket = nil, nil
		return
	}
	// Connect to pool
	if cbsh.Pool, err = cbsh.Client.GetPool(cbsh.Poolname); err != nil {
		cbsh.U, cbsh.Bucket = nil, nil
		return
	}
	// Connect to bucket
	if cbsh.Bucket, err = cbsh.Pool.GetBucket(cbsh.Bucketname); err != nil {
		cbsh.U, cbsh.Bucket = nil, nil
		return
	}
	return
}

func init() {
	knownShells[api.SHELL_CB] = &Cbsh{}
}
