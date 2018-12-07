// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/G-Node/gogs/pkg/setting"
	"github.com/G-Node/libgin/libgin"
)

type ArchiveType int

const (
	ZIP ArchiveType = iota + 1
	TARGZ
	GIN
)

func (c *Commit) CreateArchive(target string, archiveType ArchiveType, cloneL string) error {
	var format string
	switch archiveType {
	case ZIP:
		format = "zip"
	case TARGZ:
		format = "tar.gz"
	case GIN:
		to := filepath.Join(setting.Repository.Upload.TempPath, "archives", filepath.Base(strings.TrimSuffix(c.repo.Path, ".git")))
		defer os.RemoveAll(to)
		_, err := NewCommand("clone", c.repo.Path, to).RunTimeout(-1)
		if err != nil {
			return err
		}
		_, err = NewCommand("remote", "set-url", "origin", cloneL).RunInDirTimeout(-1, to)
		if err != nil {
			return err
		}
		fp, err := os.Create(target)
		defer fp.Close()
		if err != nil {
			return err
		}
		err = libgin.MkZip(to, fp)
		return err
	default:
		return fmt.Errorf("unknown format: %v", archiveType)
	}

	_, err := NewCommand("archive", "--prefix="+filepath.Base(strings.TrimSuffix(c.repo.Path, ".git"))+"/", "--format="+format, "-o", target, c.ID.String()).RunInDir(c.repo.Path)
	return err
}
