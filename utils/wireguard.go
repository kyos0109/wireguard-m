package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/kyos0109/WireGuard-M/models"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const (
	PersistentKeepalive = 25
)

var (
	client  *wgctrl.Client
	once    sync.Once
	initErr error
)

func newWGctrl() (*wgctrl.Client, error) {
	once.Do(func() {
		client, initErr = wgctrl.New()
	})
	return client, initErr
}

func CloseWGClient() {
	if client != nil {
		if err := client.Close(); err != nil {
			log.Printf("Error closing wgctrl client: %v", err)
		} else {
			log.Println("wgctrl client closed.")
		}
	}
}

func ListDevices() (interfaces []*models.WireGuardInterface, err error) {
	client, err := newWGctrl()
	if err != nil {
		return nil, err
	}

	ds, err := client.Devices()
	if err != nil {
		return nil, err
	}

	interfaces = make([]*models.WireGuardInterface, len(ds))
	for i, d := range ds {
		interfaces[i] = &models.WireGuardInterface{
			Name:            d.Name,
			ServerPublicKey: d.PublicKey.String(),
		}
	}

	return interfaces, nil
}

func AddPeerToInterface(interfaceName string, peer *models.Peer) error {
	client, err := newWGctrl()
	if err != nil {
		return err
	}

	_, err = client.Device(interfaceName)
	if err != nil {
		return err
	}

	peerKey, err := wgtypes.ParseKey(peer.PublicKey)
	if err != nil {
		return err
	}

	ipnet, err := parseAllowIPs(peer.AllowedIPs)
	if err != nil {
		return err
	}

	duration := time.Duration(PersistentKeepalive) * time.Second

	cfg := wgtypes.PeerConfig{
		PublicKey:                   peerKey,
		Remove:                      false,
		ReplaceAllowedIPs:           true,
		AllowedIPs:                  ipnet,
		PersistentKeepaliveInterval: &duration,
	}

	config := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{cfg},
	}
	return client.ConfigureDevice(interfaceName, config)
}

func parseAllowIPs(s string) ([]net.IPNet, error) {
	parts := strings.Split(s, ",")
	var ipNets []net.IPNet
	for _, part := range parts {
		cidrStr := strings.TrimSpace(part)
		_, ipnet, err := net.ParseCIDR(cidrStr)
		if err != nil {
			return nil, fmt.Errorf("cannot parser %q: %v", cidrStr, err)
		}
		ipNets = append(ipNets, *ipnet)
	}
	return ipNets, nil
}

func RemovePeerFromInterface(interfaceName, publicKey string) error {
	client, err := newWGctrl()
	if err != nil {
		return err
	}

	pkey, err := wgtypes.ParseKey(publicKey)
	if err != nil {
		return err
	}
	cfg := wgtypes.PeerConfig{
		PublicKey: pkey,
		Remove:    true,
	}
	config := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{cfg},
	}
	return client.ConfigureDevice(interfaceName, config)
}

func GeneratePeerConfig(peer *models.Peer) (string, error) {
	cfg, err := loadServerConfig()
	if err != nil {
		return "", err
	}

	configTemplate := `[Interface]
PrivateKey = %s
Address = %s
DNS = %s

[Peer]
PublicKey = %s
Endpoint = %s
AllowedIPs = 0.0.0.0/0, ::/0
PersistentKeepalive = %d
`
	config := fmt.Sprintf(configTemplate, peer.PrivateKey, peer.Address, cfg.Wireguard.DNS, peer.ServerPublicKey, cfg.Wireguard.Endpoint, PersistentKeepalive)
	return config, nil
}

func loadServerConfig() (*models.Config, error) {
	if _, err := os.Stat(*models.ConfigPath); os.IsNotExist(err) {
		return &models.Config{}, nil
	}
	data, err := os.ReadFile(*models.ConfigPath)
	if err != nil {
		return nil, err
	}

	var cfg models.Config
	if len(data) > 0 {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}
	}

	return &cfg, nil
}
