package gannex

import (
	"github.com/G-Node/git-module"
)

const (
	BYTE     = 1.0
	KILOBYTE = 1024 * BYTE
	MEGABYTE = 1024 * KILOBYTE
	GIGABYTE = 1024 * MEGABYTE
	TERABYTE = 1024 * GIGABYTE
)

type ACommand struct {
	git.Command
	name string
	args []string
	env  []string
}

func AInit(dir string, args ...string) (string, error) {
	cmd := git.NewACommand("init")
	return cmd.AddArguments(args...).RunInDir(dir)
}

func AUInit(dir string, args ...string) (string, error) {
	cmd := git.NewACommand("uninit")
	return cmd.AddArguments(args...).RunInDir(dir)
}

func Worm(dir string) (string, error) {
	cmd := git.NewCommand("config", "annex.backends", "WORM")
	return cmd.RunInDir(dir)
}

func Md5(dir string) (string, error) {
	cmd := git.NewCommand("config", "annex.backends", "MD5")
	return cmd.RunInDir(dir)
}

func ASync(dir string, args ...string) (string, error) {
	cmd := git.NewACommand("sync")
	return cmd.AddArguments(args...).RunInDir(dir)
}

func Add(filename, dir string) (string, error) {
	cmd := git.NewACommand("add", filename)
	return cmd.RunInDir(dir)
}
