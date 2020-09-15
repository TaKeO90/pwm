package pwhasher

import (
	"crypto/sha256"
	"encoding/hex"
)

// PwHash interface that implement hashedPw
type PwHash interface {
	Phash() (string, error)
}

type hashedPw struct {
	passw string
}

// NewHash Initialize new hashedPw struct with it instances.
func NewHash(pw string) PwHash {
	var pwH PwHash
	h := &hashedPw{pw}
	pwH = h
	return pwH
}

func hashP(pw string) (string, error) {
	hs := sha256.New()
	_, err := hs.Write([]byte(pw))
	if err != nil {
		return "", err
	}
	hx := hs.Sum(nil)
	out := hex.EncodeToString(hx)
	return out, nil
}

// Phash hashedPw method that hashes the password given and return a string.
// or error if something went wrong.
func (h *hashedPw) Phash() (string, error) {
	out, err := hashP(h.passw)
	if err != nil {
		return "", err
	}
	return out, nil
}
