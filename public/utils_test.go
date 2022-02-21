package public

import (
	"fmt"
	"testing"
)

func TestGenSaltPassword(t *testing.T) {
	salt := "admin"
	password := "123456"
	fmt.Println("1=======", GenSaltPassword(password, salt))
}
