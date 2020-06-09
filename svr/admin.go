package svr

import (
	"fmt"
	"net/http"
	"time"
)

func startAdminIntf(cf Config) {
	adm := http.Server{
		Addr:         fmt.Sprintf(":%d", cf.AdminPort),
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	}
	go func() {
		assert(adm.ListenAndServe())
	}()
}
