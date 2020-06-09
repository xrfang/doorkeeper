package cli

import (
	"dk/base"
	"fmt"
	"net"
	"sync"
	"time"
)

type Config struct {
	Name    string `yaml:"name"`
	SvrHost string `yaml:"svr_host"`
	SvrPort int    `yaml:"svr_port"`
	Auth    string `yaml:"auth"`
	MaxConn int    `yaml:"max_conn"`
}

type proxy struct {
	live bool
	serv *net.TCPConn
	dsts map[string]*net.TCPConn
	sync.Mutex
}

func (p *proxy) isAlive() bool {
	p.Lock()
	defer p.Unlock()
	return p.live
}

func (p *proxy) setLive(stat bool) {
	p.Lock()
	p.live = stat
	p.Unlock()
}

func (p *proxy) addConn(src, dst *net.TCPAddr) (c *net.TCPConn) {
	p.Lock()
	defer func() {
		if e := recover(); e != nil {
			//TODO: log error
			fmt.Println(trace("%v", e))
		}
		p.Unlock()
	}()
	fmt.Printf("addConn: src=%v; dst=%v\n", src, dst) //TODO: add logging
	conn, err := net.Dial("tcp", dst.String())
	assert(err)
	c = conn.(*net.TCPConn)
	p.dsts[src.String()] = c
	go func(conn *net.TCPConn) {
		defer func() {
			e := recover()
			//TODO: log error
			fmt.Println(trace("%v", e))
			p.delConn(src)
		}()
		buf := make([]byte, 4096)
		for p.isAlive() {
			n, err := conn.Read(buf)
			assert(err)
			c := base.Chunk{
				Type: base.CT_DAT,
				Src:  dst,
				Dst:  src,
				Buf:  buf[:n],
			}
			assert(c.Send(p.serv))
		}
	}(c)
	return
}

func (p *proxy) delConn(addr *net.TCPAddr) {
	p.Lock()
	defer p.Unlock()
	fmt.Println("TODO: cli.proxy.delConn:", addr)
	tag := addr.String()
	conn := p.dsts[tag]
	if conn != nil {
		conn.Close()
		delete(p.dsts, tag)
	}
}

func (p *proxy) delConns(ip string) {
	p.Lock()
	defer p.Unlock()
	for src, c := range p.dsts {
		host, _, _ := net.SplitHostPort(src)
		if host == ip {
			c.Close()
			fmt.Println("TODO: cli.proxy.delConns:", ip)
			delete(p.dsts, host)
		}
	}
}

func (p *proxy) run(cf Config) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println(trace("TODO: proxy.run: %v", e))
		}
		p.setLive(false)
	}()
	addr := fmt.Sprintf("%s:%d", cf.SvrHost, cf.SvrPort)
	conn, err := net.Dial("tcp", addr)
	assert(err)
	fmt.Println("TODO: client connected")
	p.Lock()
	p.serv = conn.(*net.TCPConn)
	p.Unlock()
	p.setLive(true)
	handshake := base.Authenticate(nil, cf.Name, cf.Auth)
	p.serv.Write(handshake)
	for p.isAlive() {
		var c base.Chunk
		c.Recv(p.serv)
		switch c.Type {
		case base.CT_CLS:
			p.delConns(c.Src.IP.String())
		case base.CT_DAT:
			src := c.Src.String()
			dst := p.dsts[src]
			if dst == nil {
				dst = p.addConn(c.Src, c.Dst)
			}
			if dst != nil {
				_, err := dst.Write(c.Buf)
				if err != nil {
					fmt.Printf("TODO: relay: %v\n", err)
					//TODO: log data transfer error
					p.delConn(c.Src)
				}
			}
		case base.CT_QRY:
			fmt.Println("TODO: handle CT_QRY")
		}
	}
}

func (p *proxy) init() {
	p.Lock()
	defer p.Unlock()
	p.dsts = make(map[string]*net.TCPConn)
	//TODO: send data, clean up, heartbeat etc.
}

func (p *proxy) Run(cf Config) {
	p.init()
	for {
		p.run(cf)
		time.Sleep(time.Second)
	}
}

func Start(cf Config) {
	var p proxy
	p.Run(cf)
}
