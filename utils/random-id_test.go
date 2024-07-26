package utils

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRamdomId8(t *testing.T) {

	// Success
	var (
		got  = RandomId(8)
		test = `[abcdef0-9]{8}`
		re   = regexp.MustCompile(test)
	)

	assert.Regexpf(t, re, got, "RamdomId8() = %v doesn't appply for %v", got, test)

	// Fail
	orig_randRead := randRead
	defer func() { randRead = orig_randRead }()

	randRead = func(b []byte) (n int, err error) { return 0, fmt.Errorf("test fail RamdomId8") }
	assert.Panicsf(t, func() { RandomId(8) }, "RamdomId8() expected to panic")
}
