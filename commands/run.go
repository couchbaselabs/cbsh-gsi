package commands

import (
	"flag"
	"fmt"
	"github.com/couchbaselabs/cbsh/api"
	"github.com/couchbaselabs/cbsh/shells"
	"github.com/couchbaselabs/cbsh/sshc"
)

var runDescription = `Execute configuration for seconday index cluster`
var runHelp = `
    run [-c <config-file>]

if [-c <config-file>] option is supplied, execute the specified configuration
file as seconday index cluster, otherwise execute the configuration that was
already selected using "config" command. In either case, if the cluster is
already running a configuration, it is killed and the new configuration is
executed.
`

type RunCommand struct{}

var runOption struct {
	configFile string
	program    string
}

func (cmd *RunCommand) Name() string {
	return "run"
}

func (cmd *RunCommand) Description() string {
	return runDescription
}

func (cmd *RunCommand) Help() string {
	return runHelp
}

func (cmd *RunCommand) Shells() []string {
	return []string{api.SHELL_INDEX}
}

func (cmd *RunCommand) Complete(c *api.Context, cursor int) []string {
	return []string{}
}

func (cmd *RunCommand) Interpret(c *api.Context) (err error) {
	if idx, ok := c.Cursh.(*shells.Indexsh); ok {
		err = cmd.runForIndex(idx, c)
	} else {
		err = fmt.Errorf("Error: need to be in index-shell")
	}
	return
}

func (cmd *RunCommand) indexArgParse(line string) (err error) {
	f := flag.NewFlagSet("run", flag.ContinueOnError)
	f.StringVar(&runOption.configFile, "c", "",
		"Specify configuration file")
	f.StringVar(&runOption.program, "p", "",
		"Specify program to run or restart")
	err = f.Parse(api.ParseScript(line)[0][1:])
	return
}

func (cmd *RunCommand) runForIndex(idx *shells.Indexsh, c *api.Context) (err error) {
	cmd.indexArgParse(c.Line)
	switch {
	case runOption.configFile != "":
		if err = configForIndex(idx, c, runOption.configFile); err == nil {
			runConfig(idx, c)
		}
	case idx.Config != nil && runOption.program != "":
		for name, _ := range idx.Config["programs"].(map[string]interface{}) {
			if name == runOption.program {
				if p := idx.Programs[name]; p != nil {
					p.Kill()
				}
				runProgram(idx, c, name)
				break
			}
		}
	case idx.Config != nil:
		runConfig(idx, c)
	}
	return
}

func runConfig(idx *shells.Indexsh, c *api.Context) (err error) {
	progConfigs := idx.Config["programs"].(map[string]interface{})
	for name, _ := range progConfigs {
		runProgram(idx, c, name)
	}
	return
}

func runProgram(idx *shells.Indexsh, c *api.Context, name string) {
	p := sshc.RunProgram(name, idx.Config)
	idx.Programs[name] = p
	go idx.GetLog(p, c)
}

func init() {
	knownCommands["run"] = &RunCommand{}
}
