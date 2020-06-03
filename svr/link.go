package svr

import (
	"net"
	"sync"
)

type link struct {
	stat int //连接状态：-1=已终止；0=待认证/主控连接；1=活跃
	conn net.Conn
}

func (l *link) setup(conn net.Conn, stat int) {
	l.stat = stat
	l.conn = conn
	go func() {
		//TODO:
	}()
}

type peer struct {
	role  int //对端状态：0=未知；1=DK客户端（后端服务）；2=远端访问者
	links []link
}

func (p *peer) validate() bool {
	conn := p.links[len(p.links)-1]
	conn.SetReadDeadline()
	return false //TODO
}

func (p *peer) relay() {

}

type connMgr struct {
	peers map[string]*peer
	sync.Mutex
}

func (cm *connMgr) Register(conn net.Conn) {
	cm.Lock()
	defer cm.Unlock()
	host, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
	p := cm.peers[host]
	if p == nil {
		p = &peer{role: 0, links: []link{{stat: 0, conn: conn}}}
		if p.validate() {
			cm.peers[host] = p
		}
		return
	}
	switch p.role {
	case 0: //一个等待确认身份的远端又一次发起连接
		conn.Close() //关闭该连接，不予理会
	case 1: //一个后端服务再次发起新的主控连接
		p.links = append(p.links, link{stat: 0, conn: conn})
		if p.validate() { //新建主控连接仍然需要握手认证
			cm.peers[host] = p
		}
	case 2: //一个远端客户发起连接
		p.links = append(p.links, link{stat: 1, conn: conn})
		p.relay() //为该连接启动转发服务
		cm.peers[host] = p
	}
}

var conns connMgr

func init() {
	conns.Lock()
	defer conns.Unlock()
	conns.peers = make(map[string]*peer)
}
