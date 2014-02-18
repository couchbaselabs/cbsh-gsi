package sshc

import (
	"fmt"
	"github.com/couchbaselabs/cbsh/api"
)

type Log struct {
	lines  []string
	cursor int
}

type Program struct {
	Name    string
	Config  api.Config
	Outch   chan<- string
	Errch   chan<- string
	fabric  *Fabric
	outlog  *Log
	errlog  *Log
	quit    chan bool
	healthy bool
}

func (fabric *Fabric) RunProgram(name string, printch outStr) (*Program, error) {
	logMaxSize := fabric.Config.LogMaxsize()
	// construct the program structure
	program := Program{
		Name:    name,
		Config:  fabric.Config,
		Outch:   printch,
		Errch:   printch,
		fabric:  fabric,
		outlog:  &Log{lines: make([]string, logMaxSize)},
		errlog:  &Log{lines: make([]string, logMaxSize)},
		quit:    make(chan bool),
		healthy: true,
	}
	fabric.SetProgram(name, &program)
	go program.runProgram()
	return &program, nil
}

func (p *Program) runProgram() (err error) {
	chout := make(chan string)
	cherr := make(chan string)
	go func() {
		var s string
		var ok bool
	loop:
		for {
			select {
			case s, ok = <-chout:
				if ok {
					p.appendLog(p.outlog, s)
					p.Outch <- p.Sprintf("%v", s)
				}
			case s, ok = <-cherr:
				if ok {
					p.appendLog(p.errlog, s)
					p.Errch <- p.Sprintf("%v", s)
				}
			case <-p.quit:
				ok = false
			}
			if ok == false {
				break loop
			}
		}
	}()
	err = p.fabric.ExecRemoteCommand(&remoteCommand{
		host:    p.Config.TargetHost(p.Name),
		user:    p.Config.User(p.Name),
		environ: p.Config.ProgramEnviron(p.Name),
		command: p.Config.ProgramCommand(p.Name),
		outch:   chout,
		errch:   cherr,
		quit:    p.quit,
	}, true)
	p.Close()
	return
}

func (p *Program) Kill() {
	p.Outch <- p.Sprintf("Getting Killed\n")
	p.Close()
}

func (p *Program) Close() {
	defer func() { recover() }()
	close(p.quit)
	p.healthy = true
}

func (p *Program) Sprintf(format string, args ...interface{}) string {
	var prefix string
	colorstr := p.Config.LogColor(p.Name)
	switch colorstr {
	case "black":
		prefix = fmt.Sprintf("[%v] ", api.Black(p.Name))
	case "red":
		prefix = fmt.Sprintf("[%v] ", api.Red(p.Name))
	case "green":
		prefix = fmt.Sprintf("[%v] ", api.Green(p.Name))
	case "blue":
		prefix = fmt.Sprintf("[%v] ", api.Blue(p.Name))
	case "magenta":
		prefix = fmt.Sprintf("[%v] ", api.Magenta(p.Name))
	case "cyan":
		prefix = fmt.Sprintf("[%v] ", api.Cyan(p.Name))
	case "white":
		prefix = fmt.Sprintf("[%v] ", api.White(p.Name))
	case "yellow":
		prefix = fmt.Sprintf("[%v] ", api.Yellow(p.Name))
	default:
		prefix = fmt.Sprintf("[%v] ", p.Name)
	}
	s := fmt.Sprintf(format, args...)
	return prefix + s
}

func (p *Program) appendLog(log *Log, s string) {
	maxsize := p.Config.LogMaxsize()
	l := len(log.lines)
	if len(log.lines) >= maxsize {
		copy(log.lines[1:], log.lines[:l-1])
		log.lines[0] = s
	}
	if log.cursor < (maxsize - 1) {
		log.cursor += 1
	}
}
