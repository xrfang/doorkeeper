package svr

import (
	"fmt"
	"net/http"
	"strings"
)

func auths(cf Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.Trim(r.URL.Path[6:], " /\t")
		if !allowed(r, path) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		fmt.Fprintln(w, "OK")
		auths := ra.getSummary()
		if len(auths) == 0 {
			fmt.Fprintln(w, "no active authorization")
			return
		}
		for _, a := range auths {
			fmt.Fprintln(w, a)
		}
		_, delete := r.URL.Query()["delete"]
		if delete {
			ra.flush()
			fmt.Fprintln(w, "all authorization revoked")
		}
	}
}
