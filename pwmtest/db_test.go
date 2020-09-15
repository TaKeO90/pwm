package pwmtest

import (
	"testing"

	"github.com/TaKeO90/pwm/psql"
)

func TestCreateTables(t *testing.T) {
	nDb, err := psql.NewDb()
	if err != nil {
		t.Fail()
		t.Log(err)
	}
	err = nDb.CreateTables()
	if err != nil {
		t.Fail()
		t.Log(err)
	}
}

type registerFields struct {
	username string
	password string
	email    string
}

type loginFields struct {
	username string
	password string
}

func TestStoreUsers(t *testing.T) {
	cases := map[registerFields]bool{
		*(&registerFields{"yassine", "securepasswordbtw", "m@gmail.com"}): true,
		*(&registerFields{"yassine", "otherpw", "o@gmail.com"}):           false,
	}
	db, err := psql.NewDb()
	if err != nil {
		t.Fail()
		t.Log(err)
	}

	for k, v := range cases {
		ok, err := db.StoreUsers(k.username, k.password, k.email)
		if err != nil && err.Error() != "username or email already exist" {
			t.Fail()
			t.Log(err)
		}
		if ok != v {
			t.Fail()
		}
	}
}

func TestLoginUser(t *testing.T) {
	cases := map[loginFields]bool{
		*(&loginFields{"yassine", "securepasswordbtw"}): true,
		*(&loginFields{"yassine", "wrong"}):             false,
		*(&loginFields{"bob", "securebob"}):             false,
	}
	db, err := psql.NewDb()
	if err != nil {
		t.Fail()
		t.Log(err)
	}
	for k, v := range cases {
		ok, err := db.GetUsers(k.username, k.password)
		if err != nil {
			t.Fail()
			t.Log(err)
		}
		if v != ok {
			t.Log(v, ok)
			t.Fail()
		}

	}
}
