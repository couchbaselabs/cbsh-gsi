package sshc

import (
	"fmt"
	"github.com/couchbaselabs/cbsh/api"
	"net/url"
	"strings"
)

func (fabric *Fabric) InstallProgram(prog string, printch outStr, force bool) (err error) {
	host := fabric.Config.TargetHost(prog)
	repos := fabric.Config.Repository(prog)
	environ := fabric.Config.ProgramEnviron(prog)
	user := fabric.Config.User(prog)
	for _, i := range repos {
		repo := i.(api.Config)
		target := repo["target"].(string)
		fmt.Println(target)
		if force || fabric.IsDir(host, user, target) == false {
			err = fabric.CloneRepository(host, prog, repo, printch) // Clone
			if err != nil {
				return
			}
			err = fabric.PatchRepository(host, prog, repo, printch) // Patch
			if err != nil {
				return
			}
		} else {
			printch <- fmt.Sprintf("target %q already exists\n", target)
		}

		if install_commands, ok := repo["install"].([]interface{}); ok {
			for _, i := range install_commands { // Install
				command := i.(string)
				printch <- fmt.Sprintf("%v\n", command)
				err = fabric.ExecRemoteCommand(&remoteCommand{
					host:    host,
					user:    user,
					environ: environ,
					command: command,
					outch:   printch,
					errch:   printch,
				}, false)
			}
		}
	}
	return
}

func (fabric *Fabric) UninstallProgram(prog string, printch outStr) (err error) {
	host := fabric.Config.TargetHost(prog)
	repos := fabric.Config.Repository(prog)
	environ := fabric.Config.ProgramEnviron(prog)
	user := fabric.Config.User(prog)
	for _, i := range repos {
		repo := i.(api.Config)
		target := repo["target"].(string)
		err = fabric.RemoveRemoteDir(host, user, target, printch)
		if err != nil {
			return
		}
		for _, i := range repo["uninstall"].([]interface{}) { // Uninstall
			command := i.(string)
			printch <- fmt.Sprintf("%v\n", command)
			err = fabric.ExecRemoteCommand(&remoteCommand{
				host:    host,
				user:    user,
				environ: environ,
				command: command,
				outch:   printch,
				errch:   printch,
			}, false)
		}
	}
	return
}

func (fabric *Fabric) CloneRepository(
	host, progname string, repo api.Config, printch outStr) (err error) {

	target, source := repo["target"].(string), repo["source"].(string)
	user := fabric.Config.User(progname)

	// Remove target repository
	err = fabric.RemoveRemoteDir(host, user, target, printch)
	if err != nil {
		return
	}
	// Create target repository
	err = fabric.MakeRemoteDirs(host, user, target, printch)
	if err != nil {
		return
	}
	// Clone repository from source to target
	command := fmt.Sprintf("git clone --quiet %v %v", source, target)
	printch <- fmt.Sprintf("%v\n", command)
	err = fabric.ExecRemoteCommand(&remoteCommand{
		host:    host,
		user:    user,
		environ: fabric.Config.ProgramEnviron(progname),
		command: command,
		outch:   printch,
		errch:   printch,
	}, false)
	return
}

func (fabric *Fabric) PatchRepository(
	host, progname string, repo api.Config, printch outStr) (err error) {

	var diff string
	if diff, err = fabric.DiffRepository(progname, repo, printch); err != nil {
		return
	} else if diff == "" {
		return
	}

	target := repo["target"].(string)
	inch := make(chan string)
	command := fmt.Sprintf("cd %v; git apply -", target)
	printch <- fmt.Sprintf("%v\n", command)
	go func() {
		inch <- diff
		close(inch)
	}()
	err = fabric.ExecRemoteCommand(&remoteCommand{
		host:    host,
		user:    fabric.Config.User(progname),
		environ: fabric.Config.ProgramEnviron(progname),
		command: command,
		inch:    inch,
		outch:   printch,
		errch:   printch,
	}, false)
	return
}

func (fabric *Fabric) DiffRepository(
	prog string, repo api.Config, printch outStr) (string, error) {

	var host, path string
	var user *url.Userinfo
	var err error
	var u *url.URL

	source := repo["source"].(string)
	if u, err = url.Parse(source); err != nil {
		return "", err
	}
	if u.Scheme == "ssh" {
		user, host, path = u.User, u.Host, u.Path
	}

	diffs := make([]string, 0)
	if host != "" {
		fabric.getConnectionPool(host, user.Username())
		command := fmt.Sprintf("cd %v; git diff", path)
		diffch := make(chan string)
		errch := make(chan string)
		q := make(chan bool)

		go func() {
			printch <- fmt.Sprintf("%v\n", command)
			err = fabric.ExecRemoteCommand(&remoteCommand{
				host:    host,
				user:    fabric.Config.User(prog),
				command: command,
				outch:   diffch,
				errch:   errch,
			}, false)
			if err != nil {
				printch <- fmt.Sprintln(err)
			}
			close(q)
		}()

	loop:
		for {
			var s string
			var ok bool
			select {
			case s, ok = <-diffch:
				diffs = append(diffs, s)
			case s, ok = <-errch:
				printch <- s
				break loop
			case <-q:
				break loop
			}
			if !ok {
				break
			}
		}
		return strings.Join(diffs, ""), err
	}
	return "", fmt.Errorf("Cannot execute local commands")
}
