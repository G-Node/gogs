package annex

import (
	"fmt"

	git "github.com/G-Node/git-module"
)

const (
	BYTE     = 1.0
	KILOBYTE = 1024 * BYTE
	MEGABYTE = 1024 * KILOBYTE
	GIGABYTE = 1024 * MEGABYTE
	TERABYTE = 1024 * GIGABYTE
)

func Init(dir string, args ...string) (string, error) {
	cmd := git.NewACommand("init")
	return cmd.AddArguments(args...).RunInDir(dir)
}

func Uninit(dir string, args ...string) (string, error) {
	cmd := git.NewACommand("uninit")
	return cmd.AddArguments(args...).RunInDir(dir)
}

func Worm(dir string) (string, error) {
	cmd := git.NewCommand("config", "annex.backends", "WORM")
	return cmd.RunInDir(dir)
}

func MD5(dir string) (string, error) {
	cmd := git.NewCommand("config", "annex.backends", "MD5")
	return cmd.RunInDir(dir)
}

func ASync(dir string, args ...string) (string, error) {
	cmd := git.NewACommand("sync")
	return cmd.AddArguments(args...).RunInDir(dir)
}

func Add(dir string, args ...string) (string, error) {
	cmd := git.NewACommand("add")
	cmd.AddArguments(args...)
	return cmd.RunInDir(dir)
}

func SetAddUnlocked(dir string) (string, error) {
	cmd := git.NewCommand("config", "annex.addunlocked", "true")
	return cmd.RunInDir(dir)
}

func SetAnnexSizeFilter(dir string, size int64) (string, error) {
	cmd := git.NewCommand("config", "annex.largefiles", fmt.Sprintf("largerthan=%d", size))
	return cmd.RunInDir(dir)
}
