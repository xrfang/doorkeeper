package svr

import (
	"fmt"
	"sync"
)

var chm sync.Map

func addChan(ip string, port int) chan []byte {
	key := fmt.Sprintf("%s:%d", ip, port)
	val, ok := chm.Load(key)
	if ok {
		close(val.(chan []byte))
		chm.Delete(key)
	}
	ch := make(chan []byte)
	chm.Store(key, ch)
	return ch
}

func delChan(ip string, port int) {
	key := fmt.Sprintf("%s:%d", ip, port)
	chm.Delete(key)
}

func getChan(ip string, port int) chan []byte {
	key := fmt.Sprintf("%s:%d", ip, port)
	val, ok := chm.Load(key)
	if !ok {
		return nil
	}
	return val.(chan []byte)
}
