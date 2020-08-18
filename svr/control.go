package svr

import (
	"dk/base"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

/* 访问/help/<otp>获取帮助 */
func controller(cf Config) func(http.ResponseWriter, *http.Request) {
	fwr := regexp.MustCompile(`for=(?:["[])*([0-9a-fA-F:.]+)`)
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.Trim(r.URL.Path, " /\t")
		s := strings.Split(path, "/")
		if len(s) == 0 || !allowed(r, s[0]) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		switch len(s) {
		case 1:
			fmt.Fprintln(w, "OK")
			for n := range cf.Auth {
				stat := "offline"
				if sm.getBackend(n) != nil {
					stat = "online"
				}
				fmt.Fprintf(w, "%s: %s\n", n, stat)
			}
			return
		case 3:
			s = append(s, "127.0.0.1")
		case 4:
			if s[3] == "" {
				s[3] = "127.0.0.1"
			}
		default:
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if cf.Auth[s[1]] == "" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		port, _ := strconv.Atoi(s[2])
		if port <= 0 || port > 65535 {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		srch, srcp, _ := net.SplitHostPort(r.RemoteAddr)
		src := srch
		ff := fwr.FindStringSubmatch(r.Header.Get("Forwarded"))
		if len(ff) > 1 {
			src = ff[1]
		}
		srcIP := net.ParseIP(src)
		if srcIP == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if s[3] == "*" {
			b := sm.getBackend(s[1])
			if b == nil {
				fmt.Fprintln(w, "ERR")
				fmt.Fprintf(w, "backend \"%s\" not found\n", s[1])
				return
			}
			ip := net.ParseIP(srch)
			sp, _ := strconv.Atoi(srcp)
			qry := base.Chunk{
				Type: base.CT_QRY,
				Src:  &net.TCPAddr{IP: ip, Port: sp},
				Dst:  &net.TCPAddr{IP: srcIP, Port: port},
			}
			err := qry.Send(b.serv)
			if err != nil {
				fmt.Fprintln(w, "ERR")
				fmt.Fprintln(w, err)
				return
			}
			select {
			case msg := <-addChan(srch, sp):
				fmt.Fprint(w, string(msg))
			case <-time.After(time.Minute):
				http.Error(w, "timeout", http.StatusRequestTimeout)
			}
			delChan(srch, sp)
			return
		}
		ip := net.ParseIP(s[3])
		if ip == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		now := time.Now()
		maxIdle := now.Add(time.Duration(cf.IdleClose) * time.Second)
		maxLife := now.Add(time.Duration(cf.AuthTime) * time.Second)
		addr := net.TCPAddr{IP: ip, Port: port}
		ra.Register(src, s[1], &addr)
		fmt.Fprintln(w, "OK")
		fmt.Fprintln(w, "start_before:", maxIdle.Format(time.RFC3339))
		fmt.Fprintln(w, "expire_after:", maxLife.Format(time.RFC3339))
	}
}
