package sshc

import (
	"bufio"
	"code.google.com/p/go.crypto/ssh"
	"fmt"
	"github.com/couchbaselabs/cbsh/api"
	"io"
	"strings"
	"sync"
)

// TODO: lock protect Fabric.pools and Fabric.programs

type outStr chan<- string // for stdout and stderr
type inStr <-chan string  // for stdin

type remoteCommand struct {
	host    string
	user    string
	environ api.Environ
	command string
	inch    chan string
	outch   outStr
	errch   outStr
	quit    chan bool
}

// Fabric is an instance of cluster managment.
type Fabric struct {
	Config   api.Config
	mu       sync.Mutex
	pools    map[string]*connectionPool
	programs map[string]*Program
}

// StartFabric creates a new instace of cluster management.
func StartFabric(config api.Config) (*Fabric, error) {
	fabric := Fabric{
		Config:   config,
		pools:    make(map[string]*connectionPool),
		programs: make(map[string]*Program),
	}
	return &fabric, nil
}

// IsHealthy return whether fabric is a healthy state.
func (fabric *Fabric) IsHealthy() bool {
	return fabric != nil
}

// Atomically get a running program's managment structure from running-list
func (fabric *Fabric) GetProgram(name string) *Program {
	fabric.mu.Lock()
	defer fabric.mu.Unlock()
	if fabric.programs != nil {
		return fabric.programs[name]
	}
	return nil
}

// Atomically set a program's managment structure to running-list.
func (fabric *Fabric) SetProgram(name string, p *Program) {
	fabric.mu.Lock()
	defer fabric.mu.Unlock()
	if fabric.programs != nil {
		fabric.programs[name] = p
	}
}

// Atomically delete a program's management structure from running-list
func (fabric *Fabric) DeleteProgram(name string) {
	fabric.mu.Lock()
	defer fabric.mu.Unlock()
	if fabric.programs != nil {
		delete(fabric.programs, name)
	}
}

// Atomically get a connection pool for host
func (fabric *Fabric) GetPool(host string) *connectionPool {
	fabric.mu.Lock()
	defer fabric.mu.Unlock()
	if fabric.pools != nil {
		return fabric.pools[host]
	}
	return nil
}

// Atomically set a connection pool for host
func (fabric *Fabric) SetPool(host string, cp *connectionPool) {
	fabric.mu.Lock()
	defer fabric.mu.Unlock()
	if fabric.pools != nil {
		fabric.pools[host] = cp
	}
}

// Atomically delete a connection pool for host
func (fabric *Fabric) DeletePool(host string) {
	fabric.mu.Lock()
	defer fabric.mu.Unlock()
	if fabric.pools != nil {
		delete(fabric.pools, host)
	}
}

func (fabric *Fabric) Close() {
	fabric.mu.Lock()
	defer fabric.mu.Unlock()
	for _, cp := range fabric.pools {
		cp.Close()
	}
	for _, p := range fabric.programs {
		p.Kill()
	}
	fabric.pools, fabric.programs = nil, nil
}

func (fabric *Fabric) ExecRemoteCommand(cmd *remoteCommand, daemon bool) (err error) {
	var client *ssh.ClientConn
	var cp *connectionPool

	// Get connection pool for `host`
	if cp, err = fabric.getConnectionPool(cmd.host, cmd.user); err != nil {
		return
	}
	if client, err = cp.Get(); err != nil {
		return
	}

	// Get ssh session
	session, _ := client.NewSession()
	stdin, stdout, stderr, err := remoteStandardio(session)
	if err != nil {
		return
	}
	if daemon {
		modes := ssh.TerminalModes{
			ssh.ECHO:          0,     // disable echoing
			ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
			ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
		}
		if err = session.RequestPty("xterm", 80, 40, modes); err != nil {
			return
		}
	}
	// Setup stdin, stdout & stderr readers
	if cmd.inch != nil {
		go writeIn(stdin, cmd.inch, cmd.errch)
	}
	if cmd.outch != nil {
		go readOut(stdout, cmd.outch, false)
	}
	if cmd.errch != nil {
		go readOut(stderr, cmd.errch, false)
	}
	// Setup remote's environment and run the command
	if err = setEnviron(cmd.environ, session); err == nil {
		if daemon {
			go func() {
				defer func() { recover() }()
				err = session.Run(cmd.command)
				close(cmd.quit)
			}()
			<-cmd.quit
		} else {
			err = session.Run(cmd.command)
		}
		if err != nil && cmd.errch != nil {
			cmd.errch <- fmt.Sprintln(err)
		}
	}
	if cmd.inch != nil {
		close(cmd.inch)
	}
	session.Signal(ssh.SIGTERM)
	session.Close()
	cp.Return(client)
	return
}

func (fabric *Fabric) MakeRemoteDirs(host, user, dir string, ch outStr) (err error) {
	command := fmt.Sprintf("mkdir -p %v", dir)
	ch <- fmt.Sprintf("Creating directory %q\n", dir)
	cmd := &remoteCommand{
		host: host, user: user, command: command, outch: ch, errch: ch}
	err = fabric.ExecRemoteCommand(cmd, false)
	return err
}

func (fabric *Fabric) RemoveRemoteDir(host, user, dir string, ch outStr) (err error) {
	command := fmt.Sprintf("rm -rf %v", dir)
	ch <- fmt.Sprintf("Removing directory %q\n", dir)
	cmd := &remoteCommand{
		host: host, user: user, command: command, outch: ch, errch: ch}
	err = fabric.ExecRemoteCommand(cmd, false)
	return err
}

func (fabric *Fabric) IsDir(host, user, dir string) bool {
	command := fmt.Sprintf(`eval 'if [ -d %v ]; then echo "true"; else echo "false"; fi'`, dir)
	ch := make(chan string)
	cmd := &remoteCommand{host: host, user: user, command: command, outch: ch}
	go func() {
		fabric.ExecRemoteCommand(cmd, false)
	}()
	s, _ := <-ch
	s = strings.Trim(s, "\n")
	close(ch)
	if s == "true" {
		return true
	}
	return false
}

func (fabric *Fabric) Killall() {
	fabric.mu.Lock()
	defer fabric.mu.Unlock()
	for _, p := range fabric.programs {
		if p != nil {
			p.Kill()
		}
	}
}

func (fabric *Fabric) KillProgram(progname string) error {
	fabric.mu.Lock()
	defer fabric.mu.Unlock()
	if p := fabric.programs[progname]; p != nil {
		p.Kill()
		delete(fabric.programs, progname)
		return nil
	}
	return fmt.Errorf("Program name %v not found", progname)
}

func (fabric *Fabric) getConnectionPool(host, user string) (*connectionPool, error) {
	cp := fabric.GetPool(host)
	if cp == nil {
		poolSize := fabric.Config.SshPoolSize()
		poolOverflow := fabric.Config.SshPoolOverflow()
		cp = newConnectionPool(host, user, poolSize, poolOverflow)
		if cp == nil {
			return nil, fmt.Errorf("Unable to create pool for %v", host)
		}
		fabric.SetPool(host, cp)
	}
	return cp, nil
}

func remoteStandardio(s *ssh.Session) (io.WriteCloser, io.Reader, io.Reader, error) {
	var stdin io.WriteCloser
	var stdout io.Reader
	var stderr io.Reader
	var err error

	// plumb into standard input
	if stdin, err = s.StdinPipe(); err != nil {
		return nil, nil, nil, fmt.Errorf("Error: %v\n", err)
	}
	// plumb into standard output
	if stdout, err = s.StdoutPipe(); err != nil {
		return nil, nil, nil, fmt.Errorf("Error: %v\n", err)
	}
	// plumb into standard error
	if stderr, err = s.StderrPipe(); err != nil {
		return nil, nil, nil, fmt.Errorf("Error: %v\n", err)
	}
	return stdin, stdout, stderr, nil
}

func readOut(rd io.Reader, ch outStr, doclose bool) {
	r := bufio.NewReader(rd)
	for {
		if buf, _ := r.ReadBytes(api.NEWLINE); len(buf) > 0 {
			ch <- string(buf)
		} else {
			if doclose {
				close(ch)
			}
			break
		}
	}
}

func writeIn(w io.WriteCloser, inch inStr, errch outStr) {
	for {
		if s, ok := <-inch; ok {
			bs := []byte(s)
			if _, err := w.Write(bs); err != nil {
				errch <- fmt.Sprintln(err)
			}
		} else {
			w.Close()
			break
		}
	}
}

func setEnviron(environ api.Environ, session *ssh.Session) (err error) {
	//if environ != nil {
	//    for key, value := range environ {
	//        if err = session.Setenv(key, value); err != nil {
	//            fmt.Println(10, key, value)
	//            return
	//        }
	//    }
	//}
	return
}
