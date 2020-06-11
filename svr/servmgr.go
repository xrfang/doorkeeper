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
				for n, b := range sm.backends {
					if !b.isAlive() {
						b.destroy()
						delete(sm.backends, n)
						base.Dbg(`backend "%s" offline, removed`, n)
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
		for {
			time.Sleep(30 * time.Second)
			c := base.Chunk{Type: base.CT_PNG}
			if c.Send(b.serv) != nil {
				base.Dbg(`ping backend "%s" failed`, ra)
				b.setLive(false)
				return
			}
		}
	}()
	sm.backends[name] = b
	b.Run()
	return
}

func (sm *serviceMgr) Relay(conn net.Conn, token *accessToken) {
	sm.Lock()
	defer sm.Unlock()
	b := sm.backends[token.dst]
	if b == nil {
		base.Log(`sm.Relay: backend "%s" not found`, token.dst)
		conn.Close()
		return
	}
	b.addConn(conn)
}

var sm serviceMgr
