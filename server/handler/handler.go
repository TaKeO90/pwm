package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/TaKeO90/pwm/server/dialer"
)

// ReqHandler handles http request made by the user
func ReqHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[%s] %s\n", r.Method, r.URL.Path)
	reqPath := strings.Trim(r.URL.Path, "/")
	switch reqPath {
	case "login":
		dialer.ServeLogin(w, r)
	case "register":
		dialer.ServeRegister(w, r)
	case "user":
		dialer.CookieDecode(w, r)
	case "show":
		dialer.ServeShow(w, r)
	case "forgot":
		//dialer.ServepwForget(w, r)
	case "creds":
		//dialer.ServeCreds(w, r)
	case "logout":
		dialer.GetLogout(w, r)
	case "upload":
		dialer.UploadCredFile(w, r)
	case "genpw":
		dialer.ServeGenPw(w, r)
	default:
		fmt.Fprintf(w, "404 Not Found", http.StatusNotFound)
	}
}
