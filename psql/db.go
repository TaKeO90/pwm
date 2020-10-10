package psql

import (
	"database/sql"
	"fmt"
	"os"
	"sync"

	"github.com/TaKeO90/pwm/services/pwhasher"
	_ "github.com/lib/pq"
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

func createDBs(db *sql.DB) error {
	isTestdbExist, err := isExistDatabase(db, testDB)
	if err != nil {
		return fmt.Errorf("Cannot check if testdb exists or not")
	}
	isPwmdbExist, err := isExistDatabase(db, prodDB)
	if err != nil {
		return fmt.Errorf("Cannot check if prodDb exists or not")
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
	var dbNumber int
	if err := db.QueryRow("SELECT COUNT(datname) FROM pg_database WHERE datname = $1",
		dbname).Scan(&dbNumber); err != nil {
		return false, err
	}
	if dbNumber != 0 {
		return true, nil
	}
	return false, nil
}

// IstablishAndCreateDB istablish a connection to the posgtres database and create the databases.(test database & production database)
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

func (p *Psql) DropTestTables() error {
	script := `
		DROP TABLE IF EXISTS passwords;
		DROP TABLE IF EXISTS users;
		DROP TABLE IF EXISTS sessions;
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

		CREATE TABLE IF NOT EXISTS sessions(
			user_id INTEGER PRIMARY KEY NOT NULL,
			session_token VARCHAR
		);
	`
	_, err := p.Db.Exec(tables)
	if err != nil {
		return err
	}
	return nil
}

func checkUserDuplicated(user, email string, Db *sql.DB) (userExist bool, emailExist bool, err error) {
	rows, err := Db.Query(`SELECT username, email FROM users`)
	if err != nil {
		return false, false, err
	}
	defer rows.Close()
	var (
		username string
		emailA   string
	)
	for rows.Next() {
		err = rows.Scan(&username, &emailA)
		if err != nil {
			return false, false, err
		}
		if username == user {
			userExist = true
		}
		if emailA == email {
			emailExist = true
		}
	}
	return
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
	isUserExist, isEmailExist, err := checkUserDuplicated(user, email, p.Db)
	if err != nil {
		return false, err
	}
	if !isUserExist && !isEmailExist {
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

// StoreSessionToken Checks if we already have user sessions then revoke them if they exist and create new one.
func (p *Psql) StoreSessionToken(userID int, sessionToken string) (created bool, err error) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	// Check if we already have a session token in sessions table
	// maybe we need to allow for one session for the user at the moment.
	sessionTokenNumber, err := getSessionToken(userID, p.Db)
	if err != nil {
		return false, err
	}
	if sessionTokenNumber == 1 {
		return false, fmt.Errorf("Session Already Exist")
	} else {
		isCreated, err := createNewSession(userID, sessionToken, p.Db)
		if err != nil {
			return false, err
		}
		created = isCreated
	}
	return
}

//
// Aux Function for StoreSessionToken
func getSessionToken(userID int, db *sql.DB) (sessionTokensNumber int, err error) {
	if err = db.QueryRow(`SELECT COUNT(session_token) FROM sessions WHERE user_id=$1`,
		userID).Scan(&sessionTokensNumber); err != nil {
		return sessionTokensNumber, fmt.Errorf("Cannot get the number of user tokens")
	}
	fmt.Println(sessionTokensNumber)
	return
}

func revokeSession(userID int, db *sql.DB) (bool, error) {
	tag, err := db.Exec(`DELETE FROM sessions WHERE user_id=$1`, userID)
	if err != nil {
		return false, fmt.Errorf("Cannot revoke session token")
	}
	rowAff, err := tag.RowsAffected()
	if err != nil {
		return false, err
	}
	if rowAff != 1 {
		return false, fmt.Errorf("Cannot revoke sessions token")
	}
	return true, nil
}

func createNewSession(userID int, sessionToken string, db *sql.DB) (bool, error) {
	rows, err := db.Exec(`INSERT INTO sessions (user_id,session_token) 
		VALUES ($1,$2)`, userID, sessionToken)
	if err != nil {
		return false, fmt.Errorf("Cannot Insert New session token")
	}
	rowAffected, err := rows.RowsAffected()
	if err != nil {
		return false, err
	}
	if rowAffected != 1 {
		return false, fmt.Errorf("Cannot Create New Session for the user")
	}
	return true, nil
}

//
//

// GetUserID get user id by providing username.
func (p *Psql) GetUserID(username string) (userID int, err error) {
	err = p.Db.QueryRow(`select id from users where username=$1`, username).Scan(&userID)
	return
}

func (p *Psql) LogoutUser(userID int) (ok bool, err error) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	ok, err = revokeSession(userID, p.Db)
	if err != nil {
		return false, err
	}
	return
}

// UpdateUsers update user password.
func (p *Psql) UpdateUsers(username, password string) {
	//TODO: .... Check if the user exist first
	//			then get the update it password with the new one
	// 			after hashing it of course.
}

//	sessionTokens, err := getSessionToken(user, p.Db)
//	if err != nil {
//		return false, err
//	}
//	sessionTokens = append(sessionTokens, sessionToken)
//	var query string
//
//	if len(sessionTokens) == 0 {
//		query = `INSERT INTO users (session_token) VALUES($1) WHERE username=$2;`
//		_, err = p.Db.Exec(query, pq.Array(sessionTokens), user)
//		if err != nil {
//			return false, err
//		}
//		return true, nil
//	} else {
//		query = `UPDATE users SET session_token=array_append(session_token,$1) WHERE username=$2;`
//		_, err = p.Db.Exec(query, sessionToken, user)
//		if err != nil {
//			return false, err
//		}
//		return true, nil
//	}
