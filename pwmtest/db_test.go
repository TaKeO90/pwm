package pwmtest

import (
	"os"
	"testing"

	"github.com/TaKeO90/pwm/psql"
)

func TestCreateTables(t *testing.T) {
	if err := os.Setenv("test", "true"); err != nil {
		t.Fail()
		t.Log(err)
	}
	db, err := psql.NewDb()
	if err != nil {
		t.Fail()
		t.Log(err)
	}
	if err := db.DropTestTables(); err != nil {
		t.Fail()
		t.Log(err)
	}
	err = db.CreateTables()
	if err != nil {
		t.Fail()
		t.Log(err)
	}
}
