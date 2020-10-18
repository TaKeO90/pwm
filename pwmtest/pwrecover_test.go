package pwmtest

import (
	"fmt"
	"testing"
)

var currentToken string

func TestEmailValidation(t *testing.T) {
	cases := map[string]int{
		`{"email":"rand@gmail.com"}`: 200,
		`{"email":"y@gmail.com"}`:    401,
		`{"email":}`:                 500,
	}

	for k, v := range cases {
		rr, err := makeRequest("POST", "/forgot", []byte(k))
		if err != nil {
			t.Fail()
			t.Log(err)
		}
		if rr.Code != v {
			t.Log(rr.Code)
			t.Log(rr.Body.String())
			t.Fail()
		}
		if v == 200 {
			currentToken = rr.Header().Get("Vertification-Token")
			t.Log("Verification Code: ", currentToken)
		}
		if v == 500 {
			t.Log("INTERNAL SERVER ERROR: ", rr.Body.String())
		}
	}
}

func TestPasswordRecovery(t *testing.T) {
	cases := map[string]int{
		fmt.Sprintf(`{"token":"%s","newpassword":"securepw"}:`, currentToken): 200,
	}
	fmt.Println(currentToken)
	for k, v := range cases {
		rr, err := makeRequest("POST", "/recover", []byte(k))
		if err != nil {
			t.Fail()
			t.Log(err)
		}
		if rr.Code != v {
			t.Fail()
			t.Log(rr.Code, rr.Body.String())
		}
		if rr.Code == 200 {
			t.Log(rr.Body.String())
		}
	}
}
