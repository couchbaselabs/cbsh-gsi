package commands

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/couchbaselabs/cbsh/api"
	"github.com/couchbaselabs/cbsh/shells"
	"github.com/prataprc/go-couchbase"
	"strings"
)

var bucketDescription = `Choose bucket as current bucket`
var bucketHelp = `
    bucket [-nodes] <bucketname>

using this command will connect with <bucketname> and all kv operations
will correspond to this bucket.
`

type BucketCommand struct{}

type bucketOptions struct {
	nodes bool
}

func (cmd *BucketCommand) Name() string {
	return "bucket"
}

func (cmd *BucketCommand) Description() string {
	return bucketDescription
}

func (cmd *BucketCommand) Help() string {
	options := bucketOptions{}
	fl := cmd.argParse(&options, []string{})
	buf := bytes.NewBuffer([]byte{})
	fl.SetOutput(buf)
	fl.PrintDefaults()
	return bucketHelp + string(buf.Bytes())
}

func (cmd *BucketCommand) Shells() []string {
	return []string{api.SHELL_CB}
}

func (cmd *BucketCommand) Complete(c *api.Context, cursor int) []string {
	return []string{}
}

func (cmd *BucketCommand) Interpret(c *api.Context) (err error) {
	if cbsh, ok := c.Cursh.(*shells.Cbsh); ok {
		err = cmd.bucketForCbsh(cbsh, c)
	} else {
		err = fmt.Errorf("Shell not supported")
	}
	return
}

// Local functions

func (cmd *BucketCommand) argParse(options *bucketOptions, args []string) *flag.FlagSet {
	fl := flag.NewFlagSet("bucket", flag.ContinueOnError)
	fl.BoolVar(&options.nodes, "nodes", false,
		"list nodes in which the bucket is distributed")
	fl.Parse(args)
	return fl
}

func (cmd *BucketCommand) bucketForCbsh(cbsh *shells.Cbsh, c *api.Context) (err error) {
	var bucket *couchbase.Bucket
	// Close present bucket.
	if cbsh.Bucket != nil {
		cbsh.Bucket.Close()
	}

	args, _ := api.ParseCmdline(c.Line)
	options := bucketOptions{}
	fl := cmd.argParse(&options, args[1:])
	args = fl.Args()

	if cbsh.U == nil {
		err = fmt.Errorf("Not connected to any server")
	} else if len(args) > 0 {
		if bucket, err = cbsh.Pool.GetBucket(args[0]); err == nil {
			cbsh.Bucketname, cbsh.Bucket = args[0], bucket
		}
	}

	if err == nil && options.nodes {
		nodes := make(map[string]bool)
		servermap := cbsh.Bucket.VBServerMap()
		for vb, servers := range servermap.VBucketMap {
			ss := make([]string, 0, len(servers))
			for _, i := range servers {
				if i >= 0 {
					nodes[servermap.ServerList[i]] = true
					ss = append(ss, servermap.ServerList[i])
				}
			}
			fmt.Fprintf(c.W, "  %v: %v\n", vb, strings.Join(ss, " "))
		}
		ss := make([]string, 0, len(nodes))
		for node, _ := range nodes {
			ss = append(ss, node)
		}
		fmt.Fprintf(c.W, "list of nodes: %v\n", ss)
	}
	return
}

func init() {
	knownCommands["bucket"] = &BucketCommand{}
}
