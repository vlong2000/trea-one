package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"mml-server/internal/mapping"
	"mml-server/internal/session"
)

type Server struct {
	addr     string
	mapper   *mapping.Table
	mu       sync.Mutex
	listener net.Listener
	conns    map[net.Conn]struct{}
	started  chan struct{}
}

func New(addr string, mapper *mapping.Table) *Server {
	return &Server{
		addr:    addr,
		mapper:  mapper,
		conns:   make(map[net.Conn]struct{}),
		started: make(chan struct{}),
	}
}

func (s *Server) Addr() string {
	<-s.started
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.listener.Addr().String()
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", s.addr, err)
	}
	s.mu.Lock()
	s.listener = ln
	s.mu.Unlock()
	close(s.started)

	go func() {
		<-ctx.Done()
		ln.Close()
	}()

	var wg sync.WaitGroup
	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				break
			}
			log.Printf("accept: %v", err)
			continue
		}
		s.track(conn, true)
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer s.track(conn, false)
			session.Handle(ctx, conn, s.mapper)
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-time.After(3 * time.Second):
		s.closeAll()
		<-done
		return nil
	}
}

func (s *Server) track(conn net.Conn, add bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if add {
		s.conns[conn] = struct{}{}
		return
	}
	delete(s.conns, conn)
}

func (s *Server) closeAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for c := range s.conns {
		c.Close()
	}
}
