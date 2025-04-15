package models

type Config struct {
	Server    struct{}
	Wireguard struct {
		DNS                 string `json:"dns"`
		Endpoint            string `json:"endpoint"`
		InterfaceName       string `json:"interface_name"`
		ServerPublicKey     string `json:"server_public_key"`
		PersistentKeepalive int    `json:"persistent_keepalive"`
	}
}

var ConfigPath *string
