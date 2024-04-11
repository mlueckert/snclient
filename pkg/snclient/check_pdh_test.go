//go:build !windows

package snclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckPDH(t *testing.T) {
	snc := StartTestAgent(t, "")

	res := snc.RunCheck("check_pdh", []string{})
	assert.Equalf(t, CheckExitOK, res.State, "state ok")

	StopTestAgent(t, snc)
}
