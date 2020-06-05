package base

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

type Chunk struct {
	Src *net.TCPAddr
	Dst *net.TCPAddr
	Buf []byte
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

func (c Chunk) Serialize() []byte {
	buf := []byte{0, 0}
	buf = append(buf, []byte(c.Src.String())...)
	buf = append(buf, '\n')
	buf = append(buf, []byte(c.Dst.String())...)
	if len(c.Buf) > 0 {
		buf = append(buf, '\n')
		buf = append(buf, c.Buf...)
	}
	binary.BigEndian.PutUint16(buf, uint16(len(buf)-2))
	return buf
}

func (c *Chunk) Recv(conn *net.TCPConn) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = trace("base.Chunk.Recv(): %v", e)
		}
	}()
	assert(conn.SetReadDeadline(time.Now().Add(time.Minute)))
	defer func() {
		assert(conn.SetReadDeadline(time.Time{}))
	}()
	buf := make([]byte, 4096)
	_, err = io.ReadFull(conn, buf[:2])
	assert(err)
	clen := int(binary.BigEndian.Uint16(buf[:2]))
	data := make([]byte, clen)
	for {
		n, err := conn.Read(buf)
		assert(err)
		data = append(data, buf[:n]...)
		clen -= n
		if clen <= 0 {
			break
		}
	}
	cs := strings.SplitN(string(data), "\n", 3)
	if len(cs) != 3 {
		panic(errors.New("invalid chunk format"))
	}
	c.Src = parseAddr(cs[0])
	c.Dst = parseAddr(cs[1])
	c.Buf = []byte(cs[2])
	return
}

func (c Chunk) Send(conn *net.TCPConn) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = trace("base.Chunk.Send(): %v", e)
		}
	}()
	assert(conn.SetWriteDeadline(time.Now().Add(time.Minute)))
	defer func() {
		assert(conn.SetWriteDeadline(time.Time{}))
	}()
	_, err = conn.Write(c.Serialize())
	assert(err)
	return
}
