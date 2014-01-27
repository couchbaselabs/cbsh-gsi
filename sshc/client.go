package sshc

import (
	"bufio"
	"code.google.com/p/go.crypto/ssh"
	"fmt"
	"github.com/couchbaselabs/cbsh/api"
	"io"
	"net"
	"os"
	"os/exec"
)

type Log struct {
	lines  []string
	cursor int
}

type localCommand struct {
	name string
	args []string
}

type Program struct {
	Name   string
	Outch  chan string
	Errch  chan string
	Config map[string]interface{}
	config map[string]interface{}
	outlog *Log
	errlog *Log
	client *ssh.ClientConn
	host   string
	user   string
	local  []localCommand
	remote []string
	quit   chan bool
}

type stdPlumber interface {
	StderrPipe() (io.ReadCloser, error)
	StdinPipe() (io.WriteCloser, error)
	StdoutPipe() (io.ReadCloser, error)
}

var fabprog = Program{
	Name:   "fab",
	local:  make([]localCommand, 0),
	remote: make([]string, 0),
}

func RunProgram(name string, conf map[string]interface{}) *Program {
	programs := conf["programs"].(map[string]interface{})
	progConfig := (programs[name]).(map[string]interface{})
	logMaxSize := int(conf["log.maxsize"].(float64))
	// construct the program structure
	program := Program{
		Name:   name,
		Config: conf,
		Outch:  make(chan string),
		Errch:  make(chan string),
		quit:   make(chan bool),
		config: progConfig,
		host:   progConfig["host"].(string),
		user:   progConfig["user"].(string),
		local:  localCommands(progConfig["local"].([]interface{})),
		remote: remoteCommands(progConfig["remote"].([]interface{})),
		outlog: &Log{lines: make([]string, logMaxSize)},
		errlog: &Log{lines: make([]string, logMaxSize)},
	}

	// ssh-agen
	agent_sock, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		panic(err)
	}
	defer agent_sock.Close()

	// ssh-client
	config := &ssh.ClientConfig{
		User: program.user,
		Auth: []ssh.ClientAuth{
			ssh.ClientAuthAgent(ssh.NewAgentClient(agent_sock)),
		},
	}
	dest := program.host + ":22"
	program.client, err = ssh.Dial("tcp", dest, config)
	if err != nil {
		panic(err)
	}
	go program.runProgram()
	return &program
}

func Killall(programs []*Program) {
	for _, p := range programs {
		if p != nil {
			p.Kill()
		}
	}
}

func (p *Program) Kill() {
	if p.client != nil {
		p.Outch <- p.Sprintf("Getting Killed\n")
		close(p.quit)
		p.client.Close()
		close(p.Outch)
		close(p.Errch)
		p.client = nil
	}
}

func (p *Program) IsRunning() bool {
	return p.client != nil
}

func (p *Program) runProgram() {
	var err error
	for _, cmd := range p.local {
		if err = p.runLocalCommand(cmd.name, cmd.args); err != nil {
			p.Kill()
			break
		}
	}
	if err == nil {
		for _, cmd := range p.remote {
			if err = p.runRemoteCommand(cmd); err != nil {
				p.Kill()
				break
			}
		}
	}
}

func (p *Program) runLocalCommand(name string, args []string) (err error) {
	// Create a command session
	session := exec.Command(name, args...)
	_, stdout, stderr, errstr := localStandardio(session)
	if errstr != "" {
		p.Errch <- p.Sprintf(errstr)
		p.Kill()
	}

	chout := make(chan string)
	cherr := make(chan string)
	go readOut(chout, stdout)
	go readErr(cherr, stderr)

	go func() {
		var s string
		var ok bool
	loop:
		for {
			select {
			case s, ok = <-chout:
				if ok {
					p.Outch <- p.Sprintf("%v", s)
				}
			case s, ok = <-cherr:
				if ok {
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

	p.Outch <- p.Sprintf("Executing command %v %v...\n", name, args)
	err = session.Run()
	if err != nil {
		p.Kill()
	}
	return
}

func (p *Program) runRemoteCommand(cmd string) (err error) {
	// Create a session
	session, _ := p.client.NewSession()
	_, stdout, stderr, errstr := remoteStandardio(session)
	if errstr != "" {
		p.Errch <- p.Sprintf(errstr)
		p.Kill()
	}

	chout := make(chan string)
	cherr := make(chan string)
	go readOut(chout, stdout)
	go readErr(cherr, stderr)

	go func() {
		var s string
		var ok bool
	loop:
		for {
			select {
			case s, ok = <-chout:
				if ok {
					appendLog(p.outlog, s, p.Config)
					p.Outch <- p.Sprintf("%v", s)
				}
			case s, ok = <-cherr:
				if ok {
					appendLog(p.errlog, s, p.Config)
					p.Errch <- p.Sprintf("%v", s)
				}
			case <-p.quit:
				ok = false
			}
			if ok == false {
				session.Signal(ssh.SIGTERM)
				session.Close()
				break loop
			}
		}
	}()

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	if err = session.RequestPty("xterm", 80, 40, modes); err != nil {
		msg := fmt.Sprintf("request for pseudo terminal failed: %s", err)
		p.Errch <- p.Sprintf("Error: %v\n", msg)
	}
	p.Outch <- p.Sprintf("Executing command %v ...", cmd)
	err = session.Run(cmd)
	if err != nil {
		p.Kill()
	}
	return err
}

func (p *Program) Sprintf(format string, args ...interface{}) string {
	var prefix string
	colorstr, ok := p.config["log.color"].(string)
	if !ok {
		colorstr = ""
	}
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

func appendLog(log *Log, s string, config map[string]interface{}) {
	maxsize := int(config["log.maxsize"].(float64))
	l := len(log.lines)
	if len(log.lines) >= maxsize {
		copy(log.lines[1:], log.lines[:l-1])
		log.lines[0] = s
	}
	if log.cursor < (maxsize - 1) {
		log.cursor += 1
	}
}

func remoteCommands(vs []interface{}) []string {
	ss := make([]string, 0)
	for _, v := range vs {
		ss = append(ss, v.(string))
	}
	return ss
}

func localCommands(vs []interface{}) []localCommand {
	cmds := make([]localCommand, 0)
	for _, cs := range vs {
		ss := make([]string, 0)
		for _, c := range cs.([]interface{}) {
			ss = append(ss, c.(string))
		}
		if len(ss) > 0 {
			cmds = append(cmds, localCommand{name: ss[0], args: ss[1:]})
		}
	}
	return cmds
}

func localStandardio(
	s *exec.Cmd) (io.WriteCloser, io.ReadCloser, io.ReadCloser, string) {

	var stdin io.WriteCloser
	var stdout io.ReadCloser
	var stderr io.ReadCloser
	var err error

	// plumb into standard input
	if stdin, err = s.StdinPipe(); err != nil {
		return nil, nil, nil, fmt.Sprintf("Error: %v\n", err)
	}
	// plumb into standard output
	if stdout, err = s.StdoutPipe(); err != nil {
		return nil, nil, nil, fmt.Sprintf("Error: %v\n", err)
	}
	// plumb into standard error
	if stderr, err = s.StderrPipe(); err != nil {
		return nil, nil, nil, fmt.Sprintf("Error: %v\n", err)
	}
	return stdin, stdout, stderr, ""
}

func remoteStandardio(
	s *ssh.Session) (io.WriteCloser, io.Reader, io.Reader, string) {

	var stdin io.WriteCloser
	var stdout io.Reader
	var stderr io.Reader
	var err error

	// plumb into standard input
	if stdin, err = s.StdinPipe(); err != nil {
		return nil, nil, nil, fmt.Sprintf("Error: %v\n", err)
	}
	// plumb into standard output
	if stdout, err = s.StdoutPipe(); err != nil {
		return nil, nil, nil, fmt.Sprintf("Error: %v\n", err)
	}
	// plumb into standard error
	if stderr, err = s.StderrPipe(); err != nil {
		return nil, nil, nil, fmt.Sprintf("Error: %v\n", err)
	}
	return stdin, stdout, stderr, ""
}

func readOut(ch chan string, stdout io.Reader) {
	r := bufio.NewReader(stdout)
	for {
		if buf, err := r.ReadBytes(api.NEWLINE); len(buf) > 0 {
			ch <- string(buf)
		} else if err != nil {
			close(ch)
			return
		}
	}
}

func readErr(ch chan string, stderr io.Reader) {
	r := bufio.NewReader(stderr)
	for {
		if buf, err := r.ReadBytes(api.NEWLINE); len(buf) > 0 {
			ch <- string(buf)
		} else if err != nil {
			close(ch)
			return
		}
	}
}
