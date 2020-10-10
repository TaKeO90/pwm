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
		`{"username":"yassine","password":"test","email":"rand@gmail.com"}`: resp{
			resJSON:    `{IsReg:false}`,
			statusCode: 201,
		},
		`{"username":"mark","password":"test","email":"rand@gmail.com"}`: resp{
			resJSON:    `{IsReg:true}`,
			statusCode: 409,
		},
	}
	for k, v := range cases {
		data := []byte(k)
		rr, err := makeRequest("POST", "/register", data)
		if err != nil {
			t.Log(err)
		}
		if rr.Code != v.statusCode && rr.Body.String() != v.resJSON {
			t.Log(v.statusCode, v.resJSON)

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

func TestLoginNewSession(t *testing.T) {
	specialCase := struct {
		data     string
		response resp
	}{
		`{"username":"yassine","password":"test"}`, resp{
			resJSON:    `{IsLog:false}`,
			statusCode: 409,
		},
	}

	rr, err := makeRequest("POST", "/login", []byte(specialCase.data))
	if err != nil {
		t.Log("Cannot make request")
		t.Log(err)
		t.Fail()
	}
	if rr.Code != specialCase.response.statusCode && rr.Body.String() != specialCase.response.resJSON {
		t.Log(specialCase.response.statusCode, specialCase.response.resJSON)
		t.Fail()
	}

}

type logoutResponse struct {
	statusCode int
	jsonData   string
}

type logoutStuff struct {
	requestData string
	resp        logoutResponse
}

func TestLogout(t *testing.T) {
	data := logoutStuff{
		requestData: `{"username":"yassine"}`,
		resp: logoutResponse{
			statusCode: 200,
			jsonData:   `IsLogout:true`,
		},
	}
	rr, err := makeRequest("POST", "/logout", []byte(data.requestData))
	if err != nil {
		t.Fail()
		t.Log(err)
	}
	if rr.Code != data.resp.statusCode && rr.Body.String() != data.resp.jsonData {
		t.Log(rr.Code, rr.Body.String())
		t.Fail()
	}
}
