package api

import (
	"testing"
)

func TestAddPeer(t *testing.T) {
	ps := NewPeerStorage()
	peer := PeerInfo{
		PeerID: "testPeer",
		// Populate other fields as necessary
	}

	ps.AddPeer(peer)

	if _, exists := ps.GetPeer("testPeer"); !exists {
		t.Errorf("AddPeer failed to add the peer")
	}
}
