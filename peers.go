package ipn

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

// A Peer is another machine connected on a users ipn.
type Peer struct {
	Hostname string `json:"HostName"`
	Addr     net.IP `json:"TailAddr"`
	OS       string `json:"OS"`
}

// Me returns a peer object describing this node.
func Me() (Peer, error) {
	host, err := os.Hostname()
	if err != nil {
		return Peer{}, err
	}
	addr, _, err := NetInterface()
	if err != nil {
		return Peer{}, err
	}
	return Peer{
		Hostname: host,
		Addr:     addr,
		OS:       runtime.GOOS,
	}, nil
}

// A Queryer can be used to query for peers.
type Queryer interface {
	Query() (map[string]Peer, error)
}

// DefaultQueryer creates a new default queryer.
func DefaultQueryer() Queryer {
	return &defaultQueryer{}
}

type defaultQueryer struct{}

func (dq *defaultQueryer) parseCmdOutput(output []byte) (map[string]Peer, error) {
	var buf struct {
		Peers map[string]Peer `json:"Peer"`
	}
	if err := json.Unmarshal(output, &buf); err != nil {
		return nil, fmt.Errorf("ipn: failed to parse command output: %v", err)
	}
	peers := make(map[string]Peer, len(buf.Peers))
	for _, p := range buf.Peers {
		peers[p.Addr.String()] = p
	}
	return peers, nil
}

func (dq *defaultQueryer) Query() (map[string]Peer, error) {
	out, err := exec.Command("tailscale", "status", "--json").Output()
	if err != nil {
		return nil, err
	}
	return dq.parseCmdOutput(out)
}

// CachedQueryer creates a new cached queryer.
func CachedQueryer(ttl time.Duration) Queryer {
	return &cachedQueryer{q: &defaultQueryer{}, ttl: ttl}
}

type cachedQueryer struct {
	q     Queryer
	cmu   sync.RWMutex // protects cache
	cache map[string]Peer
	ttl   time.Duration
	last  time.Time
}

// requires rlock
func (cq *cachedQueryer) shouldUseCache() bool {
	expiry := cq.last.Add(cq.ttl)
	return cq.cache != nil && time.Now().Before(expiry)
}

func (cq *cachedQueryer) Query() (map[string]Peer, error) {
	cq.cmu.RLock()
	if cq.shouldUseCache() {
		defer cq.cmu.RUnlock()
		return cq.cache, nil
	}
	cq.cmu.RUnlock()
	cq.cmu.Lock()
	defer cq.cmu.Unlock()
	var err error
	cq.cache, err = cq.q.Query()
	if err != nil {
		return nil, err
	}
	cq.last = time.Now()
	return cq.cache, nil
}
