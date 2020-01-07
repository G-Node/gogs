package annex

import "github.com/G-Node/git-module"

func isAnnexed(dir string) (bool, error) {
	return false, nil
}

func Upgrade(dir string) (string, error) {
	cmd := git.NewACommand("upgrade")
	return cmd.RunInDir(dir)
}
