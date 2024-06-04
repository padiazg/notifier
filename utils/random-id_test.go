package utils

import (
	"regexp"
	"testing"
)

func TestRamdomId8(t *testing.T) {

	var (
		got  = RamdomId8()
		test = `[abcdef0-9]{8}`
		re   = regexp.MustCompile(test)
	)

	if !re.Match([]byte(got)) {
		t.Errorf("RamdomId8() = %v doesn't appply for %v", got, test)
	}
}
