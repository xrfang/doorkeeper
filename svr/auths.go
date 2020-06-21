package svr

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/pquerna/otp/totp"
)

func auths(cf Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.Trim(r.URL.Path, " /\t")
		s := strings.Split(path, "/")
		if len(s) == 0 || !totp.Validate(s[0], cf.OTP.Key) {
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
	}
}
