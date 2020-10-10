package authentication

import (
	"net/http"

	"github.com/TaKeO90/pwm/psql"
)

// Login interface implement the StartUserSession method of UserLogin.
type Login interface {
	StartUserSession() (bool, error)
}

// UserLogin structure has Element that we need to check if the user
// is registred and the password and username he provide are the identical
type UserLogin struct {
	username string
	password string
	w        http.ResponseWriter
}

// Register interface implement UserRegister.
type Register interface {
	CreateNewUser() (bool, error)
}

// UserRegister structure has element that we need to register a user.
type UserRegister struct {
	username string
	password string
	email    string
}

// NewLogin function returns the interface Login
func NewLogin(username, password string, w http.ResponseWriter) Login {
	var login Login
	ul := &UserLogin{username, password, w}
	login = ul
	return login
}

type Logout struct {
	Username string
	w        http.ResponseWriter
}

func NewLogout(username string, w http.ResponseWriter) Logout {
	l := &Logout{username, w}
	return *l
}

func (o Logout) StopUser() (bool, error) {
	pq, err := psql.NewDb()
	if err != nil {
		return false, err
	}
	userID, err := pq.GetUserID(o.Username)
	if err != nil {
		return false, err
	}
	ok, err := pq.LogoutUser(userID)
	if err != nil {
		return false, err
	}
	if ok {
		clearSession(o.w)
		return true, nil
	}
	return false, nil
}

// StartUserSession UserLogin's method check for username and password
// in the Database and if they exist we set the cookie token and also
// we store it into the Db.
func (l *UserLogin) StartUserSession() (bool, error) {
	if l.username != "" && l.password != "" {
		pq, err := psql.NewDb()
		if err != nil {
			return false, err
		}
		ok, err := pq.GetUsers(l.username, l.password)
		if err != nil {
			return false, err
		}
		if ok {
			s := NewSession()
			err = s.setSession(l.username, l.w)
			if err != nil {
				return false, err
			}
			//need to set & store the token.
			userid, err := pq.GetUserID(l.username)
			if err != nil {
				return false, err
			}
			isStored, err := pq.StoreSessionToken(userid, s.cookieToken)
			if err != nil {
				return false, err
			}
			if isStored {
				return true, nil
			}
		}
	}
	return false, nil
}

// NewRegister return Register interface.
func NewRegister(username, password, email string) Register {
	var register Register
	uR := &UserRegister{username, password, email}
	register = uR
	return register
}

// CreateNewUser register new User
func (r *UserRegister) CreateNewUser() (bool, error) {
	pq, err := psql.NewDb()
	if err != nil {
		return false, err
	}
	ok, err := pq.StoreUsers(r.username, r.password, r.email)
	if err != nil {
		return false, err
	}
	return ok, nil
}

////UpdatePassword update user password
//func UpdatePassword(email, password string) bool {
//	db, err := sql.Open("sqlite3", DB)
//	if err != nil {
//		panic(err)
//	}
//	defer db.Close()
//	isUpdated := sqlite.UpdatePw(password, email, db)
//	return isUpdated
//}
