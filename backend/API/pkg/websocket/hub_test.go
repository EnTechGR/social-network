package websocket

import (
	"bytes"
	"testing"
)

func testClient(userID, nickname string) *Client {
	return &Client{
		UserID:   userID,
		Nickname: nickname,
		Send:     make(chan []byte, 8),
	}
}

func drainSendQueue(c *Client) {
	for {
		select {
		case <-c.Send:
		default:
			return
		}
	}
}

func TestHub_MultipleConnectionsSameUser(t *testing.T) {
	hub := NewHub()

	c1 := testClient("u1", "alice")
	c2 := testClient("u1", "alice")

	hub.registerClient(c1)
	drainSendQueue(c1) // online_users_list sync payload

	hub.registerClient(c2)
	drainSendQueue(c2) // online_users_list sync payload

	if !hub.IsUserOnline("u1") {
		t.Fatalf("expected user to be online after registering clients")
	}
	if got := len(hub.Clients["u1"]); got != 2 {
		t.Fatalf("expected 2 active connections for user, got %d", got)
	}

	hub.unregisterClient(c1)
	if !hub.IsUserOnline("u1") {
		t.Fatalf("expected user to remain online while one connection is still active")
	}
	if got := len(hub.Clients["u1"]); got != 1 {
		t.Fatalf("expected 1 active connection after first unregister, got %d", got)
	}

	hub.unregisterClient(c2)
	if hub.IsUserOnline("u1") {
		t.Fatalf("expected user to be offline after removing all connections")
	}
}

func TestHub_BroadcastToUserFanoutAllConnections(t *testing.T) {
	hub := NewHub()
	c1 := testClient("u1", "alice")
	c2 := testClient("u1", "alice")

	hub.mu.Lock()
	hub.Clients["u1"] = map[*Client]struct{}{
		c1: {},
		c2: {},
	}
	hub.mu.Unlock()

	want := []byte(`{"type":"chat"}`)
	hub.BroadcastToUser("u1", want)

	got1 := <-c1.Send
	got2 := <-c2.Send

	if !bytes.Equal(got1, want) {
		t.Fatalf("first connection got unexpected payload: %q", got1)
	}
	if !bytes.Equal(got2, want) {
		t.Fatalf("second connection got unexpected payload: %q", got2)
	}
}
