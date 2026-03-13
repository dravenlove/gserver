package router

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"openclaw-go/internal/protocol"
)

var ErrHandlerNotFound = errors.New("handler not found")

type Session interface {
	ID() string
	RemoteAddr() string
	Send(packet protocol.Packet) error
}

type Handler func(ctx context.Context, session Session, packet protocol.Packet) error

type Router struct {
	mu       sync.RWMutex
	handlers map[uint16]Handler
}

func New() *Router {
	return &Router{
		handlers: make(map[uint16]Handler),
	}
}

func (r *Router) Register(msgID uint16, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[msgID] = handler
}

func (r *Router) Dispatch(ctx context.Context, session Session, packet protocol.Packet) error {
	r.mu.RLock()
	handler, ok := r.handlers[packet.MsgID]
	r.mu.RUnlock()
	if !ok {
		return fmt.Errorf("%w: msg_id=%d", ErrHandlerNotFound, packet.MsgID)
	}
	return handler(ctx, session, packet)
}
