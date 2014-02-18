package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"text/template"
	"time"
)

type Config map[string]interface{}
type Environ map[string]string

func LoadConfig(context Config, configfile, dirname string) (Config, error) {
	var err error
	var data []byte

	config := make(Config)
	configfile = ResolveFile(configfile, dirname)
	if data, err = ioutil.ReadFile(configfile); err != nil {
		return nil, err
	}
	if err = json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if context == nil {
		context = make(Config)
	}
	if config, err = expandConfig(context, config); err != nil {
		return nil, err
	}
	if context, err = Overlay(context, config); err != nil {
		return nil, err
	}
	return LoadNestedConfig(context, config, path.Dir(configfile))
}

func LoadNestedConfig(context, config Config, dirname string) (Config, error) {
	sources := []Config{config}
	if includeFiles, ok := config["include"].([]interface{}); ok {
		for _, includeFile := range includeFiles {
			config, err := LoadConfig(context, includeFile.(string), dirname)
			if err != nil {
				return nil, err
			}
			sources = append(sources, config)
		}
	} else if includeFile, ok := config["include"].(interface{}); ok {
		config, err := LoadConfig(context, includeFile.(string), dirname)
		if err != nil {
			return nil, err
		}
		sources = append(sources, config)
	}
	return Overlay(sources...)
}

func expandTerm(context Config, term interface{}) (interface{}, error) {
	if sl, ok := term.([]interface{}); ok {
		return expandSlice(context, sl)
	} else if m, ok := term.(map[string]interface{}); ok {
		return expandConfig(context, m)
	} else {
		return expandString(context, term)
	}
	return term, nil
}

func expandSlice(context Config, sl []interface{}) ([]interface{}, error) {
	newsl := make([]interface{}, 0, len(sl))
	for _, s := range sl {
		if term, err := expandTerm(context, s); err != nil {
			return nil, err
		} else {
			newsl = append(newsl, term)
		}
	}
	return newsl, nil
}

func expandConfig(context, config Config) (Config, error) {
	var ncontext Config
	var err error

	if ncontext, err = Overlay(context, config); err != nil {
		return nil, err
	}
	newconfig := make(Config)
	for key, value := range config {
		if term, err := expandTerm(ncontext, value); err != nil {
			return nil, err
		} else {
			newconfig[key] = term
		}
	}
	return newconfig, nil
}

func expandString(context Config, value interface{}) (interface{}, error) {
	if s, ok := value.(string); ok {
		for i := 0; i < 10; i++ {
			buf := bytes.NewBuffer([]byte{})
			t := template.New(fmt.Sprintln(time.Now().UnixNano()))
			if t, err := t.Parse(s); err != nil {
				return nil, err
			} else if err = t.Execute(buf, context); err != nil {
				return nil, err
			}
			s = string(buf.Bytes())
		}
		return s, nil
	}
	return value, nil
}

// GetProgramConfig extracts program configuration from master configuration.
func (config *Config) GetProgramConfig(name string) *Config {
	programs := (*config)["programs"].([]interface{})
	for _, program := range programs {
		prog := program.(Config)
		if prog["name"] == name {
			return &prog
		}
	}
	return nil
}

// TargetHost returns program's host name/ip from configuration.
func (config *Config) TargetHost(name string) string {
	pconf := config.GetProgramConfig(name)
	return (*pconf)["targethost"].(string)
}

// TargetRoot returns program's root directory from configuration.
func (config *Config) TargetRoot(name string) string {
	pconf := config.GetProgramConfig(name)
	return (*pconf)["targetroot"].(string)
}

// TargetHost returns slice of program's source repositories
func (config *Config) Repository(name string) []interface{} {
	pconf := config.GetProgramConfig(name)
	return (*pconf)["repository"].([]interface{})
}

// User returns username to use for ssh login.
func (config *Config) User(name string) (user string) {
	if name != "" {
		pconf := config.GetProgramConfig(name)
		user, _ = (*pconf)["user"].(string)
	}
	if user == "" {
		user = (*config)["user"].(string)
	}
	return user
}

// SshPoolSize returns allowed ssh clients in a connection pool
func (config *Config) SshPoolSize() int {
	return int((*config)["ssh.pool.size"].(float64))
}

// SshPoolOverflow returns allowed overflow of ssh clients in a connection pool
func (config *Config) SshPoolOverflow() int {
	return int((*config)["ssh.pool.overflow"].(float64))
}

// LogMaxsize returns maximum in-memory log lines
func (config *Config) LogMaxsize() int {
	return int((*config)["log.maxsize"].(float64))
}

// LogMaxsize returns maximum in-memory log lines
func (config *Config) LogColor(name string) string {
	pconf := config.GetProgramConfig(name)
	color, _ := (*pconf)["log.color"].(string)
	return color
}

// ProgramEnviron extracts environment from program configuration and returns
// a map of key, value string.
func (config *Config) ProgramEnviron(name string) Environ {
	if name != "" {
		config = config.GetProgramConfig(name)
	}
	environ := make(map[string]string)
	if confenv, ok := (*config)["environ"].(Config); ok {
		for key, i := range confenv {
			environ[key] = i.(string)
		}
	}
	return environ
}

// ProgramCommand returns the remote command and its argument.
func (config *Config) ProgramCommand(name string) string {
	if name != "" {
		config = config.GetProgramConfig(name)
	}
	command := (*config)["command"].(string)
	args := (*config)["commandargs"]
	if args != nil {
		for _, arg := range args.([]interface{}) {
			command += " " + arg.(string)
		}
	}
	return command
}
