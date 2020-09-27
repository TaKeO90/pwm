package authentication

import (
	"fmt"
	"net/http"

	"github.com/gorilla/securecookie"
)

// sessionCookie cookie name
const sessionCookie = "session"

type session struct {
	cookieHandler *securecookie.SecureCookie
	cookieToken   string
}

func NewSession() *session {
	s := new(session)
	s.cookieHandler = securecookie.New(
		securecookie.GenerateRandomKey(64),
		securecookie.GenerateRandomKey(32))
	return s
}

// SetSession function takes username and the http response Writer as inputs
// Then Encode User's username and use it as a session cookie
// finally sets the cookie for the user otherwise it returns an error
func (s *session) setSession(username string, w http.ResponseWriter) error {
	value := map[string]string{
		"name": username,
	}
	var err error
	encoded, err := s.cookieHandler.Encode(sessionCookie, value)
	if err != nil {
		return err
	}
	s.cookieToken = encoded
	cookie := &http.Cookie{
		Name:   sessionCookie,
		Value:  encoded,
		Path:   "/",
		MaxAge: 3600,
	}
	//set cookie
	http.SetCookie(w, cookie)
	return nil
}

//GetCookieValue From *http.request get the Cookie if available
func GetCookieValue(r *http.Request) (string, error) {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// GetUsername function takes http request as input and returns the username who was encoded before in the cookie
func (s *session) GetUsername(r *http.Request) (username string) {
	if cookie, err := r.Cookie(sessionCookie); err == nil {
		cookieValue := make(map[string]string)
		if err = s.cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
			username = cookieValue["name"]
		} else {
			fmt.Println(err)
		}
	}
	return username
}

// ClearSession function Clears the cookie by  modifying the cookie's MaxAge
func ClearSession(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   sessionCookie,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
}

// CheckCookie function checks if there is already a cookie or not
func CheckCookie(r *http.Request) bool {
	var ok bool
	if cookie, _ := r.Cookie(sessionCookie); cookie != nil {
		ok = true
	} else {
		ok = false
	}
	return ok
}
