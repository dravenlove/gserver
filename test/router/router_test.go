package router_test

import (
	"context"
	"testing"

	"openclaw-go/internal/protocol"
	"openclaw-go/internal/router"
)

type mockSession struct{}

func (m *mockSession) ID() string         { return "1" }
func (m *mockSession) RemoteAddr() string { return "127.0.0.1:10000" }
func (m *mockSession) Send(packet protocol.Packet) error {
	return nil
}

func TestDispatch(t *testing.T) {
	r := router.New()

	called := false
	r.Register(1001, func(ctx context.Context, session router.Session, packet protocol.Packet) error {
		called = true
		return nil
	})

	if err := r.Dispatch(context.Background(), &mockSession{}, protocol.Packet{MsgID: 1001}); err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}
	if !called {
		t.Fatal("handler not called")
	}
}

func TestDispatchHandlerNotFound(t *testing.T) {
	r := router.New()

	err := r.Dispatch(context.Background(), &mockSession{}, protocol.Packet{MsgID: 404})
	if err == nil {
		t.Fatal("expected error")
	}
}
