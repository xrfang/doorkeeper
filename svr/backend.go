package svr

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"dk/base"
)

type backend struct {
	live bool
	serv *net.TCPConn
	clis map[string]*net.TCPConn
	sync.Mutex
}

func (b *backend) destroy() {
	b.Lock()
	defer b.Unlock()
	b.serv.Close()
	for _, c := range b.clis {
		c.Close()
	}
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
				fmt.Println(trace("TODO: client recv: %v", e))
				b.delConn(c)
			}
		}()
		src := c.RemoteAddr().(*net.TCPAddr)
		at := ra.Lookup(src.IP)
		if at == nil {
			panic(errors.New("no access"))
		}
		for {
			buf := make([]byte, 4096)
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
				fmt.Println(trace("TODO: %v", e))
				b.setLive(false)
			}
		}()
		for b.isAlive() {
			var c base.Chunk
			assert(c.Recv(b.serv))
			tag := c.Dst.String()
			cli := b.clis[tag]
			if cli == nil {
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
				cli.Close()         //close remote connection
				delete(b.clis, tag) //unregister connection
			case base.CT_DAT: //data transfer
				_, err := cli.Write(c.Buf)
				assert(err)
			}
		}
	}()
}
