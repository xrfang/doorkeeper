package svr

import (
	"errors"
	"io"
	"net"
	"strings"
	"sync"

	"dk/base"
)

type backend struct {
	live bool
	serv *net.TCPConn
	clis map[string]*net.TCPConn
	sync.Mutex
}

func (b *backend) flush() {
	b.Lock()
	defer b.Unlock()
	for _, c := range b.clis {
		c.Close()
	}
	b.clis = make(map[string]*net.TCPConn)
}

func (b *backend) destroy() {
	b.Lock()
	defer b.Unlock()
	b.serv.Close()
	for _, c := range b.clis {
		c.Close()
	}
	b.clis = make(map[string]*net.TCPConn)
}

func (b *backend) setLive(live bool) {
	b.Lock()
	defer b.Unlock()
	b.live = live
}

func (b *backend) isAlive() bool {
	b.Lock()
	defer b.Unlock()
	return b.live
}

func (b *backend) listConns() []string {
	b.Lock()
	defer b.Unlock()
	var cns []string
	for c := range b.clis {
		cns = append(cns, c)
	}
	return cns
}

func (b *backend) getConn(tag string) *net.TCPConn {
	b.Lock()
	defer b.Unlock()
	return b.clis[tag]
}

func (b *backend) delConn(conn *net.TCPConn) {
	b.Lock()
	defer b.Unlock()
	tag := conn.RemoteAddr().String()
	delete(b.clis, tag)
	conn.Close()
}

func (b *backend) addConn(conn net.Conn) {
	b.Lock()
	defer b.Unlock()
	tag := conn.RemoteAddr().String()
	c := b.clis[tag]
	if c != nil {
		c.Close()
	}
	b.clis[tag] = conn.(*net.TCPConn)
	go func(c *net.TCPConn) { //从本地端读入原始数据，装配成chunk
		defer func() {
			if e := recover(); e != nil {
				if e == io.EOF {
					base.Dbg(`remote "%s" disconnected`, tag)
				} else {
					base.Err("%v", e)
				}
				ra.Disconnect(c)
				b.delConn(c)
			}
		}()
		src := c.RemoteAddr().(*net.TCPAddr)
		at := ra.Lookup(src.IP.String())
		if at == nil {
			panic(errors.New("no access"))
		}
		buf := make([]byte, 4000) //buf长度必须小于4096，因为包头表示长度用了12bit
		for {
			n, err := c.Read(buf)
			assert(err)
			c := base.Chunk{
				Type: base.CT_DAT,
				Src:  src,
				Dst:  at.addr,
				Buf:  buf[:n],
			}
			c.Send(b.serv)
		}
	}(conn.(*net.TCPConn))
}

func (b *backend) Run() {
	b.setLive(true)
	go func() {
		defer func() {
			if e := recover(); e != nil {
				if e == io.EOF {
					ra := b.serv.RemoteAddr().String()
					base.Log("backend %s disconnected", ra)
				} else {
					base.Err("backend.Run: %v", e)
				}
				b.setLive(false)
			}
		}()
		for b.isAlive() {
			var c base.Chunk
			err := c.Recv(b.serv)
			if err != nil {
				addr := b.serv.RemoteAddr().String()
				ne, ok := err.(net.Error)
				if ok && ne.Timeout() {
					base.Dbg(`backend "%s" recv timeout`, addr)
					continue
				}
				b.setLive(false)
				if err == io.EOF {
					base.Log(`backend "%s" disconnected`, addr)
				} else {
					base.Err("%v", err)
				}
				break
			}
			tag := c.Dst.String()
			cli := b.getConn(tag)
			if cli == nil && c.Type != base.CT_QRY && c.Type != base.CT_PNG {
				rep := base.Chunk{
					Type: base.CT_CLS, //local disconnected, notify DKC
					Src:  c.Dst,
					Dst:  c.Src,
				}
				rep.Send(b.serv)
				continue
			}
			switch c.Type {
			case base.CT_CLS: //DKC disconnected
				b.delConn(cli) //close remote connection
			case base.CT_DAT: //data transfer
				_, err := cli.Write(c.Buf)
				assert(err)
			case base.CT_QRY:
				ch := getChan(c.Dst.IP.String(), c.Dst.Port)
				if ch == nil {
					base.Dbg(`CT_QRY recipient "%s" not found, dropped`, c.Dst)
					continue
				}
				ch <- c.Buf
			case base.CT_PNG:
				s := strings.SplitN(string(c.Buf), "=", 2)
				base.Dbg(`backend "%s" reply: conns=%s`, s[0], s[1])
			}
		}
	}()
}
