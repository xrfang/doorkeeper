package cli

import (
	"dk/base"
	"fmt"
	"io"
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
	defer p.Unlock()
	base.Dbg("proxy: %s <=> %s", src, dst)
	conn, err := net.Dial("tcp", dst.String())
	if err != nil {
		base.Log("proxy.Dial(%s): %v", dst.String(), err)
		return
	}
	c = conn.(*net.TCPConn)
	p.dsts[src.String()] = c
	go func(conn *net.TCPConn) {
		defer func() {
			e := recover()
			if e == io.EOF {
				base.Dbg(`backend "%s" disconnected`, dst)
			} else {
				base.Err("%v", e)
			}
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
			delete(p.dsts, host)
		}
	}
}

func (p *proxy) run(cf Config) {
	defer func() {
		if e := recover(); e != nil {
			base.Err("proxy.run: %v", e)
		}
		p.setLive(false)
	}()
	addr := fmt.Sprintf("%s:%d", cf.SvrHost, cf.SvrPort)
	conn, err := net.Dial("tcp", addr)
	assert(err)
	base.Log("connected to %s", addr)
	p.Lock()
	p.serv = conn.(*net.TCPConn)
	p.Unlock()
	p.setLive(true)
	handshake := base.Authenticate(nil, cf.Name, cf.Auth)
	p.serv.Write(handshake)
	for p.isAlive() {
		var c base.Chunk
		err := c.Recv(p.serv)
		if err != nil {
			ne, ok := err.(net.Error)
			if ok && ne.Timeout() {
				base.Dbg(`recv(%s): timeout`, addr)
				continue
			}
			p.setLive(false)
			if err == io.EOF {
				base.Dbg(`server "%s" disconnected`, addr)
			} else {
				base.Err("%v", err)
			}
			break
		}
		switch err.(type) {
		case net.Error:
			if err.(net.Error).Timeout() {
				base.Dbg(`server "%s" disconnected`, addr)
			} else {
				base.Err("%v", err)
			}
		case error:
			base.Err("%v", err)
			break
		}
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
					target := dst.LocalAddr().String()
					if err == io.EOF {
						base.Dbg(`target "%s" disconnected`, target)
					} else {
						base.Err("target(%s): %v", target, err)
					}
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
