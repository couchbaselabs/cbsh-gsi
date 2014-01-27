package commands

import (
	"fmt"
	"github.com/couchbaselabs/cbsh/api"
	"github.com/couchbaselabs/cbsh/shells"
)

var bucketDescription = `Choose a bucket as current bucket`
var bucketHelp = `
`

type BucketCommand struct{}

func (cmd *BucketCommand) Name() string {
	return "bucket"
}

func (cmd *BucketCommand) Description() string {
	return bucketDescription
}

func (cmd *BucketCommand) Help() string {
	return bucketHelp
}

func (cmd *BucketCommand) Shells() []string {
	return []string{api.SHELL_CB}
}

func (cmd *BucketCommand) Complete(c *api.Context, cursor int) []string {
	return []string{}
}

func (cmd *BucketCommand) Interpret(c *api.Context) (err error) {
	if cbsh, ok := c.Cursh.(*shells.Cbsh); ok {
		bucketForCbsh(cbsh, c)
	} else {
		err = fmt.Errorf("Shell not supported")
	}
	return
}

func bucketForCbsh(cbsh *shells.Cbsh, c *api.Context) (err error) {
	if cbsh.Bucket != nil {
		cbsh.Bucket.Close()
	}
	parts := api.SplitArgs(c.Line, " ")
	if len(parts) < 2 {
		err = fmt.Errorf("Need argument to bucket")
	} else if cbsh.U == nil {
		err = fmt.Errorf("Not connected to any server")
	} else if bucket, err := cbsh.Pool.GetBucket(parts[1]); err == nil {
		cbsh.Bucketname, cbsh.Bucket = parts[1], bucket
	}
	return
}

func init() {
	knownCommands["bucket"] = &BucketCommand{}
}
