package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"
	"text/scanner"
)

// get user's home directory
func HomeDir() string {
	hdir := os.Getenv("HOME") // try to find a HOME environment variable
	if hdir == "" {           // then try USERPROFILE for Windows
		hdir = os.Getenv("USERPROFILE")
		if hdir == "" {
			fmt.Printf("Unable to determine home directory, history file disabled\n")
		}
	}
	return hdir
}

func ResolveFile(filename, dirname string) string {
	switch {
	case filename[0] == os.PathSeparator:
		return filename
	case dirname != "":
		return path.Join(dirname, filename)
	default:
		cwd, _ := os.Getwd()
		return path.Join(cwd, filename)
	}
}

func Overlay(sources ...Config) (Config, error) {
	var err error

	if len(sources) == 0 {
		return nil, nil
	} else if len(sources) == 1 {
		return sources[0], nil
	}
	dest := sources[0]
	for _, source := range sources[1:] {
		if dest, err = OverlayProperty(dest, source); err != nil {
			return nil, err
		}
	}
	return dest, nil
}

func OverlaySlice(one, two []interface{}) []interface{} {
	sl := make([]interface{}, len(one)+len(two))
	copy(sl, one)
	copy(sl[len(one):], two)
	return sl
}

func OverlayProperty(one, two Config) (Config, error) {
	var err error

	newconfig := make(Config)
	for key, value := range one {
		newconfig[key] = value
	}
	for key, value := range two {
		if newconfig[key] == nil {
			newconfig[key] = value
			continue
		}
		switch value.(type) {
		case int, float32, float64, string, bool:
			newconfig[key] = value
		case []interface{}:
			if sl, ok := newconfig[key].([]interface{}); ok {
				newconfig[key] = OverlaySlice(sl, value.([]interface{}))
			} else {
				err = fmt.Errorf("OverlayProperty: Key type expected slice")
				return nil, err
			}
		case map[string]interface{}:
			if m, ok := newconfig[key].(map[string]interface{}); ok {
				newconfig[key], err = OverlayProperty(m, value.(Config))
				if err != nil {
					return nil, err
				}
			} else {
				err = fmt.Errorf("OverlayProperty: Key type expected property")
				return nil, err
			}
		default:
			return nil, fmt.Errorf("Unknown type %t", value)
		}
	}
	return newconfig, nil
}

// PrettyPrint converts `obj` into human readable format that can be directly
// rendered on the screen or file. If `attr` is not empty string and `obj` is
// map or struct, then `attr` is treated as key-to-map or struct-field.
func PrettyPrint(obj interface{}, attr string) (s string, err error) {
	var v, bs []byte
	var mobj map[string]interface{}
	var sobj []interface{}

	if v, err = json.Marshal(obj); err == nil {
		switch reflect.TypeOf(obj).Kind() {
		case reflect.Slice:
			json.Unmarshal(v, &sobj)
			obj = sobj
		case reflect.Map, reflect.Struct:
			json.Unmarshal(v, &mobj)
			obj = mobj
			if attr != "" {
				obj = mobj[attr]
			}
		default:
			err = fmt.Errorf("Neither slice nor map")
		}
		if bs, err = json.MarshalIndent(obj, "", "  "); err != nil {
			return "", err
		}
		s = string(bs)
	}
	return
}

func CommandLineTokens(line string) []string {
	var s scanner.Scanner
	toks := make([]string, 0, 8)

	s.Init(strings.NewReader(line))
	tok := s.Scan()
	for tok != scanner.EOF {
		toks = append(toks, s.TokenText())
		tok = s.Scan()
	}
	return toks
}

func SplitArgs(argstr string, sep string) []string {
	parts := strings.Split(strings.Trim(argstr, " "), sep)
	return trimArgs(parts)
}

func SplitArgN(argstr string, sep string, count int) []string {
	parts := strings.SplitN(strings.Trim(argstr, " "), sep, count)
	return trimArgs(parts)
}

func trimArgs(parts []string) []string {
	args := make([]string, 0)
	for _, s := range parts {
		s1 := strings.Trim(s, " ")
		if s1 != "" {
			args = append(args, s1)
		}
	}
	return args
}

func IsCommand(cmdname string, commands CommandMap) bool {
	for name, _ := range commands {
		if cmdname == name {
			return true
		}
	}
	return false
}

func ParseScript(s string) [][]string {
	cmdsargs := make([][]string, 0)
	for _, s := range strings.Split(s, SEP) {
		if strings.Trim(s, " ") == "" {
			continue
		}
		cmdsargs = append(cmdsargs, ParseCmdsline(s)...)
	}
	return cmdsargs
}

func ParseCmdsline(s string) [][]string {
	var args []string

	cmdsargs := make([][]string, 0)
	for {
		args, s = ParseCmdline(s)
		s = strings.Trim(s, " ")
		cmdsargs = append(cmdsargs, args)
		if s == "" {
			break
		}
	}
	return cmdsargs
}

func ParseCmdline(s string) ([]string, string) {
	var remstr string

	args := make([]string, 0)
	arg := make([]rune, 0)
	inStr := false

loop:
	for i, x := range s {
		switch {
		case inStr, x == '"':
			args = append(args, string(arg))
			inStr, arg = false, make([]rune, 0)
		case x == '"':
			inStr = true
		case x == ' ', x == '\t':
			if len(arg) > 0 {
				args = append(args, string(arg))
			}
			arg = make([]rune, 0)
		case x == ';':
			remstr = s[i+1:]
			break loop
		default:
			arg = append(arg, x)
		}
	}
	if len(arg) > 0 {
		args = append(args, string(arg))
	}
	return args, remstr
}

func CreateFile(filepath string, force bool) (err error) {
	create := true
	if _, err := os.Stat(filepath); err == nil {
		create = force
	}
	if create {
		_, err = os.Create(filepath)
	}
	return
}

func ShellDatadir() string {
	return path.Join(HomeDir(), CBSH_DIR)
}

func IsKill(kill chan bool) bool {
	select {
	case <-kill:
		return true
	default:
		return false
	}
}
