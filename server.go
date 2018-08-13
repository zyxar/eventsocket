// Copyright 2015 Nevio Vesic
// Please check out LICENSE file for more information about what you CAN and what you CANNOT do!
// Basically in short this is a free software for you to do whatever you want to do BUT copyright must be included!
// I didn't write all of this code so you could say it's yours.
// MIT License

package eventsocket

import (
	"fmt"
	"net"
	"os"
	"sync"
)

// Server - In case you need to start server, this Struct have it covered
type Server struct {
	net.Listener
	closeWg   sync.WaitGroup
	closeOnce sync.Once
	closeChan chan struct{}
}

// Start - Will start new outbound server
func (s *Server) Start(fn func(conn *SocketConnection)) {
	s.closeWg.Add(1)
	go func() {
		defer s.closeWg.Done()
		for {
			c, err := s.Accept()
			if err != nil {
				select {
				case <-s.closeChan:
				default:
					Error(EListenerConnection, err)
				}
				return
			}
			conn := &SocketConnection{
				Conn: c,
				err:  make(chan error),
				m:    make(chan *Message),
			}
			Notice("Got new connection from: %s", conn.OriginatorAddr())
			go conn.Handle()
			go fn(conn)
		}
	}()
}

// Close - Will close server connection
func (s *Server) Close() (err error) {
	s.closeOnce.Do(func() {
		close(s.closeChan)
		err = s.Listener.Close()
	})
	return
}

// Wait waits until backgroud routines quit
func (s *Server) Wait() { s.closeWg.Wait() }

// NewServer - Will instanciate new outbound server
func NewServer(addr string) (s *Server, err error) {
	if len(addr) < 2 {
		addr = os.Getenv("GOESL_OUTBOUND_SERVER_ADDR")
		if addr == "" {
			return nil, fmt.Errorf(EInvalidServerAddr, addr)
		}
	}
	s = &Server{}
	if s.Listener, err = net.Listen("tcp", addr); err != nil {
		Error(ECouldNotStartListener, err)
		return
	}
	s.closeChan = make(chan struct{})
	return
}
