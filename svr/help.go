package svr

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/pquerna/otp/totp"
)

const helptxt = `
How to use DoorKeepr

1. Request access to a site:
   - request SSH access for the gateway host of site "test":
     curl dksvr:<PADM>/<otp>/test/22
   - request RDP access for host "192.168.1.9" in site "test":
     curl dksvr:<PADM>/<otp>/test/3389/192.168.1.9
2. Connect to the site
   - SSH to the gateway just requested access:
     ssh -p<PWRK> dksvr
   - RDP to the host just requested access:
     rdesktop dksvr:<PWRK> 
3. Manage sites
   - list sites: 
     curl dksvr:<PADM>/<otp>
   - list hosts in site "test" with SSH port open:
     curl dksvr:<PADM>/<otp>/test/22/*
4. Manage connections
   - list active connections to site "test":
     curl dksvr:<PADM>/conns/<otp>/test
   - terminate all connections to site "test":
     curl dksvr:<PADM>/conns/<otp>/test?delete
5. Manage access tokens
   - list active access grants:
     curl dksvr:<PADM>/auths/<otp>
   - revoke all access grants:
     curl dksvr:<PADM>/auths/<otp>?delete
`

func helper(cf Config) func(http.ResponseWriter, *http.Request) {
	usage := strings.ReplaceAll(helptxt, "<PADM>", strconv.Itoa(cf.AdminPort))
	usage = strings.ReplaceAll(usage, "<PWRK>", strconv.Itoa(cf.ServePort))
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.Trim(r.URL.Path[6:], " /\t")
		if !totp.Validate(path, cf.OTP.Key) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=en_US.UTF-8")
		fmt.Fprintln(w, usage)
	}
}
