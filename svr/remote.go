package svr

import (
	"dk/base"
	"net"
	"sync"
	"time"
)

type accessToken struct {
	dst     string       //DKC所在网络的名称
	addr    *net.TCPAddr //需访问的DKC网络内的主机地址
	ref     int          //引用计数
	created time.Time
	updated time.Time
}

type remoteAdmin struct {
	tokens  map[string]*accessToken
	maxIdle time.Duration
	maxLife time.Duration
	sync.Mutex
}

func (ra *remoteAdmin) Register(src, dst string, addr *net.TCPAddr) {
	ra.Lock()
	defer ra.Unlock()
	ra.tokens[src] = &accessToken{
		dst:     dst,
		addr:    addr,
		created: time.Now(),
		updated: time.Now(),
	}
}

func (ra *remoteAdmin) Lookup(ip net.IP) *accessToken {
	ra.Lock()
	defer ra.Unlock()
	return ra.tokens[ip.String()]
}

func (ra *remoteAdmin) Connect(conn net.Conn) *accessToken {
	ra.Lock()
	defer ra.Unlock()
	p := conn.RemoteAddr().(*net.TCPAddr)
	src := p.IP.String()
	t := ra.tokens[src]
	if t == nil {
		return nil
	}
	t.ref++
	ra.tokens[src] = t
	return t
}

func (ra *remoteAdmin) Disconnect(conn net.Conn) {
	ra.Lock()
	defer ra.Unlock()
	p := conn.RemoteAddr().(*net.TCPAddr)
	src := p.IP.String()
	t := ra.tokens[src]
	if t != nil {
		t.ref--
		ra.tokens[src] = t
	}
}

func (ra *remoteAdmin) Init(cf Config) {
	ra.Lock()
	defer ra.Unlock()
	ra.tokens = make(map[string]*accessToken)
	ra.maxIdle = time.Duration(cf.IdleClose) * time.Second
	ra.maxLife = time.Duration(cf.AuthTime) * time.Second
	go func() {
		for {
			time.Sleep(5 * time.Second)
			func() {
				ra.Lock()
				defer ra.Unlock()
				for s, t := range ra.tokens {
					if t.ref > 0 {
						t.updated = time.Now()
						ra.tokens[s] = t
						continue
					}
					remove := false
					if time.Since(t.updated) > ra.maxIdle {
						base.Dbg(`token for "%v" idle timeout`, s)
						remove = true
					}
					if time.Since(t.created) > ra.maxLife {
						base.Dbg(`token for "%v" end of life`, s)
						remove = true
					}
					if remove {
						delete(ra.tokens, s)
					}
				}
			}()
		}
	}()
}

var ra remoteAdmin
