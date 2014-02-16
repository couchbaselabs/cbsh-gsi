package sshc

import (
	"code.google.com/p/go.crypto/ssh"
	"errors"
	"net"
	"os"
	"time"
)

var errClosedPool = errors.New("the pool is closed")
var errNoPool = errors.New("no pool")
var errTimeout = errors.New("timeout")

// Default timeout for retrieving a connection from the pool.
var ConnPoolTimeout = time.Hour * 24 * 30

// ConnPoolAvailWaitTime is the amount of time to wait for an existing
// connection from the pool before considering the creation of a new
// one.
var ConnPoolAvailWaitTime = time.Millisecond

type connectionPool struct {
	host        string
	username    string
	connections chan *ssh.ClientConn
	createsem   chan bool
}

func newConnectionPool(host, user string, poolSize, poolOverflow int) *connectionPool {
	return &connectionPool{
		host:        host,
		username:    user,
		connections: make(chan *ssh.ClientConn, poolSize),
		createsem:   make(chan bool, poolSize+poolOverflow),
	}
}

// ConnPoolTimeout is notified whenever connections are acquired from a pool.
var ConnPoolCallback func(host string, source string, start time.Time, err error)

func mkConn(user, host string) (client *ssh.ClientConn, err error) {
	// ssh-agent
	agent_sock, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return nil, err
	}
	defer agent_sock.Close()

	// ssh-client
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.ClientAuth{
			ssh.ClientAuthAgent(ssh.NewAgentClient(agent_sock)),
		},
	}
	dest := host + ":22"
	if client, err = ssh.Dial("tcp", dest, config); err != nil {
		return nil, err
	}
	return client, nil
}

func (cp *connectionPool) Close() (err error) {
	defer func() { err, _ = recover().(error) }()
	close(cp.connections)
	for c := range cp.connections {
		c.Close()
	}
	return
}

func (cp *connectionPool) GetWithTimeout(d time.Duration) (rv *ssh.ClientConn, err error) {
	if cp == nil {
		return nil, errNoPool
	}

	path := ""

	if ConnPoolCallback != nil {
		defer func(path *string, start time.Time) {
			ConnPoolCallback(cp.host, *path, start, err)
		}(&path, time.Now())
	}

	path = "short-circuit"

	// short-circuit available connetions.
	select {
	case rv, isopen := <-cp.connections:
		if !isopen {
			return nil, errClosedPool
		}
		return rv, nil
	default:
	}

	t := time.NewTimer(ConnPoolAvailWaitTime)
	defer t.Stop()

	// Try to grab an available connection within 1ms
	select {
	case rv, isopen := <-cp.connections:
		path = "avail1"
		if !isopen {
			return nil, errClosedPool
		}
		return rv, nil
	case <-t.C:
		// No connection came around in time, let's see
		// whether we can get one or build a new one first.
		t.Reset(d) // Reuse the timer for the full timeout.
		select {
		case rv, isopen := <-cp.connections:
			path = "avail2"
			if !isopen {
				return nil, errClosedPool
			}
			return rv, nil
		case cp.createsem <- true:
			path = "create"
			// Build a connection if we can't get a real one.
			// This can potentially be an overflow connection, or
			// a pooled connection.
			rv, err := mkConn(cp.username, cp.host)
			if err != nil {
				// On error, release our create hold
				<-cp.createsem
			}
			return rv, err
		case <-t.C:
			return nil, errTimeout
		}
	}
}

func (cp *connectionPool) Get() (*ssh.ClientConn, error) {
	return cp.GetWithTimeout(ConnPoolTimeout)
}

func (cp *connectionPool) Hijack() (*ssh.ClientConn, error) {
	client, err := cp.Get()
	<-cp.createsem
	return client, err
}

func (cp *connectionPool) Return(c *ssh.ClientConn) {
	if c == nil {
		return
	}

	if cp == nil {
		c.Close()
	}

	defer func() {
		if recover() != nil {
			// This happens when the pool has already been
			// closed and we're trying to return a
			// connection to it anyway.  Just close the
			// connection.
			c.Close()
		}
	}()

	select {
	case cp.connections <- c:
	default:
		// Overflow connection.
		<-cp.createsem
		c.Close()
	}
}
