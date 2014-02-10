package commands

import (
	"flag"
	"fmt"
	"github.com/couchbaselabs/cbsh/api"
	"github.com/couchbaselabs/cbsh/shells"
	//"github.com/couchbaselabs/cbsh/sshc"
)

var runDescription = `Execute configuration for seconday index cluster`
var runHelp = `
    run [-c configfile] [-i] [-if] <programnames>

run specified programs. 'programnames' can be a single program name or list of
program names separated by white-space.
`

type RunCommand struct{}

type runOptions struct {
	configfile   string
	install      bool
	forceinstall bool
	programs     []string
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
		args, _ := api.ParseCmdline(c.Line)
		options := runOptions{}
		fl := cmd.argParse(&options, args[1:])
		options.programs = fl.Args()
		err = runForIndex(idx, &options, c)
	} else {
		err = fmt.Errorf("Error: need to be in index-shell")
	}
	return
}

// Local functions

func (cmd *RunCommand) argParse(options *runOptions, args []string) *flag.FlagSet {
	fl := flag.NewFlagSet("run", flag.ContinueOnError)
	fl.StringVar(&options.configfile, "c", "",
		"load configuration file before installing or running the program")
	fl.BoolVar(&options.install, "i", false,
		"install programs before running them")
	fl.BoolVar(&options.forceinstall, "if", false,
		"force install programs before running them")
	fl.Parse(args)
	return fl
}

func runForIndex(idx *shells.Indexsh, options *runOptions, c *api.Context) (err error) {
	switch {
	case options.configfile != "":
		if err = configForIndex(idx, c, options.configfile); err == nil {
			runPrograms(idx, c, options.programs)
		}
	case idx.Config != nil:
		opts := installOptions{programs: options.programs}
		if options.forceinstall {
			opts.force = options.forceinstall
		}
		if options.forceinstall || options.install {
			installForIndex(idx, &opts, c)
		}
		runPrograms(idx, c, options.programs)
	default:
		return fmt.Errorf("Configuration file not loaded")
	}
	return
}

func runPrograms(idx *shells.Indexsh, c *api.Context, programs []string) (err error) {
	for _, name := range programs {
		idx.Fabric.KillProgram(name)
		idx.Fabric.RunProgram(name, idx.Printch)
	}
	return
}

func init() {
	knownCommands["run"] = &RunCommand{}
}
