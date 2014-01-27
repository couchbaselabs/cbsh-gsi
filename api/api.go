package api

import (
	"fmt"
	"github.com/prataprc/liner"
	"io"
)

const (
	SHELL_CB    = "cbsh"
	SHELL_N1QL  = "n1ql"
	SHELL_INDEX = "index"
)
const CBSH_DIR = "./.cbsh"
const HISTORY_FILE_TMPL = "./%s_history"

const NEWLINE = byte(10) // '\n'
const SEP = "\n"

type CommandMap map[string]CommandHandler

type Context struct {
	Cursh    ShellHandler // current active shell
	Liner    *liner.State
	Line     string    // current line
	W        io.Writer // output for this application
	Shells   map[string]ShellHandler
	Commands CommandMap
}

// Interface to be implemented by individual shells
type ShellHandler interface {
	// Name returns name of the shell implementing this interface.
	Name() string

	// Description returns one line description of the shell.
	Description() string

	// Init initializes the shell. Init needs to be called when ever a shell is
	// activated.
	Init(*Context, CommandMap) error

	// HistoryFile return filename to store history log. Shells can have
	// individual history log under CBSH_DIR (typically $HOME/.cbsh/)
	HistoryFile() string

	// ArgParse parses command line arguments for the shell.
	ArgParse()

	// Prompt return the prompt for the shell
	Prompt() string

	// GetCommand returns the handler for command entered via shell.
	GetCommand(string) CommandHandler

	// Handle will handle the command.
	Handle(*Context) error // handle the shell command

	// Close closes the shell. Close needs to be called when ever a shell is
	// deactivated.
	Close(*Context)
}

// Interface to handle individual commands
type CommandHandler interface {
	// Name returns the name of this command.
	Name() string

	// Description return one line description of the command.
	Description() string

	// Help returns the long description of the command
	Help() string

	// Shell returns a list of shell-names in which this command is supported.
	Shells() []string

	// Complete can be used for tab completion, returns a list of possible
	// completion at the cursor.
	Complete(c *Context, cursor int) []string

	// Interpret the shell command.
	Interpret(*Context) error
}

func (c *Context) SetShell(shell ShellHandler) (err error) {
	fmt.Fprintf(c.W, "Using %q shell: %v\n", shell.Name(), shell.Description())
	c.Cursh = shell
	return c.Cursh.Init(c, c.ShellCommands(c.Cursh.Name()))
}

func (c *Context) ShellCommands(shname string) CommandMap {
	commands := make(CommandMap)
	for cname, command := range c.Commands {
		for _, sname := range command.Shells() {
			if sname == shname {
				commands[cname] = command
			}
		}
	}
	return commands
}

func (c *Context) Close() {
	c.Liner.Close()
	c.Cursh.Close(c)
}
