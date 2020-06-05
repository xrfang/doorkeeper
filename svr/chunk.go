package svr

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

type chunk struct {
	src *net.TCPAddr
	dst *net.TCPAddr
	buf []byte
}

func (c chunk) Serialize() []byte {
	buf := []byte{0, 0}
	buf = append(buf, []byte(c.src.String())...)
	buf = append(buf, '\n')
	buf = append(buf, []byte(c.dst.String())...)
	if len(c.buf) > 0 {
		buf = append(buf, '\n')
		buf = append(buf, c.buf...)
	}
	binary.BigEndian.PutUint16(buf, uint16(len(buf)-2))
	return buf
}

func parseAddr(addr string) *net.TCPAddr {
	h, p, err := net.SplitHostPort(addr)
	assert(err)
	ip := net.ParseIP(h)
	if ip == nil {
		panic("invalid IP address")
	}
	port, _ := strconv.Atoi(p)
	return &net.TCPAddr{IP: ip, Port: port}
}

//从远端读入chunk数据
func (b *backend) recvChunk() (c *chunk) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println(trace("recvChunk: %v", e))
			c = nil
		}
	}()
	buf := make([]byte, 4096)
	_, err := io.ReadFull(b.serv, buf[:2])
	assert(err)
	clen := int(binary.BigEndian.Uint16(buf[:2]))
	data := make([]byte, clen)
	b.serv.SetReadDeadline(time.Now().Add(time.Minute))
	for {
		n, err := b.serv.Read(buf)
		assert(err)
		data = append(data, buf[:n]...)
		clen -= n
		if clen <= 0 {
			break
		}
	}
	b.serv.SetDeadline(time.Time{})
	cs := strings.SplitN(string(data), "\n", 3)
	if len(cs) != 3 {
		panic(errors.New("invalid chunk format"))
	}
	var ck chunk
	ck.src = parseAddr(cs[0])
	ck.dst = parseAddr(cs[1])
	ck.buf = []byte(cs[2])
	return &ck
}

func (b *backend) sendChunk(c chunk) {
	assert(b.serv.SetWriteDeadline(time.Now().Add(time.Minute)))
	_, err := b.serv.Write(c.Serialize())
	assert(err)
	assert(b.serv.SetWriteDeadline(time.Time{}))
}

//从本地端读入原始数据，装配成chunk
func recvData(conn *net.TCPConn) (c *chunk, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
			conn.Close()
		}
	}()
	src := conn.RemoteAddr().(*net.TCPAddr)
	at := ra.Lookup(src.IP)
	if at == nil {
		return nil, errors.New("no access")
	}
	ck := chunk{src: src, dst: at.addr}
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	assert(err)
	ck.buf = buf[:n]
	return &ck, nil
}
