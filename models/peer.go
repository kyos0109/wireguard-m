package models

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
)

type Peer struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	PublicKey       string `json:"public_key"`
	ServerPublicKey string `json:"server_public_key"`
	PrivateKey      string `json:"private_key"`
	Address         string `json:"address"`
	AllowedIPs      string `json:"allowed_ips"`
	PresharedKey    string `json:"preshared_key,omitempty"`
}

type PeerStore struct {
	dataFile string
}

var PeersStorePath *string

func NewPeerStore() *PeerStore {
	return &PeerStore{dataFile: *PeersStorePath}
}

func (ps *PeerStore) LoadPeers() (*DevicePeers, error) {
	if _, err := os.Stat(ps.dataFile); os.IsNotExist(err) {
		return &DevicePeers{}, nil
	}
	data, err := os.ReadFile(ps.dataFile)
	if err != nil {
		return nil, err
	}

	var devicePeers DevicePeers
	if len(data) > 0 {
		if err := json.Unmarshal(data, &devicePeers); err != nil {
			return nil, err
		}
	}

	return &devicePeers, nil
}

func (ps *PeerStore) SavePeers(devPeers *DevicePeers) error {
	data, err := json.MarshalIndent(devPeers, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(ps.dataFile, data, 0644)
}

func (ps *PeerStore) getNextPeerID(peers []Peer) int {
	maxID := 0
	for _, peer := range peers {
		if peer.ID > maxID {
			maxID = peer.ID
		}
	}

	return maxID + 1
}

func (ps *PeerStore) AddPeer(newDevPeers *DevicePeers) error {
	devsPeers, err := ps.LoadPeers()
	if err != nil {
		return err
	}

	d := *devsPeers
	for dev, peers := range *newDevPeers {
		for i := range peers {
			peers[i].ID = ps.getNextPeerID(peers)

			d[dev] = append(d[dev], peers[i])
		}
	}

	return ps.SavePeers(&d)
}

func (ps *PeerStore) DeletePeer(interfaceName string, id int) error {
	devsPeers, err := ps.LoadPeers()
	if err != nil {
		return err
	}

	index := -1
	d := *devsPeers
	for dev, peers := range *devsPeers {
		if dev == interfaceName {
			for i, peer := range peers {
				if peer.ID == id {
					index = i
					break
				}
			}
		}
	}
	if index == -1 {
		return errors.New("找不到指定的 Peer")
	}

	d[interfaceName] = slices.Delete(d[interfaceName], index, len(d[interfaceName]))
	return ps.SavePeers(&d)
}

func (ps *PeerStore) GetPeerByID(interfaceName string, id int) (*Peer, error) {
	devsPeers, err := ps.LoadPeers()
	if err != nil {
		return nil, err
	}

	for dev, peers := range *devsPeers {
		if dev == interfaceName {
			for _, peer := range peers {
				if peer.ID == id {
					return &peer, nil
				}
			}
		}
	}

	return nil, errors.New("找不到指定的 Peer")
}
