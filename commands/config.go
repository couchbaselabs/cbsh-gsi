package commands

import (
	"encoding/json"
	"fmt"
	"github.com/couchbaselabs/cbsh/api"
	"github.com/couchbaselabs/cbsh/shells"
	"io/ioutil"
)

var configDescription = `Choose a configuration file for secondary index`
var configHelp = `
    config <config-file-in-json-format>

"config" command will only load the configuration file. Use "run" command to
launch the cluster. Any previous configuration that is executing in the
cluster will be killed before loading the new configuration.
`

type ConfigCommand struct{}

func (cmd *ConfigCommand) Name() string {
	return "config"
}

func (cmd *ConfigCommand) Description() string {
	return configDescription
}

func (cmd *ConfigCommand) Help() string {
	return configHelp
}

func (cmd *ConfigCommand) Shells() []string {
	return []string{api.SHELL_INDEX}
}

func (cmd *ConfigCommand) Complete(c *api.Context, cursor int) []string {
	return []string{}
}

func (cmd *ConfigCommand) Interpret(c *api.Context) (err error) {
	parts := api.SplitArgs(c.Line, " ")
	if len(parts) < 2 {
		err = fmt.Errorf("Specify a configuration file")
	} else if idx, ok := c.Cursh.(*shells.Indexsh); ok {
		filename := parts[1]
		err = configForIndex(idx, c, filename)
	}
	return
}

func configForIndex(idx *shells.Indexsh, c *api.Context,
	fname string) (err error) {
	var data []byte

	idx.ConfigFile = fname
	if data, err = ioutil.ReadFile(fname); err == nil {
		err = json.Unmarshal(data, &idx.Config)
	}
	idx.Killall(c)
	if err == nil {
		fmt.Fprintf(c.W, "Loaded config %q ...\n", idx.ConfigFile)
	}
	return
}

func init() {
	knownCommands["config"] = &ConfigCommand{}
}
