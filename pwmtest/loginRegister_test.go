package pwmtest

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TaKeO90/pwm/server/handler"
)

func makeRequest(method, url string, data []byte) (*httptest.ResponseRecorder, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	rr := httptest.NewRecorder()
	h := http.HandlerFunc(handler.ReqHandler)
	h.ServeHTTP(rr, req)
	return rr, nil
}

type resp struct {
	resJSON    string
	statusCode int
}

func TestCreateUser(t *testing.T) {
	cases := map[string]resp{
		`{"username":"yassine","password":"test","email":"rand@gmail.com"}`: *(&resp{`{IsReg:false}`, 201}),
		`{"username":"mark","password":"test","email":"rand@gmail.com"}`:    *(&resp{`{IsReg:true}`, 409}),
	}
	for k, v := range cases {
		data := []byte(k)
		rr, err := makeRequest("POST", "/register", data)
		if err != nil {
			t.Log(err)
		}
		if rr.Code != v.statusCode && rr.Body.String() != v.resJSON {
			t.Log(rr.Code, rr.Body.String())
			t.Fail()
		}
	}
}

func TestStartUser(t *testing.T) {
	cases := map[string]resp{
		`{"username":"yassine","password":"test"}`:   *(&resp{`{IsLog:true}`, 200}),
		`{"username":"yassine","password":"secure"}`: *(&resp{`{IsLog:false}`, 400}),
	}

	for k, v := range cases {
		data := []byte(k)
		rr, err := makeRequest("POST", "/login", data)
		if err != nil {
			t.Log(err)
		}
		if rr.Code != v.statusCode && rr.Body.String() != v.resJSON {
			t.Log(rr.Code, rr.Body.String())
			t.Fail()
		}
	}
}
