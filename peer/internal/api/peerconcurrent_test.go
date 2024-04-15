package api

import (
	"fmt"
	"sync"
	"testing"
)

func TestPeerStorageConcurrency(t *testing.T) {
	ps := NewPeerStorage()
	var wg sync.WaitGroup

	// Simulate concurrent add operations
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(peerID string) {
			defer wg.Done()
			ps.AddPeer(PeerInfo{PeerID: peerID})
		}(fmt.Sprintf("peer%d", i))
	}

	wg.Wait()

	// Verify that all peers were added
	for i := 0; i < 100; i++ {
		if _, exists := ps.GetPeer(fmt.Sprintf("peer%d", i)); !exists {
			t.Errorf("Peer %d was not added", i)
		}
	}
}
