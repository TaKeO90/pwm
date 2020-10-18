package gentoken

import (
	"math/rand"
	"time"
)

type VerificationCode struct {
	Length, Min, Max int
}

func (v *VerificationCode) GetToken() string {
	rand.Seed(time.Now().UTC().UnixNano())
	code := make([]byte, v.Length)
	for i := 0; i < v.Length; i++ {
		n := v.Min + rand.Intn(v.Max-v.Min)
		code[i] = byte(n)
	}
	return string(code)
}
