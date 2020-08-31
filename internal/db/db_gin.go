package db

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/G-Node/git-module"
	"github.com/G-Node/gogs/internal/setting"
	"github.com/G-Node/libgin/libgin"
	"github.com/G-Node/libgin/libgin/annex"
	"github.com/unknwon/com"
	"golang.org/x/crypto/bcrypt"
	log "gopkg.in/clog.v1"
)

// StartIndexing sends an indexing request to the configured indexing service
// for a repository.
func StartIndexing(repo Repository) {
	go func() {
		if setting.Search.IndexURL == "" {
			log.Trace("Indexing not enabled")
			return
		}
		log.Trace("Indexing repository %d", repo.ID)
		ireq := libgin.IndexRequest{
			RepoID:   repo.ID,
			RepoPath: repo.FullName(),
		}
		data, err := json.Marshal(ireq)
		if err != nil {
			log.Error(2, "Could not marshal index request: %v", err)
			return
		}
		key := []byte(setting.Search.Key)
		encdata, err := libgin.EncryptString(key, string(data))
		if err != nil {
			log.Error(2, "Could not encrypt index request: %v", err)
		}
		req, err := http.NewRequest(http.MethodPost, setting.Search.IndexURL, strings.NewReader(encdata))
		if err != nil {
			log.Error(2, "Error creating index request")
		}
		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			log.Error(2, "Error submitting index request for [%d: %s]: %v", repo.ID, repo.FullName(), err)
			return
		}
	}()
}

// RebuildIndex sends all repositories to the indexing service to be indexed.
func RebuildIndex() error {
	indexurl := setting.Search.IndexURL
	if indexurl == "" {
		return fmt.Errorf("Indexing service not configured")
	}

	// collect all repo ID -> Path mappings directly from the DB
	repos := make(RepositoryList, 0, 100)
	if err := x.Find(&repos); err != nil {
		return fmt.Errorf("get all repos: %v", err)
	}
	log.Trace("Found %d repositories to index", len(repos))
	for _, repo := range repos {
		StartIndexing(*repo)
	}
	log.Trace("Rebuilding search index")
	return nil
}

func annexUninit(path string) {
	// walker sets the permission for any file found to 0660, to allow deletion
	var mode os.FileMode
	walker := func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return nil
		}

		mode = 0660
		if info.IsDir() {
			mode = 0770
		}

		if err := os.Chmod(path, mode); err != nil {
			log.Error(3, "failed to change permissions on '%s': %v", path, err)
		}
		return nil
	}

	log.Trace("Uninit annex at '%s'", path)
	if msg, err := annex.Uninit(path); err != nil {
		log.Error(3, "uninit failed: %v (%s)", err, msg)
		if werr := filepath.Walk(path, walker); werr != nil {
			log.Error(3, "file permission change failed: %v", werr)
		}
	}
}

func annexSetup(path string) {
	log.Trace("Running annex add (with filesize filter) in '%s'", path)

	// Initialise annex in case it's a new repository
	if msg, err := annex.Init(path); err != nil {
		log.Error(2, "Annex init failed: %v (%s)", err, msg)
		return
	}

	// Upgrade to v8 in case the directory was here before and wasn't cleaned up properly
	if msg, err := annex.Upgrade(path); err != nil {
		log.Error(2, "Annex upgrade failed: %v (%s)", err, msg)
		return
	}

	// Enable addunlocked for annex v8
	if msg, err := annex.SetAddUnlocked(path); err != nil {
		log.Error(2, "Failed to set 'addunlocked' annex option: %v (%s)", err, msg)
	}

	// Set MD5 as default backend
	if msg, err := annex.MD5(path); err != nil {
		log.Error(2, "Failed to set default backend to 'MD5': %v (%s)", err, msg)
	}

	// Set size filter in config
	if msg, err := annex.SetAnnexSizeFilter(path, setting.Repository.Upload.AnnexFileMinSize*annex.MEGABYTE); err != nil {
		log.Error(2, "Failed to set size filter for annex: %v (%s)", err, msg)
	}
}

func annexSync(path string) error {
	log.Trace("Synchronising annexed data")
	if msg, err := annex.ASync(path, "--content"); err != nil {
		// TODO: This will also DOWNLOAD content, which is unnecessary for a simple upload
		// TODO: Use gin-cli upload function instead
		log.Error(2, "Annex sync failed: %v (%s)", err, msg)
		return fmt.Errorf("git annex sync --content [%s]", path)
	}

	// run twice; required if remote annex is not initialised
	if msg, err := annex.ASync(path, "--content"); err != nil {
		log.Error(2, "Annex sync failed: %v (%s)", err, msg)
		return fmt.Errorf("git annex sync --content [%s]", path)
	}
	return nil
}

func annexAdd(repoPath string, all bool, files ...string) error {
	cmd := git.NewCommand("annex", "add")
	if all {
		cmd.AddArguments(".")
	}
	_, err := cmd.AddArguments(files...).RunInDir(repoPath)
	return err
}

func annexUpload(repoPath, remote string) error {
	log.Trace("Synchronising annex info")
	if msg, err := git.NewCommand("annex", "sync").RunInDir(repoPath); err != nil {
		log.Error(2, "git-annex sync failed: %v (%s)", err, msg)
	}
	log.Trace("Uploading annexed data")
	cmd := git.NewCommand("annex", "copy", fmt.Sprintf("--to=%s", remote), "--all")
	if msg, err := cmd.RunInDir(repoPath); err != nil {
		log.Error(2, "git-annex copy failed: %v (%s)", err, msg)
		return fmt.Errorf("git annex copy [%s]", repoPath)
	}
	return nil
}

// isAddressAllowed returns true if the email address is allowed to sign up
// based on the regular expressions found in the email filter file
// (custom/addressfilters).
// In case of errors (opening or reading file) or no matches, the function
// defaults to 'true'.
func isAddressAllowed(email string) bool {
	fpath := path.Join(setting.CustomPath, "addressfilters")
	if !com.IsExist(fpath) {
		// file doesn't exist: default allow everything
		return true
	}

	f, err := os.Open(fpath)
	if err != nil {
		log.Error(2, "Failed to open file %q: %v", fpath, err)
		// file read error: default allow everything
		return true
	}
	defer f.Close()

	emailBytes := []byte(email)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// Check provided email address against each line regex
		// Failure to match any line returns true (allowed)
		// Matching a line prefixed with + returns true (allowed)
		// Matching a line prefixed with - returns false (blocked)
		// Erroneous patterns are logged and ignored
		var allow bool
		line := scanner.Text()
		if line[0] == '-' {
			allow = false
		} else if line[0] == '+' {
			allow = true
		} else {
			log.Error(2, "Invalid line in addressfilters: %s", line)
			log.Error(2, "Prefix invalid (must be '-' or '+')")
			continue
		}
		pattern := strings.TrimSpace(line[1:])
		match, err := regexp.Match(pattern, emailBytes)
		if err != nil {
			log.Error(2, "Invalid line in addressfilters: %s", line)
			log.Error(2, "Invalid pattern: %v", err)
		}
		if match {
			return allow
		}
	}

	// No match: Default to allow
	return true
}

func IsBlockedDomain(email string) bool {
	fpath := path.Join(setting.CustomPath, "blocklist")
	if !com.IsExist(fpath) {
		return false
	}

	f, err := os.Open(fpath)
	if err != nil {
		log.Error(2, "Failed to open file %q: %v", fpath, err)
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// Check provided email address against each line as suffix
		if strings.HasSuffix(email, scanner.Text()) {
			log.Trace("New user email matched blocked domain: %q", email)
			return true
		}
	}

	return false
}

func (u *User) OldGinVerifyPassword(plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Passwd), []byte(plain))
	return err == nil
}
