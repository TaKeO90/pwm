package pwmtest

import (
	"os"
)

func setupEnv() error {
	if err := os.Setenv("test", "true"); err != nil {
		return err
	}
	return nil
}
