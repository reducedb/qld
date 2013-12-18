// Copyright (c) 2013 Jian Zhen <zhenjl@gmail.com>
//
// All rights reserved.
//
// Use of this source code is governed by the Apache 2.0 license.

package qld

import (
	"fmt"
	"github.com/dustin/randbo"
	"github.com/golang/glog"
	"log"
	"net"
	"sync"
)

var _ = log.Ldate

type server struct {
	cfg *config

	connId      int
	connIdMutex sync.RWMutex

	lns     []net.Listener
	netQuit chan bool

	// We will keep track of all the connections in this slice. However, this means
	// this slice could potentially grow large. If there's a million connections created
	// over time, it means this slice will be 4MB. So a potentialy memory "leak" here.
	conns []net.Conn
}

func newServer(cfg *config) (*server, error) {
	s := &server{
		cfg: cfg,
	}

	return s, nil
}

func (this *server) run() error {
	this.netQuit = make(chan bool)

	tcp, err := net.Listen("tcp", ":3306")
	if err != nil {
		return err
	}
	defer tcp.Close()

	this.lns = append(this.lns, tcp)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer func() {
			glog.V(3).Info("Quitting Accept() goroutine")
			for _, conn := range this.conns {
				conn.Close()
			}

			wg.Done()
		}()

		for {
			glog.V(3).Info("Listening for connections")
			conn, err := tcp.Accept()
			if err != nil {
				select {
				case <-this.netQuit:
					return
				default:
				}

				continue
			}

			this.conns = append(this.conns, conn)
			go this.handleConnection(conn, len(this.conns)-1)
		}
	}()

	wg.Wait()

	return nil
}

func (this *server) handleConnection(conn net.Conn, id int) error {
	defer func() {
		glog.V(3).Infof("Closing connection #%d", id)
		conn.Close()
		this.conns[id] = nil
	}()

	glog.V(3).Infof("Starting connection #%d", id)

	c := &connection{}
	c.rand = randbo.New()
	c.cfg = this.cfg
	c.Conn = conn
	c.id = id

	if n, err := c.rand.Read(c.cipher[:]); err != nil {
		return err
	} else if n != 20 {
		return fmt.Errorf("Connection/NewConnection: Error generating 20 bytes for random cipher. Only generated %d", n)
	}

	if err := c.handleConnectionPhase(); err != nil {
		return err
	}

	if err := c.handleCommandPhase(); err != nil {
		return err
	}

	return nil
}

func (this *server) quit() {
	close(this.netQuit)
	for _, ln := range this.lns {
		ln.Close()
	}
}
