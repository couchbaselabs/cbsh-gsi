package commands

import (
	"flag"
	"fmt"
	"github.com/couchbaselabs/cbsh/api"
	"github.com/couchbaselabs/cbsh/shells"
)

const installDescription = `Install remote program`
const installHelp = `
    install -f <programnames>

install one or more programs, typically this involved, cloning the repository
patching it with un-commited changes from the source repository and compiling
relevant portions of target repository. Refer to configuration spec. for more
details. 'programnames' can be a single program name or list of program names
separated by white-space.
`

type InstallCommand struct{}

type installOptions struct {
	force    bool
	programs []string
}

func (cmd *InstallCommand) Name() string {
	return "install"
}

func (cmd *InstallCommand) Description() string {
	return installDescription
}

func (cmd *InstallCommand) Help() string {
	return installHelp
}

func (cmd *InstallCommand) Shells() []string {
	return []string{api.SHELL_INDEX}
}

func (cmd *InstallCommand) Complete(c *api.Context, cursor int) []string {
	return []string{}
}

func (cmd *InstallCommand) Interpret(c *api.Context) (err error) {
	if idx, ok := c.Cursh.(*shells.Indexsh); ok {
		args, _ := api.ParseCmdline(c.Line)
		options := installOptions{}
		fl := cmd.argParse(&options, args[1:])
		options.programs = fl.Args()
		err = installForIndex(idx, &options, c)
	} else {
		err = fmt.Errorf("Shell not supported")
	}
	return
}

// Local functions

func (cmd *InstallCommand) argParse(options *installOptions, args []string) *flag.FlagSet {
	fl := flag.NewFlagSet("install", flag.ContinueOnError)
	fl.BoolVar(&options.force, "f", false,
		"force install programs")
	fl.Parse(args)
	return fl
}

func installForIndex(
	idx *shells.Indexsh, options *installOptions, c *api.Context) (err error) {

	for _, progname := range options.programs {
		idx.Printch <- fmt.Sprintf("** Installing %v ...\n", progname)
		host := idx.Fabric.Config.TargetHost(progname)
		dir := idx.Fabric.Config.TargetRoot(progname)
		user := idx.Fabric.Config.User(progname)
		if dir != "" {
			err = idx.Fabric.MakeRemoteDirs(host, user, dir, idx.Printch)
		}
		if err != nil {
			return
		}
		err = idx.Fabric.InstallProgram(progname, idx.Printch, options.force)
	}
	return
}

func init() {
	knownCommands["install"] = &InstallCommand{}
}
