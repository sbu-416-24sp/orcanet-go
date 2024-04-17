package api

import (
	"log"
	"sync"
)

// PeerInfo struct holds the information about a peer.
type PeerInfo struct {
	Location    string `json:"location"`
	Latency     string `json:"latency"`
	PeerID      string `json:"peerID"`
	Connection  string `json:"connection"`
	OpenStreams string `json:"openStreams"`
	FlagUrl     string `json:"flagUrl"`
}

// PeerStorage is a concurrent safe storage for peer information.
type PeerStorage struct {
	mutex sync.RWMutex
	peers map[string]PeerInfo
}

// NewPeerStorage initializes a new instance of PeerStorage.
func NewPeerStorage() *PeerStorage {
	return &PeerStorage{
		peers: make(map[string]PeerInfo),
	}
}

// AddPeer adds a new peer to the storage.
func (ps *PeerStorage) AddPeer(peer PeerInfo) {
	if ps == nil {

		log.Println("PeerStorage is nil")
		return

	}
	if peer.PeerID == "" {
		// Handle invalid peerID
		return
	}
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	ps.peers[peer.PeerID] = peer
	log.Println("Succesfully added a peer")
}

// GetPeer retrieves a peer's information by its ID.
func (ps *PeerStorage) GetPeer(peerID string) (PeerInfo, bool) {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()

	peer, exists := ps.peers[peerID]
	return peer, exists
}

// GetAllPeers returns all peers in the storage.
func (ps *PeerStorage) GetAllPeers() []PeerInfo {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()

	allPeers := make([]PeerInfo, 0, len(ps.peers))
	for _, peer := range ps.peers {
		allPeers = append(allPeers, peer)
	}
	return allPeers
}

// UpdatePeer updates an existing peer's information.
func (ps *PeerStorage) UpdatePeer(peerID string, newPeerInfo PeerInfo) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	if _, exists := ps.peers[peerID]; exists {
		ps.peers[peerID] = newPeerInfo
	}
}

func (ps *PeerStorage) RemovePeer(peerID string) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	delete(ps.peers, peerID)
}
