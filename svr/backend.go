package svr

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

type backend struct {
	live bool
	serv *net.TCPConn
	send chan chunk
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

func (b *backend) delClient(conn *net.TCPConn) {
	b.Lock()
	defer b.Unlock()
	tag := conn.RemoteAddr().String()
	delete(b.clis, tag)
	conn.Close()
}

func (b *backend) addClient(conn net.Conn) {
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
				b.delClient(c)
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
			b.send <- chunk{
				src: src,
				dst: at.addr,
				buf: buf[:n],
			}
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
			c := b.recvChunk()
			if c == nil {
				b.setLive(false)
				break
			}
			tag := c.dst.String()
			cli := b.clis[tag]
			if cli == nil { //local disconnected, notify DKC
				b.sendChunk(chunk{src: c.dst, dst: c.src})
				continue
			}
			if len(c.buf) == 0 { //DKC disconnected
				cli.Close()         //close remote connection
				delete(b.clis, tag) //unregister connection
				continue
			}
			_, err := cli.Write(c.buf)
			assert(err)
		}
	}()
	go func() {
		defer func() {
			if e := recover(); e != nil {
				fmt.Println(trace("TODO: %v", e))
				b.setLive(false)
			}
		}()
		for b.isAlive() {
			select {
			case c := <-b.send:
				b.sendChunk(c)
			case <-time.After(time.Second):
			}
		}
	}()
}
