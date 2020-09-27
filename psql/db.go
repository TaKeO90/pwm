package psql

import (
	"database/sql"
	"fmt"
	"os"
	"sync"

	"github.com/TaKeO90/pwm/services/pwhasher"
	"github.com/lib/pq"
)

//TODO: logout we need to remove the session_token from the db.

//TODO: Continue with Credentials logic.

//TODO: update user password.

var (
	dbUser     string = "root"
	dbPassword string = os.Getenv("psqlpw")
	testDB     string = "testdb"
	prodDB     string = "pwmdb"
)

// Psql structure holds db connection and mutex.
type Psql struct {
	Db  *sql.DB
	mtx *sync.Mutex
}

func createDBs(db *sql.DB) error {
	isTestdbExist, err := isExistDatabase(db, testDB)
	if err != nil {
		return err
	}
	isPwmdbExist, err := isExistDatabase(db, prodDB)
	if err != nil {
		return err
	}
	if !isTestdbExist {
		script := fmt.Sprintf("CREATE DATABASE %s", testDB)
		if _, err := db.Exec(script); err != nil {
			return err
		}
	}
	if !isPwmdbExist {
		script2 := fmt.Sprintf("CREATE DATABASE %s", prodDB)
		if _, err := db.Exec(script2); err != nil {
			return err
		}
	}
	return nil
}

func isExistDatabase(db *sql.DB, dbname string) (bool, error) {
	var dbName string
	if err := db.QueryRow("SELECT datname FROM pg_database WHERE datname = $1", dbname).Scan(&dbName); err != nil {
		return false, err
	}
	if dbName != "" {
		return true, nil
	}
	return false, nil
}

func IstablishAndCreateDB() error {
	connStr := fmt.Sprintf("postgresql://localhost/postgres?user=%s&password=%s&sslmode=disable", dbUser, dbPassword)
	db, err := sql.Open("postgres", connStr)
	defer db.Close()
	if err != nil {
		return err
	}
	if err := createDBs(db); err != nil {
		return err
	}
	return nil
}

// NewDb initialize the connection to the database and returns pointer to Psql or error if anything goes wrong.
func NewDb() (*Psql, error) {
	psql := new(Psql)
	mtx := new(sync.Mutex)
	var dbname string
	if isTest := os.Getenv("test"); isTest == "true" {
		dbname = testDB
	} else {
		dbname = prodDB
	}
	newconn := fmt.Sprintf("postgresql://localhost/postgres?user=%s&password=%s&sslmode=disable&dbname=%s", dbUser, dbPassword, dbname)
	db, err := sql.Open("postgres", newconn)
	if err != nil {
		return nil, err
	}
	psql.Db, psql.mtx = db, mtx
	return psql, nil
}

func (p *Psql) DropTestTables() error {
	script := `
		DROP TABLE IF EXISTS passwords;
		DROP TABLE IF EXISTS users;
	`
	_, err := p.Db.Exec(script)
	if err != nil {
		return err
	}
	return nil
}

// CreateTables Initialize the sql tables and create them if they are not exist any more.
func (p *Psql) CreateTables() error {
	tables := `
		CREATE TABLE IF NOT EXISTS users(
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) NOT NULL UNIQUE,
			password VARCHAR(100) NOT NULL,
			email VARCHAR(255) NOT NULL UNIQUE,
			session_token text []
		);

		CREATE TABLE IF NOT EXISTS passwords(
			pwid SERIAL PRIMARY KEY,
			u_name VARCHAR(255),
			passw VARCHAR(100),
			category VARCHAR(25),
			userid INTEGER,
			FOREIGN KEY(userid) REFERENCES users(id)
		);
	`
	_, err := p.Db.Exec(tables)
	if err != nil {
		return err
	}
	return nil
}

func checkUserDuplicated(user, email string, Db *sql.DB) (bool, error) {
	rows, err := Db.Query(`SELECT username, email FROM users`)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	var (
		username string
		emailA   string
	)
	for rows.Next() {
		err = rows.Scan(&username, &emailA)
		if err != nil {
			return false, err
		}
		if username == user || emailA == email {
			return true, nil
		}
	}
	return false, nil
}

func checkUserAndPw(user, pw string, db *sql.DB) (bool, error) {
	rows, err := db.Query(`SELECT username,password FROM users`)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			username string
			password string
		)
		err = rows.Scan(&username, &password)
		if err != nil {
			return false, err
		}
		if username == user && password == pw {
			return true, nil
		}
	}
	return false, nil
}

// StoreUsers the passw should be hashed by pwhasher package
func (p *Psql) StoreUsers(user, passw, email string) (bool, error) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	h := pwhasher.NewHash(passw)
	out, err := h.Phash()
	if err != nil {
		return false, err
	}
	isExist, err := checkUserDuplicated(user, email, p.Db)
	if err != nil {
		return false, err
	}
	if !isExist {
		query := `INSERT INTO users (username, password, email) VALUES ($1,$2,$3);`
		stmt, err := p.Db.Prepare(query)
		if err != nil {
			return false, err
		}
		_, err = stmt.Exec(user, out, email)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, fmt.Errorf("username or email already exist")
}

// GetUsers check if the user is the same and password for login.
func (p *Psql) GetUsers(user, password string) (bool, error) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	hshP, err := pwhasher.NewHash(password).Phash()
	if err != nil {
		return false, err
	}
	ok, err := checkUserAndPw(user, hshP, p.Db)
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}
	return false, nil
}

func getSessionToken(username string, db *sql.DB) (sessionToken []string, err error) {
	if err = db.QueryRow(`SELECT session_token FROM users WHERE username=$1`, username).Scan(pq.Array(&sessionToken)); err != nil {
		return sessionToken, err
	}
	return
}

func (p *Psql) StoreSessionToken(user, sessionToken string) (bool, error) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	sessionTokens, err := getSessionToken(user, p.Db)
	if err != nil {
		return false, err
	}
	sessionTokens = append(sessionTokens, sessionToken)
	var query string

	if len(sessionTokens) == 0 {
		query = `INSERT INTO users (session_token) VALUES($1) WHERE username=$2;`
		_, err = p.Db.Exec(query, pq.Array(sessionTokens), user)
		if err != nil {
			return false, err
		}
		return true, nil
	} else {
		query = `UPDATE users SET session_token=array_append(session_token,$1) WHERE username=$2;`
		_, err = p.Db.Exec(query, sessionToken, user)
		if err != nil {
			return false, err
		}
		return true, nil
	}
}

// UpdateUsers update user password.
func (p *Psql) UpdateUsers(username, password string) {
	//TODO: .... Check if the user exist first
	//			then get the update it password with the new one
	// 			after hashing it of course.
}
