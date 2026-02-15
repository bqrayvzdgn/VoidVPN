package wireguard

type TunnelConfig struct {
	PrivateKey         string
	Address            string
	DNS                []string
	MTU                int
	PeerPublicKey      string
	PeerEndpoint       string
	PeerAllowedIPs     []string
	PeerPresharedKey   string
	PersistentKeepalive int
}
