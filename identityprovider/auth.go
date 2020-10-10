package identityprovider

func empty() {
}

//
//import (
//	"net/http"
//
//	"github.com/TaKeO90/pwm/authentication"
//	//"github.com/TaKeO90/pwm/services/pwencrypter"
//	//"github.com/TaKeO90/pwm/services/serverenc"
//)

//// GetLoggedin Login the User & give him an identity (Cookie)
//func GetLoggedin(w http.ResponseWriter, r *http.Request, user string, password string) (bool, error) {
//	var rr bool
//	if ok := authentication.CheckCookie(r); ok {
//		rr = false
//	} else {
//		if logged := authentication.Login(user, password); logged {
//			err := authentication.SetSession(user, w)
//			if err != nil {
//				return false, err
//			}
//			rr = true
//		} else {
//			rr = false
//		}
//	}
//	return rr, nil
//}
//
//// GetRegister Register the User (Function takes http request ,user , password and email as input and returns a bool)
//func GetRegister(r *http.Request, user string, password string, email string) bool {
//	var rr bool
//	//	if ok := authentication.CheckCookie(r); ok {
//	//		rr = false
//	//	} else {
//	if isRegistred := authentication.Register(user, password, email); !isRegistred {
//		rr = false
//	} else {
//		// Generating the encryption key for the user then save it in the keys directory
//		key := pwencrypter.GenKeyP(password)
//		// Load server key and Encrypt user key
//		srvk := pwencrypter.LoadKey("server")
//		nuserK := serverenc.EncryptUserKey(key, srvk)
//		if isSaved := pwencrypter.SaveKey(nuserK, user); isSaved {
//			rr = true
//		}
//	}
//	return rr
//}

// GetLoggedout Logout the User and clear session cookie
// returns true if the process went good otherwise it returns false
//func GetLoggedout(w http.ResponseWriter, r *http.Request) bool {
//	var rr bool
//	if ok := authentication.CheckCookie(r); ok {
//		authentication.ClearSession(w)
//		rr = true
//	} else {
//		rr = false
//	}
//	return rr
//}
