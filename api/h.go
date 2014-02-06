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
		bs, err = json.MarshalIndent(obj, "", "  ")
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
	cmdsargs := make([][]string, 0)
	for {
		args, remstr := ParseCmdline(s)
		cmdsargs = append(cmdsargs, args)
		if remstr == "" {
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
			break
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
