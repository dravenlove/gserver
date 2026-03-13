package gateway

import (
	"fmt"
	"net"
	"strconv"
	"sync"

	"openclaw-go/internal/protocol"
)

type Session struct {
	id   string
	conn net.Conn
	mu   sync.Mutex
}

func NewSession(id uint64, conn net.Conn) *Session {
	return &Session{
		id:   strconv.FormatUint(id, 10),
		conn: conn,
	}
}

func (s *Session) ID() string {
	return s.id
}

func (s *Session) RemoteAddr() string {
	if s.conn == nil {
		return ""
	}
	return s.conn.RemoteAddr().String()
}

func (s *Session) Send(packet protocol.Packet) error {
	data, err := protocol.EncodePacket(packet)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	total := 0
	for total < len(data) {
		n, werr := s.conn.Write(data[total:])
		if werr != nil {
			return fmt.Errorf("write packet: %w", werr)
		}
		total += n
	}
	return nil
}

func (s *Session) Close() error {
	if s.conn == nil {
		return nil
	}
	return s.conn.Close()
}
