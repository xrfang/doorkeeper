package svr

import (
	"net"
	"sync"
	"time"

	"dk/base"
)

type serviceMgr struct {
	auth      map[string]string
	handshake time.Duration
	backends  map[string]*backend
	sync.Mutex
}

func (sm *serviceMgr) getBackend(name string) *backend {
	sm.Lock()
	defer sm.Unlock()
	return sm.backends[name]
}

func (sm *serviceMgr) Authenticate(mac []byte) string {
	if len(mac) != 32 {
		return ""
	}
	for name, key := range sm.auth {
		var match bool
		res := base.Authenticate(mac[:16], name, key)
		for i, c := range mac {
			match = res[i] == c
			if !match {
				break
			}
		}
		if match {
			return name
		}
	}
	return ""
}

func (sm *serviceMgr) Init(cf Config) {
	sm.Lock()
	defer sm.Unlock()
	sm.auth = cf.Auth
	sm.handshake = time.Duration(cf.Handshake) * time.Second
	sm.backends = make(map[string]*backend)
	go func() {
		for {
			time.Sleep(time.Minute)
			func() {
				sm.Lock()
				defer sm.Unlock()
				base.Dbg("health check %d backends", len(sm.backends))
				for n, b := range sm.backends {
					alive := b.isAlive()
					base.Dbg(`backend "%s": alive=%v`, n, alive)
					if !alive {
						b.destroy()
						delete(sm.backends, n)
						base.Log(`backend "%s" offline, removed`, n)
					}
				}
			}()
		}
	}()
}

func (sm *serviceMgr) Validate(conn net.Conn) {
	ra := conn.RemoteAddr().String()
	assert(conn.SetDeadline(time.Now().Add(sm.handshake)))
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		err, ok := err.(net.Error)
		if !ok || !err.Timeout() {
			base.Dbg("sm.Validate(%s): %v", ra, err)
		} else {
			base.Dbg("sm.Validate(%s): %v", ra, err)
		}
		base.Log(`backend "%s" refused (handshake failed)`, ra)
		conn.Close()
		return
	}
	assert(conn.SetReadDeadline(time.Time{}))
	name := sm.Authenticate(buf[:n])
	if name == "" {
		base.Dbg("sm.Validate(%s): invalid hmac [%x]", ra, buf[:n])
		base.Log(`backend "%s" refused (handshake failed)`, ra)
		conn.Close()
		return
	}
	base.Log("backend %s connected", ra)
	sm.Lock()
	defer sm.Unlock()
	b := sm.backends[name]
	if b != nil {
		b.destroy()
		delete(sm.backends, name)
	}
	b = &backend{
		serv: conn.(*net.TCPConn),
		clis: make(map[string]*net.TCPConn),
	}
	go func() {
		lo := &net.TCPAddr{IP: net.ParseIP("127.0.0.1")}
		for {
			time.Sleep(30 * time.Second)
			c := base.Chunk{Type: base.CT_PNG, Src: lo, Dst: lo}
			if err := c.Send(b.serv); err != nil {
				b.setLive(false)
				base.Log(`ping backend "%s" failed`, ra)
				base.Dbg("ERROR: %v", err)
				return
			}
			base.Dbg(`ping backend "%s" ok`, name)
		}
	}()
	sm.backends[name] = b
	b.Run()
	return
}

func (sm *serviceMgr) Relay(conn net.Conn, token *accessToken) {
	b := sm.getBackend(token.dst)
	if b == nil {
		base.Log(`sm.Relay: backend "%s" not found`, token.dst)
		ra.Disconnect(conn)
		conn.Close()
		return
	}
	b.addConn(conn)
}

var sm serviceMgr
