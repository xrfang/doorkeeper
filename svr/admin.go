package svr

import (
	"dk/base"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/pquerna/otp/totp"
)

func notFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not found", http.StatusNotFound)
}

var allowed func(*http.Request, string) bool

func startAdminIntf(cf Config) {
	pid := strconv.Itoa(os.Getpid())
	allowed = func(r *http.Request, otp string) bool {
		if totp.Validate(otp, cf.OTP.Key) {
			return true
		}
		if !cf.M2MIntf {
			return false
		}
		host, _, _ := net.SplitHostPort(r.RemoteAddr)
		ip := net.ParseIP(host)
		if ip == nil || !ip.IsLoopback() {
			return false
		}
		return pid == otp
	}
	http.HandleFunc("/", controller(cf))
	http.HandleFunc("/help", notFound)
	http.HandleFunc("/help/", helper(cf))
	http.HandleFunc("/auths", notFound)
	http.HandleFunc("/auths/", auths(cf))
	http.HandleFunc("/conns", notFound)
	http.HandleFunc("/conns/", conns(cf))
	adm := http.Server{
		Addr:         fmt.Sprintf(":%d", cf.AdminPort),
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	}
	go func() {
		err := adm.ListenAndServe()
		base.Log("startAdminIntf: %v", err)
		time.Sleep(2 * time.Second)
		os.Exit(1)
	}()
}
