package commands

import (
	"fmt"
	"github.com/couchbaselabs/cbsh/api"
	"github.com/couchbaselabs/cbsh/shells"
)

const listDescription = `List interesting information about the cluster`
const listHelp = `
    list nodes | pools | buckets
`

type ListCommand struct{}

func (cmd *ListCommand) Name() string {
	return "list"
}

func (cmd *ListCommand) Description() string {
	return listDescription
}

func (cmd *ListCommand) Help() string {
	return listHelp
}

func (cmd *ListCommand) Shells() []string {
	return []string{api.SHELL_CB}
}

func (cmd *ListCommand) Complete(c *api.Context, cursor int) []string {
	return []string{}
}

func (cmd *ListCommand) Interpret(c *api.Context) (err error) {
	if cbsh, ok := c.Cursh.(*shells.Cbsh); ok {
		listForCbsh(cbsh, c)
	} else {
		err = fmt.Errorf("Shell not supported")
	}
	return
}

func listForCbsh(cbsh *shells.Cbsh, c *api.Context) (err error) {
	var s string

	toks := api.CommandLineTokens(c.Line)
	if len(toks) < 2 {
		return fmt.Errorf("Insufficient argument to command")
	}

	switch toks[1] {
	case "nodes":
		if cbsh.Bucket != nil {
			nodes := make([]string, 0)
			for _, node := range cbsh.Bucket.Nodes() {
				nodes = append(nodes, node.Hostname)
			}
			s, err = api.PrettyPrint(nodes, "")
			fmt.Fprintf(c.W, "%v\n", s)
		} else {
			err = fmt.Errorf("Bucket not initialized")
		}
	case "pools":
		if cbsh.U != nil {
			pools := make([]string, 0)
			for _, restPool := range cbsh.Client.Info.Pools {
				pools = append(pools, restPool.Name)
			}
			s, err = api.PrettyPrint(pools, "")
			fmt.Fprintf(c.W, "%v\n", s)
		} else {
			err = fmt.Errorf("Pool not initialized")
		}
	case "buckets":
		if cbsh.U != nil {
			buckets := make([]string, 0)
			for k, _ := range cbsh.Pool.BucketMap {
				buckets = append(buckets, k)
			}
			s, err = api.PrettyPrint(buckets, "")
			fmt.Fprintf(c.W, "%v\n", s)
		} else {
			err = fmt.Errorf("Pool not initialized")
		}
	}
	return
}

func init() {
	knownCommands["list"] = &ListCommand{}
}
