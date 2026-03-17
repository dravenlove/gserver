package gateway

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log"
	"net"
	"sync/atomic"

	"openclaw-go/internal/protocol"
	"openclaw-go/internal/router"
)

type Server struct {
	addr   string
	router *router.Router
	logger *log.Logger
	nextID atomic.Uint64
}

func NewServer(addr string, r *router.Router, logger *log.Logger) *Server {
	return &Server{
		addr:   addr,
		router: r,
		logger: logger,
	}
}

func (s *Server) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	s.logger.Printf("gateway listening at %s", s.addr)

	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				s.logger.Printf("gateway stopped")
				return nil
			}
			if errors.Is(err, net.ErrClosed) {
				s.logger.Printf("gateway listener closed")
				return nil
			}
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				s.logger.Printf("accept timeout error: %v", err)
				continue
			}
			return err
		}

		sessionID := s.nextID.Add(1)
		sess := NewSession(sessionID, conn)
		s.logger.Printf("session connected id=%s addr=%s", sess.ID(), sess.RemoteAddr())

		go s.handleConnection(ctx, sess)
	}
}

func (s *Server) handleConnection(ctx context.Context, session *Session) {
	defer func() {
		_ = session.Close()
		s.logger.Printf("session disconnected id=%s addr=%s", session.ID(), session.RemoteAddr())
	}()

	reader := bufio.NewReader(session.conn)
	for {
		packet, err := protocol.DecodePacket(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			s.logger.Printf("decode packet failed session=%s err=%v", session.ID(), err)
			return
		}

		if err := s.router.Dispatch(ctx, session, packet); err != nil {
			s.logger.Printf("dispatch failed session=%s msg_id=%d err=%v", session.ID(), packet.MsgID, err)
		}
	}
}
