package snclient

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListenerConfig(t *testing.T) {
	t.Parallel()

	conf := map[string]string{
		"port":          "8443s",
		"bind to":       "*",
		"allowed hosts": "localhost, [::1], 127.0.0.1, 192.168.123.0/24",
	}

	listen := Listener{}
	err := listen.setListenConfig(conf)
	assert.NoErrorf(t, err, "setListenConfig should not return an error")

	for _, check := range []struct {
		expect bool
		addr   string
	}{
		{true, "127.0.0.1"},
		{false, "127.0.0.2"},
		{true, "192.168.123.1"},
		{false, "192.168.125.5"},
	} {
		assert.Equalf(t, check.expect, listen.CheckAllowedHosts(check.addr), fmt.Sprintf("CheckAllowedHosts(%s) -> %v", check.addr, check.expect))
	}
}
