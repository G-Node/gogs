package db

import (
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/G-Node/gogs/internal/conf"
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

// Writes the filters to a file in the specified custom directory. This file needs to
// be cleaned up afterwards. Returns the full path to the written file as string.
func writeCustomDirFilterFile(t *testing.T, filters []string) string {
	fname := filepath.Join(conf.CustomDir(), "addressfilters")

	if err := ioutil.WriteFile(fname, []byte(strings.Join(filters, "\n")), 0777); err != nil {
		t.Fatalf("Failed to write line filters to file %q: %v", fname, err.Error())
	}
	return fname
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
	cdir := filepath.Join(conf.CustomDir())
	if _, err := os.Stat(cdir); os.IsNotExist(err) {
		_ = os.Mkdir(cdir, 0777)
	}

	ffile := writeCustomDirFilterFile(t, allowGNodeFilter)
	defer os.Remove(ffile)

	for _, address := range emails {
		filterExpect(t, address, false)
	}

	filterExpect(t, "me@g-node.org", true)
	filterExpect(t, "malicious@g-node.org@piracy.tk", false)
}

func TestEverythingFilters(t *testing.T) {
	cdir := filepath.Join(conf.CustomDir())
	if _, err := os.Stat(cdir); os.IsNotExist(err) {
		_ = os.Mkdir(cdir, 0777)
	}

	ffile := writeCustomDirFilterFile(t, allowEverythingFilter)
	defer os.Remove(ffile)

	rand.Seed(time.Now().UnixNano())

	for idx := 0; idx < 100; idx++ {
		filterExpect(t, randAddress(), true)
	}

	ffile = writeCustomDirFilterFile(t, blockEverythingFilter)
	defer os.Remove(ffile)

	for idx := 0; idx < 100; idx++ {
		filterExpect(t, randAddress(), false)
	}
}

func TestBlockDomainFilter(t *testing.T) {
	cdir := filepath.Join(conf.CustomDir())
	if _, err := os.Stat(cdir); os.IsNotExist(err) {
		_ = os.Mkdir(cdir, 0777)
	}

	ffile := writeCustomDirFilterFile(t, blockMalicious)
	defer os.Remove(ffile)

	// 0, 3 should be allowed; 1, 2 should be blocked
	filterExpect(t, emails[0], true)
	filterExpect(t, emails[1], false)
	filterExpect(t, emails[2], false)
	filterExpect(t, emails[3], true)
}

func TestFiltersNone(t *testing.T) {
	filterExpect(t, emails[3], true)
}
