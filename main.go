// A multi-shell environment to access couchbase systems like, the kv-cluster,
// secondary index and n1ql.
package main

import (
	"flag"
	"fmt"
	"github.com/couchbaselabs/cbsh/api"
	"github.com/couchbaselabs/cbsh/commands"
	"github.com/couchbaselabs/cbsh/shells"
	"github.com/prataprc/liner"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
)

var option struct {
	shell       string
	cmdstr      string
	interactive bool
}

func argParse(c *api.Context) {
	flag.StringVar(&option.shell, "shell", "cbsh", "Select the shell to use")
	flag.StringVar(&option.cmdstr, "cmd", "", "Command to execute")
	flag.BoolVar(&option.interactive, "i", false, "Enter interactive mode")

	for _, sh := range c.Shells {
		sh.ArgParse()
	}
	flag.Parse()
}

func main() {
	createDataDir()
	// Construct the context
	context := api.Context{
		W:        os.Stdout,
		Shells:   shells.Allshells(),
		Commands: commands.Allcommands(),
	}
	argParse(&context)
	if err := context.SetShell(context.Shells[option.shell]); err != nil {
		fmt.Fprintln(context.W, err)
	}
	context.Liner = liner.NewLiner()

	// Execute command if supplied through command-line
	if option.cmdstr != "" {
		if err := doCommand(&context, option.cmdstr); err != nil {
			fmt.Fprintln(context.W, err)
		}
	} else {
		option.interactive = true
	}

	go signalCatcher(&context)
	if option.interactive {
		interactiveLoop(&context)
	}
	(&context).Close()
}

func interactiveLoop(c *api.Context) {
	LoadHistory(c)
	for {
		prompt := c.Cursh.Prompt()
		if line, err := c.Liner.Prompt(prompt + "> "); err != nil {
			fmt.Fprintln(c.W, err)
			break
		} else if line == "" {
			continue
		} else if line == "q" {
			return
		} else {
			UpdateHistory(c, line)
			if err := doCommand(c, line); err != nil {
				fmt.Fprintln(c.W, err)
			}
		}
	}
}

func doCommand(c *api.Context, line string) (err error) {
	// Switch shells if need be
	switch {
	case strings.HasPrefix(line, api.SHELL_CB):
		err = c.SetShell(c.Shells[api.SHELL_CB])
	case strings.HasPrefix(line, api.SHELL_N1QL):
		err = c.SetShell(c.Shells[api.SHELL_N1QL])
	case strings.HasPrefix(line, api.SHELL_INDEX):
		err = c.SetShell(c.Shells[api.SHELL_INDEX])
	default:
		c.Line = line
		// Handle the command for the current shell
		err = handleShellCommand(c)
	}
	return
}

// Creates a data directory for cbsh shell. Data directory is expected to
// contain logs, history, configuration etc ...
func createDataDir() {
	if err := os.MkdirAll(api.ShellDatadir(), 0700); err != nil {
		panic(err)
	}
}

func handleShellCommand(c *api.Context) (err error) {
	shell := c.Cursh
	cmdname := api.SplitArgN(c.Line, " ", 2)[0]
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(c.W, "Recovered from %q: %v\n", cmdname, r)
			fmt.Println(string(debug.Stack()))
		}
	}()
	if !api.IsCommand(cmdname, c.ShellCommands(shell.Name())) {
		return fmt.Errorf("Invalid command: %v", cmdname)
	}
	if cmd := shell.GetCommand(cmdname); cmd == nil {
		err = fmt.Errorf(
			"Command %q not supported in %v", cmdname, api.SHELL_CB)
	} else {
		err = cmd.Interpret(c)
	}
	return
}

// Attempt to clean up after ctrl-C otherwise
// terminal is left in bad shape
func signalCatcher(c *api.Context) {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT)
	<-ch
	c.Close()
	os.Exit(0)
}
