// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package ssh

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Unknwon/com"
	"golang.org/x/crypto/ssh"
	log "gopkg.in/clog.v1"

	"github.com/gogits/gogs/models"
	"github.com/gogits/gogs/pkg/setting"
	"syscall"
)

func cleanCommand(cmd string) string {
	i := strings.Index(cmd, "git")
	if i == -1 {
		return cmd
	}
	return cmd[i:]
}

func handleServerConn(keyID string, chans <-chan ssh.NewChannel) {
	for newChan := range chans {
		if newChan.ChannelType() != "session" {
			newChan.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		ch, reqs, err := newChan.Accept()
		if err != nil {
			log.Error(3, "Error accepting channel: %v", err)
			continue
		}

		go func(in <-chan *ssh.Request) {
			defer ch.Close()
			for req := range in {
				payload := cleanCommand(string(req.Payload))
				switch req.Type {
				case "env":
					args := strings.Split(strings.Replace(payload, "\x00", "", -1), "\v")
					if len(args) != 2 {
						log.Warn("SSH: Invalid env arguments: '%#v'", args)
						continue
					}
					args[0] = strings.TrimLeft(args[0], "\x04")
					_, _, err := com.ExecCmdBytes("env", args[0]+"="+args[1])
					if err != nil {
						log.Error(3, "env: %v", err)
						return
					}
				case "exec":
					cmdName := strings.TrimLeft(payload, "'()")
					log.Trace("SSH: Payload: %v", cmdName)

					args := []string{"serv", "key-" + keyID, "--config=" + setting.CustomConf}
					log.Trace("SSH: Arguments: %v", args)
					cmd := exec.Command(setting.AppPath, args...)
					cmd.Env = append(os.Environ(), "SSH_ORIGINAL_COMMAND="+cmdName)

					stdout, err := cmd.StdoutPipe()
					if err != nil {
						log.Error(3, "SSH: StdoutPipe: %v", err)
						return
					}
					stderr, err := cmd.StderrPipe()
					if err != nil {
						log.Error(3, "SSH: StderrPipe: %v", err)
						return
					}
					input, err := cmd.StdinPipe()
					if err != nil {
						log.Error(3, "SSH: StdinPipe: %v", err)
						return
					}

					// FIXME: check timeout
					if err = cmd.Start(); err != nil {
						log.Error(3, "SSH: Start: %v", err)
						return
					}

					req.Reply(true, nil)
					go io.Copy(input, ch)
					io.Copy(ch, stdout)
					io.Copy(ch.Stderr(), stderr)

					if err = cmd.Wait(); err != nil {
						log.Error(3, "SSH: Wait: %v", err)
						// Fix 255 default return value error
						if t, ok := err.(*exec.ExitError); ok {
							log.Info("t:%s", t)

							es := t.Sys().(syscall.WaitStatus).ExitStatus()
							if es == 1 {
								ch.SendRequest("exit-status", false, []byte{0, 0, 0, 1})
							}
						}
						return
					}

					ch.SendRequest("exit-status", false, []byte{0, 0, 0, 0})
					return
				default:
				}
			}
		}(reqs)
	}
}

func listen(config *ssh.ServerConfig, host string, port int) {
	listener, err := net.Listen("tcp", host+":"+com.ToStr(port))
	if err != nil {
		log.Fatal(4, "Fail to start SSH server: %v", err)
	}
	for {
		// Once a ServerConfig has been configured, connections can be accepted.
		conn, err := listener.Accept()
		if err != nil {
			log.Error(3, "SSH: Error accepting incoming connection: %v", err)
			continue
		}

		// Before use, a handshake must be performed on the incoming net.Conn.
		// It must be handled in a separate goroutine,
		// otherwise one user could easily block entire loop.
		// For example, user could be asked to trust server key fingerprint and hangs.
		go func() {
			log.Trace("SSH: Handshaking for %s", conn.RemoteAddr())
			sConn, chans, reqs, err := ssh.NewServerConn(conn, config)
			if err != nil {
				if err == io.EOF {
					log.Warn("SSH: Handshaking was terminated: %v", err)
				} else {
					log.Error(3, "SSH: Error on handshaking: %v", err)
				}
				return
			}

			log.Trace("SSH: Connection from %s (%s)", sConn.RemoteAddr(), sConn.ClientVersion())
			// The incoming Request channel must be serviced.
			go ssh.DiscardRequests(reqs)
			go handleServerConn(sConn.Permissions.Extensions["key-id"], chans)
		}()
	}
}

// Listen starts a SSH server listens on given port.
func Listen(host string, port int, ciphers []string) {
	config := &ssh.ServerConfig{
		Config: ssh.Config{
			Ciphers: ciphers,
		},
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			pkey, err := models.SearchPublicKeyByContent(strings.TrimSpace(string(ssh.MarshalAuthorizedKey(key))))
			if err != nil {
				log.Error(3, "SearchPublicKeyByContent: %v", err)
				return nil, err
			}
			return &ssh.Permissions{Extensions: map[string]string{"key-id": com.ToStr(pkey.ID)}}, nil
		},
	}

	keyPath := filepath.Join(setting.AppDataPath, "ssh/gogs.rsa")
	if !com.IsExist(keyPath) {
		os.MkdirAll(filepath.Dir(keyPath), os.ModePerm)
		_, stderr, err := com.ExecCmd(setting.SSH.KeygenPath, "-f", keyPath, "-t", "rsa", "-N", "")
		if err != nil {
			panic(fmt.Sprintf("Fail to generate private key: %v - %s", err, stderr))
		}
		log.Trace("SSH: New private key is generateed: %s", keyPath)
	}

	privateBytes, err := ioutil.ReadFile(keyPath)
	if err != nil {
		panic("SSH: Fail to load private key")
	}
	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		panic("SSH: Fail to parse private key")
	}
	config.AddHostKey(private)

	go listen(config, host, port)
}
