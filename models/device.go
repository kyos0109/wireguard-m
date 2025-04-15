package models

type DevicePeers map[string][]Peer

type WireGuardInterface struct {
	Name            string `json:"name"`
	ServerPublicKey string `json:"server_public_key"`
}
