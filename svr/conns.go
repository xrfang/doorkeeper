package svr

import (
	"net/http"
	"strings"

	"github.com/pquerna/otp/totp"
)

func conns(cf Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.Trim(r.URL.Path, " /\t")
		s := strings.Split(path, "/")
		if len(s) == 0 || !totp.Validate(s[0], cf.OTP.Key) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	}
}
