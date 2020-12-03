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
		if isAddressAllowed(address) {
			t.Fatalf("Address %q should be blocked but was allowed", address)
		}
	}

	if !isAddressAllowed("me@g-node.org") {
		t.Fatalf("G-Node address blocked but should be allowed")
	}

	if isAddressAllowed("malicious@g-node.org@piracy.tk") {
		t.Fatalf("Malicious address should be blocked but was allowed")
	}
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
		randress := randAddress()
		if !isAddressAllowed(randAddress()) {
			t.Fatalf("Address %q should be allowed but was blocked", randress)
		}
	}

	ffile = writeCustomDirFilterFile(t, blockEverythingFilter)
	defer os.Remove(ffile)

	for idx := 0; idx < 100; idx++ {
		randress := randAddress()
		if isAddressAllowed(randAddress()) {
			t.Fatalf("Address %q should be blocked but was allowed", randress)
		}
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
	if address := emails[0]; !isAddressAllowed(address) {
		t.Fatalf("Address %q should be allowed but was blocked", address)
	}

	if address := emails[1]; isAddressAllowed(address) {
		t.Fatalf("Address %q should be blocked but was allowed", address)
	}

	if address := emails[2]; isAddressAllowed(address) {
		t.Fatalf("Address %q should be blocked but was allowed", address)
	}

	if address := emails[3]; !isAddressAllowed(address) {
		t.Fatalf("Address %q should be allowed but was blocked", address)
	}
}

func TestFiltersNone(t *testing.T) {
	if address := emails[3]; !isAddressAllowed(address) {
		t.Fatalf("Address %q should be allowed but was blocked", address)
	}
}
