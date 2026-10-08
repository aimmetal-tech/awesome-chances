package security

import (
	"strings"
	"testing"
)

func TestPasswordSaltAndVerification(t *testing.T) {
	password := "correct horse battery staple"
	first, err := PasswordHash(password)
	if err != nil {
		t.Fatal(err)
	}
	second, err := PasswordHash(password)
	if err != nil {
		t.Fatal(err)
	}
	if first == second || strings.Contains(first, password) || !strings.HasPrefix(first, "$argon2id$") {
		t.Fatal("password or salt invariant")
	}
	for _, test := range []struct {
		password string
		want     bool
	}{{password, true}, {"wrong password", false}} {
		ok, err := VerifyPassword(test.password, first)
		if err != nil || ok != test.want {
			t.Fatal(ok, err)
		}
	}
	for _, bad := range []string{"plaintext", "$argon2id$v=19$m=999999999,t=99,p=255$x$y", strings.Replace(first, "v=19", "v=99", 1), strings.Replace(first, "$argon2id$", "$argon2i$", 1)} {
		if _, err := VerifyPassword(password, bad); err == nil {
			t.Fatal("malformed/unbounded hash accepted")
		}
	}
}

func TestSessionTokenEntropyAndHash(t *testing.T) {
	a, _ := RandomToken()
	b, _ := RandomToken()
	if len(a) != 43 || a == b || len(TokenHash(a)) != 64 || TokenHash(a) == a {
		t.Fatal("invalid session token")
	}
}
