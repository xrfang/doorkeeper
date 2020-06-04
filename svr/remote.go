package svr

import (
	"net"
	"sync"
	"time"
)

type accessToken struct {
	dst     net.TCPAddr
	ref     int //引用计数
	created time.Time
	updated time.Time
}

type remoteAdmin struct {
	tokens  map[string]*accessToken
	maxIdle time.Duration
	maxLife time.Duration
	sync.Mutex
}

func (ra *remoteAdmin) Register(src string, dst net.TCPAddr) {
	ra.Lock()
	defer ra.Unlock()
	ra.tokens[src] = &accessToken{dst: dst, created: time.Now()}
}

func (ra *remoteAdmin) Connect(conn net.Conn) bool {
	ra.Lock()
	defer ra.Unlock()
	p := conn.RemoteAddr().(*net.TCPAddr)
	src := p.IP.String()
	t := ra.tokens[src]
	if t == nil {
		return false
	}
	t.ref++
	ra.tokens[src] = t
	return true
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
					if time.Since(t.updated) > ra.maxIdle ||
						time.Since(t.created) > ra.maxLife {
						delete(ra.tokens, s)
					}
				}
			}()
		}
	}()
}

var ra remoteAdmin
