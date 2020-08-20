package tomluser

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
)

func Hash(mode string, password string, user *User) (string, error) {
	switch mode {
	case "md5":
		data := md5.Sum([]byte(password + user.Salt))
		return hex.EncodeToString(data[:]), nil

	case "sha256":
		data := sha256.Sum256([]byte(password + user.Salt))
		return hex.EncodeToString(data[:]), nil
	}
	return password, nil
}
