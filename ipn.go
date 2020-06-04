package ipn

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"

	"tailscale.com/net/interfaces"
)

type Peer struct {
	Host string `json:"HostName"`
	Addr net.IP `json:"TailAddr"`
	OS   string `json:"OS"`
}

func Me() (Peer, error) {
	host, err := os.Hostname()
	if err != nil {
		return Peer{}, fmt.Errorf("ipn: failed to get machine hostname: %w", err)
	}
	addr, _, err := interfaces.Tailscale()
	if err != nil {
		return Peer{}, fmt.Errorf("ipn: failed to get tailscale interface: %w", err)
	}
	return Peer{
		Host: host,
		Addr: addr,
		OS:   runtime.GOOS,
	}, nil
}

func QueryForPeers() ([]Peer, error) {
	out, err := exec.Command("tailscale", "status", "--json").Output()
	if err != nil {
		return nil, fmt.Errorf("ipn: failed to exec tailscale cmd: %w", err)
	}
	var buf struct {
		Peers map[string]Peer `json:"Peer"`
	}
	if err = json.Unmarshal(out, &buf); err != nil {
		return nil, fmt.Errorf("ipn: failed to unmarshal tailscale status: %w", err)
	}
	peers := make([]Peer, 0, len(buf.Peers))
	for _, p := range buf.Peers {
		peers = append(peers, p)
	}
	return peers, nil
}

func (p *Peer) ListenOn(network, port string) (net.Listener, error) {
	addr := net.JoinHostPort(p.Addr.String(), port)
	return net.Listen(network, addr)
}
