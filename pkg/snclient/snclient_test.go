package snclient

import (
	"fmt"
	"testing"

	_ "pkg/dump"

	"github.com/stretchr/testify/assert"
)

func TestPasswords(t *testing.T) {
	config := fmt.Sprintf(`
[/settings]
password0 =
password1 = %s
password2 = secret
password3 = SHA256:9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08
`, DefaultPassword)

	snc := StartTestAgent(t, config)
	conf := snc.Config.Section("/settings")

	disableLogsTemporarily()
	defer restoreLogLevel()

	p0, _ := conf.GetString("password0")
	assert.Truef(t, snc.verifyPassword(p0, "test"), "password check disabled -> ok")
	assert.Truef(t, snc.verifyPassword(p0, ""), "password check disabled -> ok")

	p1, _ := conf.GetString("password1")
	assert.Falsef(t, snc.verifyPassword(p1, "test"), "default password still set -> not ok")
	assert.Falsef(t, snc.verifyPassword(p1, DefaultPassword), "default password still set -> not ok")

	p2, _ := conf.GetString("password2")
	assert.Truef(t, snc.verifyPassword(p2, "secret"), "simple password -> ok")
	assert.Falsef(t, snc.verifyPassword(p2, "wrong"), "simple password wrong")

	p3, _ := conf.GetString("password3")
	assert.Truef(t, snc.verifyPassword(p3, "test"), "hashed password -> ok")
	assert.Falsef(t, snc.verifyPassword(p3, "wrong"), "hashed password wrong")

	StopTestAgent(t, snc)
}
