package svr

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type serviceMgr struct {
	handshake time.Duration
	backends  map[string]*backend
	sync.Mutex
}

func (sm *serviceMgr) Init(cf Config) {
	sm.Lock()
	defer sm.Unlock()
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
					}
				}
			}()
		}
	}()
}

func (sm *serviceMgr) Validate(conn net.Conn) {
	assert(conn.SetDeadline(time.Now().Add(sm.handshake)))
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		err, ok := err.(net.Error)
		if !ok || !err.Timeout() {
			fmt.Printf("TODO: log error %v\n", err)
		} else {
			fmt.Printf("TODO: log timeout waiting handshake\n")
		}
		conn.Close()
		return
	}
	assert(conn.SetReadDeadline(time.Time{}))
	//TODO: HMAC-SHA256 validation
	if string(buf[:n]) != "HELLO" {
		conn.Close()
		return
	}
	name := conn.RemoteAddr().String() //TODO: 使用客户端名称而非IP地址
	sm.Lock()
	defer sm.Unlock()
	b := sm.backends[name]
	if b != nil {
		b.destroy()
		delete(sm.backends, name)
	}
	b = &backend{
		serv: conn.(*net.TCPConn),
		send: make(chan chunk),
		clis: make(map[string]*net.TCPConn),
	}
	sm.backends[name] = b
	b.Run()
	return
}

func (sm *serviceMgr) Relay(conn net.Conn, token *accessToken) {
	sm.Lock()
	defer sm.Unlock()
	b := sm.backends[token.dst]
	if b == nil {
		//TODO: log
		conn.Close()
		return
	}
	b.addClient(conn)
}

var sm serviceMgr
