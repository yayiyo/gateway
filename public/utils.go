package public

import (
	"crypto/sha256"
	"encoding/hex"
)

func GenSaltPassword(password, salt string) string {
	sha := sha256.New()
	sha.Write([]byte(password + salt))
	return hex.EncodeToString(sha.Sum(nil))
}
