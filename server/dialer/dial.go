package dialer

//TODO: add a go routine to send emails

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	//"math/rand"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	//"time"

	"github.com/TaKeO90/pwm/authentication"
	"github.com/TaKeO90/pwm/identityprovider"
	//"github.com/TaKeO90/pwm/services/emailsender"
	"github.com/TaKeO90/pwm/services/genpassw"
	//"github.com/TaKeO90/pwm/services/pwcompare"
	"github.com/TaKeO90/pwm/services/pwsaver"
	"github.com/TaKeO90/pwm/services/pwshow"
	"github.com/TaKeO90/pwm/services/readcredfile"
)

// User Struct
type User struct {
	IsEmpty  bool            `json:"isEmpty"`
	CredList pwshow.UserList `json:"CredList"`
}

type InternalServerError struct {
	Error string `json:"error"`
}

// CookieUser struct holds username of the current user in the session
type CookieUser struct {
	Username string `json:"Username"`
}

//UserData User's data or Credential
type UserData struct {
	Category   string          `json:"Category"`
	Credential pwshow.UserList `json:"Credential"`
}

//FileUploaded struct holds informations about how success is the process of uploading credential file
type FileUploaded struct {
	Lines   []int `json:"Lines"`
	Success bool  `json:"Success"`
}

//Register IsReg value is sent to inform that the user is successfully registred
type Register struct {
	IsReg bool `json:"IsReg"`
}

//Creds send Updated if the db is updated
type Creds struct {
	Updated bool `json:"Updated"`
}

//Login login struct holds values that we need in the frontend to check if we have successfully logged in
type Login struct {
	IsLog bool `json:"IsLog"`
}

// LoginReq
type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisReq
type RegisReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

//Logout holds a bool value to indicate which the user is logged out or not
type Logout struct {
	IsLogout bool `json:"IsLogout"`
}

//Token is the code that we send to the user to restore his password
type Token struct {
	Code string
}

//Password password struct holds a value which is Updated to inform the frontend that the user password is updated
type Password struct {
	Updated bool `json:"Updated"`
}

//Email has Response and IsEqual instances we need them in Password recovery process.
type Email struct {
	Response bool `json:"Response"`
	IsEqual  bool `json:"IsEqual"`
}

type GenPassword struct {
	Value string
}

// HandlePost function handles Post http method
// Gets input from form and returns it as map[string][]string
func HandlePost(r *http.Request) map[string][]string {
	var m map[string][]string
	if r.Method == "POST" {
		r.ParseMultipartForm(0)
		//r.ParseForm()
		m = r.Form
	}
	return m
}

func handleFilePost(r *http.Request) (multipart.File, *multipart.FileHeader, error) {
	if r.Method == "POST" {
		r.ParseMultipartForm(30)
		file, handler, err := r.FormFile("myfile")
		if err != nil {
			return nil, nil, err
		}
		return file, handler, nil
	}
	return nil, nil, nil
}

//UploadCredFile handle uploading csv file here
func UploadCredFile(w http.ResponseWriter, r *http.Request) {
	handleOption(w, r)
	u := &FileUploaded{}
	user := authentication.NewSession().GetUsername(r)
	file, handler, err := handleFilePost(r)
	defer file.Close()
	CheckError(err)
	if user != "" && handler != nil && filepath.Ext(handler.Filename) == ".csv" {
		filebyte, err := ioutil.ReadAll(file)
		CheckError(err)
		n := readcredfile.New(filebyte)
		var dump readcredfile.DumpCreds = n
		dumpedData := dump.ExtractCreds()
		if len(dumpedData) != 0 {
			line, ok := dumpedData.Compare(user)
			if ok {
				u.Lines = line
				u.Success = false
				json.NewEncoder(w).Encode(u)
			} else if ok == false && len(line) == 0 {
				var saved bool
				for _, i := range dumpedData {
					saved = pwsaver.AddCreds(i.Username, i.Password, i.Category, user)
				}
				if saved {
					u.Lines = line
					u.Success = true
					json.NewEncoder(w).Encode(u)
				}
			} else {
				u.Success = false
				json.NewEncoder(w).Encode(u)
			}
		} else {
			u.Success = false
			json.NewEncoder(w).Encode(u)
		}
	}
}

func setupResponse(w *http.ResponseWriter, r *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "https://localhost:4430")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST,GET")
	(*w).Header().Set("Access-Control-Allow-Credentials", "true")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length,Set-Cookie , Accept-Encoding, X-CSRF-Token, Authorization")
	(*w).Header().Set("Content-Type", "application/json")
}

func handleOption(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	if r.Method == "OPTIONS" {
		return
	}
}

// CheckError function check if err not nil then log the error
func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// ServeRegister Serve Registration
func ServeRegister(w http.ResponseWriter, r *http.Request) {
	handleOption(w, r)
	if r.Method == "POST" {
		rData := new(RegisReq)
		err := json.NewDecoder(r.Body).Decode(rData)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(&InternalServerError{fmt.Sprintf("Cannot parse JSON data %s", err.Error())})
		}
		register := authentication.NewRegister(rData.Username, rData.Password, rData.Email)
		ok, err := register.CreateNewUser()
		if err != nil && err.Error() != "username or email already exist" {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(&InternalServerError{fmt.Sprintf("Something Went Wrong %s", err.Error())})
		} else if err != nil && err.Error() == "username or email already exist" {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(&Register{false})
		}
		if ok {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(&Register{true})
		} else {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(&Register{false})
		}
	}
}

// ServeLogin handle login process
func ServeLogin(w http.ResponseWriter, r *http.Request) {
	handleOption(w, r)
	if r.Body != nil {
		loginElemnt := new(LoginReq)
		err := json.NewDecoder(r.Body).Decode(loginElemnt)
		if err != nil {
			log.Fatal(err)
		}
		login := authentication.NewLogin(loginElemnt.Username, loginElemnt.Password, w)
		ok, err := login.StartUserSession()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(&InternalServerError{fmt.Sprintf("Something Went Wrong %s", err.Error())})
		}
		if ok {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(&Login{true})
		} else {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(&Login{false})
		}
	}
}

//CookieDecode decode the cookie and get the username
func CookieDecode(w http.ResponseWriter, r *http.Request) {
	handleOption(w, r)
	uname := authentication.NewSession().GetUsername(r)
	c := &CookieUser{uname}
	json.NewEncoder(w).Encode(c)
}

//GetLogout logout the user by clearing the cookie value
func GetLogout(w http.ResponseWriter, r *http.Request) {
	handleOption(w, r)
	logout := &Logout{}
	if ok := identityprovider.GetLoggedout(w, r); ok {
		logout.IsLogout = ok
		json.NewEncoder(w).Encode(logout)
	} else {
		logout.IsLogout = ok
		json.NewEncoder(w).Encode(logout)
	}
}

// ServeShow Show User credentials
func ServeShow(w http.ResponseWriter, r *http.Request) {
	handleOption(w, r)
	var l pwshow.UserList
	user := authentication.NewSession().GetUsername(r)
	if f := HandlePost(r); len(f) != 0 && user != "" {
		category := strings.Join(f["category"], "")
		l = pwshow.ShowCreds(user, category)
		u := &User{}
		if len(l) != 0 {
			u.CredList = l
			json.NewEncoder(w).Encode(u)
		} else {
			u.IsEmpty = true
			json.NewEncoder(w).Encode(u)
		}
	}
}

var t Token

//ServepwForget function that handles password recovery process
//func ServepwForget(w http.ResponseWriter, r *http.Request) {
//	handleOption(w, r)
//	e := &Email{}
//	if f := HandlePost(r); len(f) != 0 {
//		password := strings.Join(f["npassword"], "")
//		if email := strings.Join(f["email"], ""); email != "" && password == "" {
//			if mailExist := authentication.CheckMail(email); mailExist {
//				rand.Seed(time.Now().UnixNano())
//				gencode := rand.Perm(6)
//				var c string
//				// Convert []int into string
//				for _, n := range gencode {
//					c += strconv.Itoa(n)
//				}
//				if sent, err := emailsender.SendCode(c, email); err != nil {
//					log.Fatal(err)
//				} else if sent {
//					t.Code = c
//					e.Response = true
//					json.NewEncoder(w).Encode(e)
//				}
//			} else {
//				t.Code = ""
//				e.Response = false
//				json.NewEncoder(w).Encode(e)
//			}
//		} else if code := strings.Join(f["code"], ""); code != "" {
//			if code == t.Code {
//				e.IsEqual = true
//				json.NewEncoder(w).Encode(e)
//			} else {
//				e.IsEqual = false
//				json.NewEncoder(w).Encode(e)
//			}
//		} else if password != "" && email != "" {
//			p := &Password{}
//			isUpdated := authentication.UpdatePassword(email, password)
//			if isUpdated {
//				p.Updated = isUpdated
//				json.NewEncoder(w).Encode(p)
//			} else {
//				p.Updated = false
//				json.NewEncoder(w).Encode(p)
//			}
//		}
//	}
//}

var d UserData //pwshow.UserList

func handleJSONBody(r *http.Request) error {
	if r.Body != nil {
		err := json.NewDecoder(r.Body).Decode(&d)
		return err
	}
	return nil
}

//TODO: not sure 100% need test
//ServeCreds update and delete User creds if the front-end credential has been changed
//func ServeCreds(w http.ResponseWriter, r *http.Request) {
//	handleOption(w, r)
//	username := authentication.GetUsername(r)
//	err := handleJSONBody(r)
//	CheckError(err)
//	if username != "" && d.Credential != nil {
//		index := pwcompare.Compare(pwshow.FinalList, d.Credential)
//		if ok := pwcompare.DiffAnalyse(username, d.Category, pwshow.FinalList, d.Credential, index); ok {
//			c := &Creds{Updated: true}
//			json.NewEncoder(w).Encode(c)
//		} else {
//			c := &Creds{Updated: false}
//			json.NewEncoder(w).Encode(c)
//		}
//
//	} else {
//		c := &Creds{Updated: false}
//		json.NewEncoder(w).Encode(c)
//	}
//}

func ServeGenPw(w http.ResponseWriter, r *http.Request) {
	handleOption(w, r)
	g := &GenPassword{}
	if postF := HandlePost(r); len(postF) != 0 {
		var genP genpassw.PwType
		pwType := strings.Join(postF["type"], "")
		length := strings.Join(postF["length"], "")
		if pwType != "" && length != "" {
			l, err := strconv.Atoi(length)
			if err != nil {
				log.Fatal(err)
			}
			pw := genpassw.New(l)
			genP = pw
			switch pwType {
			case "character":
				g.Value = strings.TrimSpace(genP.GenerateChars())
			case "integer":
				g.Value = strings.TrimSpace(genP.GenerateInts())
			case "special":
				g.Value = strings.TrimSpace(genP.GenerateSpchars())
			case "mix":
				g.Value = strings.TrimSpace(genP.GenerateComplex())
			}
			json.NewEncoder(w).Encode(g)
		} else {
			json.NewEncoder(w).Encode(g)
		}
	}
}
