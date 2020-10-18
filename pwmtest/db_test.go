package pwmtest

import (
	"testing"

	"github.com/TaKeO90/pwm/psql"
)

func TestCreateTables(t *testing.T) {
	if err := psql.IstablishAndCreateDB(); err != nil {
		t.Log("Cannot istablish cnx with postgres", err.Error())
		t.Fail()
		t.Log(err)
	}
	if err := setupEnv(); err != nil {
		t.Fail()
		t.Log(err)
	}
	db, err := psql.NewDb()
	if err != nil {
		t.Log("Failed to initialize db Connection")
		t.Fail()
		t.Log(err)
	}
	t.Log("Dropping Tables")
	if err := db.DropTestTables(); err != nil {
		t.Log("Cannot Drop tables")
		t.Fail()
		t.Log(err)
	}
	t.Log("Creating Tables")
	err = db.CreateTables()
	if err != nil {
		t.Log("Cannot Create Tables")
		t.Fail()
		t.Log(err)
	}
}
