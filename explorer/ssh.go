package explorer

/*
There are several different ways to go about testing SSH clients. I am a fan of hermetic tests.
So we are going to provide interfaces for our commands so we don't have to turn up actual SSH
servers. dialer() will provide the real SSH by default, providing back wrappers of the SSH
objects. This can be changed during tests to fakes that just do what are asked of them.
*/

import (
	"golang.org/x/crypto/ssh"
)

var dialer = func(node string, config *ssh.ClientConfig) (client, error) {
	real, err := ssh.Dial("tcp", node, config)
	if err != nil {
		return nil, err
	}

	return sshClient{client: real}, nil
}

type conn interface {
	close()
}

type session interface {
	combinedOutput(cmd string) ([]byte, error)
	close()
}

type client interface {
	conn() conn
	newSession() (session, error)
}

// sshConn implements conn using the SSH library's Conn object.
type sshConn struct {
	conn ssh.Conn
}

// close implements conn.Close().
func (s sshConn) close() {
	s.conn.Close()
}

// sshSession implements session using the SSH library's Session object.
type sshSession struct {
	session *ssh.Session	
}

// combinedOutput implements session.combinedOutput().
func (s sshSession) combinedOutput(cmd string) ([]byte, error) {
	return s.session.CombinedOutput(cmd)
}

func (s sshSession) close() {
	s.session.Close()
}

type sshClient struct {
	client *ssh.Client
}

func (s sshClient) conn() conn {
	return sshConn{conn: s.client.Conn}
}

func (s sshClient) newSession() (session, error) {
	real, err := s.client.NewSession()
	if err != nil {
		return nil, err
	}

	return sshSession{session: real}, nil
}