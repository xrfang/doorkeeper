package svr

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/pquerna/otp/totp"
)

func conns(cf Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.Trim(r.URL.Path[6:], " /\t")
		s := strings.Split(path, "/")
		if len(s) != 2 || !totp.Validate(s[0], cf.OTP.Key) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		b := sm.getBackend(s[1])
		if b == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		fmt.Fprintln(w, "OK")
		conns := b.listConns()
		if len(conns) == 0 {
			fmt.Fprintf(w, "backend \"%s\" has no active connection\n", s[1])
			return
		}
		for _, c := range conns {
			h, _, _ := net.SplitHostPort(c)
			t := ra.Lookup(h)
			if t != nil {
				c = fmt.Sprintf("%s => %s", c, t.addr)
			}
			fmt.Fprintln(w, c)
		}
		_, delete := r.URL.Query()["delete"]
		if delete {
			b.flush()
			fmt.Fprintln(w, "all connections terminated")
		}
	}
}
