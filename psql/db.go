package psql

import (
	"database/sql"
	"fmt"
	"os"
	"sync"

	"github.com/TaKeO90/pwm/services/pwhasher"
	_ "github.com/lib/pq"
)

var (
	dbUser     string = "root"
	dbPassword string = os.Getenv("psqlpw")
)

// Psql structure holds db connection and mutex.
type Psql struct {
	Db  *sql.DB
	mtx *sync.Mutex
}

// NewDb initialize the connection to the database and returns pointer to Psql or error if anything goes wrong.
func NewDb() (*Psql, error) {
	psql := new(Psql)
	mtx := new(sync.Mutex)
	connStr := fmt.Sprintf("postgresql://localhost/postgres?user=%s&password=%s&sslmode=disable", dbUser, dbPassword)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	psql.Db, psql.mtx = db, mtx
	return psql, nil
}

// CreateTables Initialize the sql tables and create them if they are not exist any more.
func (p *Psql) CreateTables() error {
	tables := `
		CREATE TABLE IF NOT EXISTS users(
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) NOT NULL UNIQUE,
			password VARCHAR(100) NOT NULL,
			email VARCHAR(255) NOT NULL UNIQUE
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
