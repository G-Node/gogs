package db

import (
	"io/ioutil"
	"math/rand"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/G-Node/gogs/internal/setting"
)

const ALNUM = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var emails = []string{
	"foo@example.org",
	"spammer@example.com",
	"user@malicious-domain.net",
	"someone@example.mal",
}

var blockEverythingFilter = []string{
	"- .*",
}

var allowEverythingFilter = []string{
	"+ .*",
}

var allowGNodeFilter = []string{
	"+ @g-node.org$",
	"- .*",
}

var blockMalicious = []string{
	"- .*malicious-domain.net$",
	"- spammer@",
}

func filterExpect(t *testing.T, address string, expect bool) {
	if isAddressAllowed(address) != expect {
		t.Fatalf("Address %q block: expected %t got %t", address, expect, !expect)
	}
}

// Writes the filters to a file and returns the directory that contains the
// filter file (to be set as setting.CustomPath)
func writeFilterFile(t *testing.T, filters []string) string {
	path := t.TempDir()
	fname := filepath.Join(path, "addressfilters")

	if err := ioutil.WriteFile(fname, []byte(strings.Join(filters, "\n")), 0777); err != nil {
		t.Fatalf("Failed to write line filters to file %q: %v", fname, err.Error())
	}
	return path
}

// randAlnum returns a random alphanumeric (lowercase, latin) string of length 'n'.
func randAlnum(n int) string {
	N := len(ALNUM)
	chrs := make([]byte, n)
	for idx := range chrs {
		chrs[idx] = ALNUM[rand.Intn(N)]
	}

	return string(chrs)
}

func randAddress() string {
	user := randAlnum(rand.Intn(20))
	domain := randAlnum(rand.Intn(20)) + "." + randAlnum(rand.Intn(3))

	return string(user) + "@" + string(domain)
}

func TestAllowGNodeFilter(t *testing.T) {
	setting.CustomPath = writeFilterFile(t, allowGNodeFilter)

	for _, address := range emails {
		filterExpect(t, address, false)
	}

	filterExpect(t, "me@g-node.org", true)
	filterExpect(t, "malicious@g-node.org@piracy.tk", false)
}

func TestEverythingFilters(t *testing.T) {
	setting.CustomPath = writeFilterFile(t, allowEverythingFilter)
	rand.Seed(time.Now().UnixNano())

	for idx := 0; idx < 100; idx++ {
		filterExpect(t, randAddress(), true)
	}

	setting.CustomPath = writeFilterFile(t, blockEverythingFilter)

	for idx := 0; idx < 100; idx++ {
		filterExpect(t, randAddress(), false)
	}
}

func TestBlockDomainFilter(t *testing.T) {
	setting.CustomPath = writeFilterFile(t, blockMalicious)

	// 0, 3 should be allowed; 1, 2 should be blocked
	filterExpect(t, emails[0], true)
	filterExpect(t, emails[1], false)
	filterExpect(t, emails[2], false)
	filterExpect(t, emails[3], true)
}

func TestFiltersNone(t *testing.T) {
	setting.CustomPath = filepath.Join(t.TempDir(), "does", "not", "exist")
	filterExpect(t, emails[3], true)
}
