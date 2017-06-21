// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Unknwon/com"
	"github.com/urfave/cli"
	log "gopkg.in/clog.v1"

	"github.com/gogits/gogs/models"
	"github.com/gogits/gogs/models/errors"
	"github.com/gogits/gogs/pkg/setting"
	http "github.com/gogits/gogs/routers/repo"
	"syscall"
)

const (
	_ACCESS_DENIED_MESSAGE = "Repository does not exist or you do not have access"
)

var Serv = cli.Command{
	Name:        "serv",
	Usage:       "This command should only be called by SSH shell",
	Description: `Serv provide access auth for repositories`,
	Action:      runServ,
	Flags: []cli.Flag{
		stringFlag("config, c", "custom/conf/app.ini", "Custom configuration file path"),
	},
}

func fail(userMessage, logMessage string, args ...interface{}) {
	fmt.Fprintln(os.Stderr, "Gin:", userMessage)

	if len(logMessage) > 0 {
		if !setting.ProdMode {
			fmt.Fprintf(os.Stderr, logMessage+"\n", args...)
		}
		log.Fatal(3, logMessage, args...)
	}

	os.Exit(1)
}

func setup(c *cli.Context, logPath string, connectDB bool) {
	if c.IsSet("config") {
		setting.CustomConf = c.String("config")
	} else if c.GlobalIsSet("config") {
		setting.CustomConf = c.GlobalString("config")
	}

	setting.NewContext()

	level := log.TRACE
	if setting.ProdMode {
		level = log.ERROR
	}
	log.New(log.FILE, log.FileConfig{
		Level:    level,
		Filename: filepath.Join(setting.LogRootPath, logPath),
		FileRotationConfig: log.FileRotationConfig{
			Rotate:  true,
			Daily:   true,
			MaxDays: 3,
		},
	})
	log.Delete(log.CONSOLE) // Remove primary logger

	if !connectDB {
		return
	}

	models.LoadConfigs()

	if setting.UseSQLite3 {
		workDir, _ := setting.WorkDir()
		os.Chdir(workDir)
	}

	if err := models.SetEngine(); err != nil {
		fail("Internal error", "SetEngine: %v", err)
	}
}

func isAnnexShell(cmd string) bool {
	return cmd == "git-annex-shell"
}

func parseSSHCmd(cmd string) (string, string, []string) {
	ss := strings.Split(cmd, " ")
	if len(ss) < 2 {
		return "", "", nil
	}
	if isAnnexShell(ss[0]) {
		return ss[0], strings.Replace(ss[2], "/", "'", 1), ss
	} else {
		return ss[0], strings.Replace(ss[1], "/", "'", 1), ss
	}
}

func checkDeployKey(key *models.PublicKey, repo *models.Repository) {
	// Check if this deploy key belongs to current repository.
	if !models.HasDeployKey(key.ID, repo.ID) {
		fail("Key access denied", "Deploy key access denied: [key_id: %d, repo_id: %d]", key.ID, repo.ID)
	}

	// Update deploy key activity.
	deployKey, err := models.GetDeployKeyByRepo(key.ID, repo.ID)
	if err != nil {
		fail("Internal error", "GetDeployKey: %v", err)
	}

	deployKey.Updated = time.Now()
	if err = models.UpdateDeployKey(deployKey); err != nil {
		fail("Internal error", "UpdateDeployKey: %v", err)
	}
}

var (
	allowedCommands = map[string]models.AccessMode{
		"git-upload-pack":    models.ACCESS_MODE_READ,
		"git-upload-archive": models.ACCESS_MODE_READ,
		"git-receive-pack":   models.ACCESS_MODE_WRITE,
		"git-annex-shell":    models.ACCESS_MODE_READ,
	}
)

func runServ(c *cli.Context) error {
	setup(c, "serv.log", true)

	if setting.SSH.Disabled {
		println("Gins: SSH has been disabled")
		return nil
	}

	if len(c.Args()) < 1 {
		fail("Not enough arguments", "Not enough arguments")
	}

	sshCmd := strings.Replace(os.Getenv("SSH_ORIGINAL_COMMAND"), "'", "", -1)
	log.Info("SSH commadn:%s", sshCmd)
	if len(sshCmd) == 0 {
		println("Hi there, You've successfully authenticated, but Gin does not provide shell access.")
		return nil
	}

	verb, path, args := parseSSHCmd(sshCmd)
	repoFullName := strings.ToLower(strings.Trim(path, "'"))
	repoFields := strings.SplitN(repoFullName, "/", 2)
	if len(repoFields) != 2 {
		fail("Invalid repository path", "Invalid repository path: %v", path)
	}
	ownerName := strings.ToLower(repoFields[0])
	repoName := strings.TrimSuffix(strings.ToLower(repoFields[1]), ".git")
	repoName = strings.TrimSuffix(repoName, ".wiki")

	owner, err := models.GetUserByName(ownerName)
	if err != nil {
		if errors.IsUserNotExist(err) {
			fail("Repository owner does not exist", "Unregistered owner: %s", ownerName)
		}
		fail("Internal error", "Fail to get repository owner '%s': %v", ownerName, err)
	}

	repo, err := models.GetRepositoryByName(owner.ID, repoName)
	if err != nil {
		if errors.IsRepoNotExist(err) {
			fail(_ACCESS_DENIED_MESSAGE, "Repository does not exist: %s/%s", owner.Name, repoName)
		}
		fail("Internal error", "Fail to get repository: %v", err)
	}
	repo.Owner = owner

	requestMode, ok := allowedCommands[verb]
	if !ok {
		fail("Unknown git command", "Unknown git command '%s'", verb)
	}

	// Prohibit push to mirror repositories.
	if requestMode > models.ACCESS_MODE_READ && repo.IsMirror {
		fail("Mirror repository is read-only", "")
	}

	// Allow anonymous (user is nil) clone for public repositories.
	var user *models.User

	key, err := models.GetPublicKeyByID(com.StrTo(strings.TrimPrefix(c.Args()[0], "key-")).MustInt64())
	if err != nil {
		fail("Invalid key ID", "Invalid key ID '%s': %v", c.Args()[0], err)
	}

	if us, err := models.GetUserByKeyID(key.ID); err == nil {
		user = us
	} else {
		fail("Key Error", "Cannot find key %v", err)
	}

	if requestMode == models.ACCESS_MODE_WRITE || repo.IsPrivate {
		// Check deploy key or user key.
		if key.IsDeployKey() {
			if key.Mode < requestMode {
				fail("Key permission denied", "Cannot push with deployment key: %d", key.ID)
			}
			checkDeployKey(key, repo)
		} else {
			user, err = models.GetUserByKeyID(key.ID)
			if err != nil {
				fail("Internal error", "Fail to get user by key ID '%d': %v", key.ID, err)
			}

			mode, err := models.AccessLevel(user.ID, repo)
			if err != nil {
				fail("Internal error", "Fail to check access: %v", err)
			}

			if mode < requestMode {
				clientMessage := _ACCESS_DENIED_MESSAGE
				if mode >= models.ACCESS_MODE_READ {
					clientMessage = "You do not have sufficient authorization for this action"
				}
				fail(clientMessage,
					"User '%s' does not have level '%v' access to repository '%s'",
					user.Name, requestMode, repoFullName)
			}
		}
	} else {
		setting.NewService()
		// Check if the key can access to the repository in case of it is a deploy key (a deploy keys != user key).
		// A deploy key doesn't represent a signed in user, so in a site with Service.RequireSignInView activated
		// we should give read access only in repositories where this deploy key is in use. In other case, a server
		// or system using an active deploy key can get read access to all the repositories in a Gogs service.
		if key.IsDeployKey() && setting.Service.RequireSignInView {
			checkDeployKey(key, repo)
		}
	}

	// Update user key activity.
	if key.ID > 0 {
		key, err := models.GetPublicKeyByID(key.ID)
		if err != nil {
			fail("Internal error", "GetPublicKeyByID: %v", err)
		}

		key.Updated = time.Now()
		if err = models.UpdatePublicKey(key); err != nil {
			fail("Internal error", "UpdatePublicKey: %v", err)
		}
	}

	// Special handle for Windows.
	// Todo will break with annex
	if setting.IsWindows {
		verb = strings.Replace(verb, "-", " ", 1)
	}
	verbs := strings.Split(verb, " ")
	var cmd []string
	if len(verbs) == 2 {
		cmd = []string{verbs[0], verbs[1], repoFullName}
	} else if isAnnexShell(verb) {
		repoAbsPath := setting.RepoRootPath + "/" + repoFullName
		if err := secureGitAnnex(repoAbsPath, user, repo); err != nil {
			fail("Git annex failed", "Git annex failed: %s", err)
		}
		cmd = args
		// Setting full path to repo as git-annex-shell requires it
		cmd[2] = repoAbsPath
	} else {
		cmd = []string{verb, repoFullName}
	}
	runGit(cmd, requestMode, user, owner, repo)
	return nil

}

func runGit(cmd []string, requestMode models.AccessMode, user *models.User, owner *models.User,
	repo *models.Repository) error {
	log.Info("will exectute:%s", cmd)
	gitCmd := exec.Command(cmd[0], cmd[1:]...)
	if requestMode == models.ACCESS_MODE_WRITE {
		gitCmd.Env = append(os.Environ(), models.ComposeHookEnvs(models.ComposeHookEnvsOptions{
			AuthUser:  user,
			OwnerName: owner.Name,
			OwnerSalt: owner.Salt,
			RepoID:    repo.ID,
			RepoName:  repo.Name,
			RepoPath:  repo.RepoPath(),
		})...)
	}
	gitCmd.Dir = setting.RepoRootPath
	gitCmd.Stdout = os.Stdout
	gitCmd.Stdin = os.Stdin
	gitCmd.Stderr = os.Stderr
	log.Info("args:%s", gitCmd.Args)
	err := gitCmd.Run()
	log.Info("err:%s", err)
	if t, ok := err.(*exec.ExitError); ok {
		log.Info("t:%s", t)
		os.Exit(t.Sys().(syscall.WaitStatus).ExitStatus())
	}

	return nil
}

// Make sure git-annex-shell does not make "bad" changes (refectored from repo)
func secureGitAnnex(path string, user *models.User, repo *models.Repository) error {
	// "If set, disallows running git-shell to handle unknown commands."
	err := os.Setenv("GIT_ANNEX_SHELL_LIMITED", "True")
	if err != nil {
		return fmt.Errorf("ERROR: Could set annex shell to be limited.")
	}
	// "If set, git-annex-shell will refuse to run commands
	//  that do not operate on the specified directory."
	err = os.Setenv("GIT_ANNEX_SHELL_DIRECTORY", path)
	if err != nil {
		return fmt.Errorf("ERROR: Could set annex shell directory.")
	}
	mode := models.ACCESS_MODE_NONE
	if user != nil {
		mode, err = models.AccessLevel(user.ID, repo)
		if err != nil {
			fail("Internal error", "Fail to check access: %v", err)
		}
	}
	if mode < models.ACCESS_MODE_WRITE {
		err = os.Setenv("GIT_ANNEX_SHELL_READONLY", "True")
		if err != nil {
			return fmt.Errorf("ERROR: Could set annex shell to read only.")
		}
	}
	return nil
}
